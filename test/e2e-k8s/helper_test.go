//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	svcsdk "github.com/aws/aws-sdk-go/service/eventbridge"
	r53svcsdk "github.com/aws/aws-sdk-go/service/route53"
	sqssvcsdk "github.com/aws/aws-sdk-go/service/sqs"
	"gotest.tools/v3/assert"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"

	eventbridge "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

func mutateController(image string, env []corev1.EnvVar) decoder.DecodeOption {
	return decoder.MutateOption(func(obj k8s.Object) error {
		d, ok := obj.(*v1.Deployment)
		if !ok {
			// ignore non-deployment objects in input, e.g. ack-system namespace
			return nil
		}

		podSpec := &d.Spec.Template.Spec
		container := &d.Spec.Template.Spec.Containers[0]

		// only patch the ack controller in case of multiple deployments
		if d.Name != controllerName {
			return nil
		}

		/*// enables leader election
		// TODO: https://github.com/aws-controllers-k8s/community/issues/1578
		d.Spec.Replicas = aws.Int32(2)*/

		container.Image = image
		container.Command = []string{"/ko-app/controller"}
		container.Args = []string{
			"--enable-development-logging",
			"--log-level=debug",
			// "--enable-leader-election",
		}

		envVars := container.Env
		envVars = append(envVars, env...)
		container.Env = envVars

		// go coverage data handling
		const coverVolName = "coverdir"

		coverMount := corev1.VolumeMount{
			Name:      coverVolName,
			ReadOnly:  false,
			MountPath: coverDirPath,
		}
		mounts := container.VolumeMounts
		mounts = append(mounts, coverMount)
		container.VolumeMounts = mounts

		coverVol := corev1.Volume{
			Name: coverVolName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: coverDirPath,
				},
			},
		}

		volumes := podSpec.Volumes
		volumes = append(volumes, coverVol)
		podSpec.Volumes = volumes
		return nil
	})
}

func eventBusFor(name, namespace string, tags ...*eventbridge.Tag) eventbridge.EventBus {
	return eventbridge.EventBus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventbridge.EventBusSpec{
			Name: aws.String(name),
			Tags: tags,
		},
	}
}

func ruleFor(name, namespace, bus, pattern string, targets []*eventbridge.Target, tags ...*eventbridge.Tag) eventbridge.Rule {
	return eventbridge.Rule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventbridge.RuleSpec{
			EventBusRef: &v1alpha1.AWSResourceReferenceWrapper{
				From: &v1alpha1.AWSResourceReference{
					Name: aws.String(bus),
				},
			},
			EventPattern: aws.String(pattern),
			Name:         aws.String(name),
			Tags:         tags,
			Targets:      targets,
		},
	}
}

func archiveFor(name, namespace, busArn, pattern string) eventbridge.Archive {
	return eventbridge.Archive{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventbridge.ArchiveSpec{
			Name:           aws.String(name),
			EventPattern:   aws.String(pattern),
			EventSourceARN: aws.String(busArn),
			RetentionDays:  aws.Int64(0), // forever,
		},
	}
}

func endpointFor(name, namespace, primaryBusARN, secondaryBusARN, secondaryRegion, healthcheckID string) eventbridge.Endpoint {
	return eventbridge.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: eventbridge.EndpointSpec{
			EventBuses: []*eventbridge.EndpointEventBus{
				{EventBusARN: aws.String(primaryBusARN)},
				{EventBusARN: aws.String(secondaryBusARN)},
			},
			Name: aws.String(name),
			ReplicationConfig: &eventbridge.ReplicationConfig{
				State: aws.String("DISABLED"),
			},
			RoutingConfig: &eventbridge.RoutingConfig{
				FailoverConfig: &eventbridge.FailoverConfig{
					Primary: &eventbridge.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/" + healthcheckID),
					},
					Secondary: &eventbridge.Secondary{
						Route: aws.String(secondaryRegion),
					},
				},
			},
		},
	}
}

func ebSDKClient(t *testing.T) *svcsdk.EventBridge {
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(envCfg.Region),
	})
	assert.NilError(t, err, "create eventbridge service client")

	return svcsdk.New(s)
}

func sqsSDKClient() (*sqssvcsdk.SQS, error) {
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(envCfg.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("create aws session: %w", err)
	}

	return sqssvcsdk.New(s), nil
}

func route53SDKClient() (*r53svcsdk.Route53, error) {
	s, err := session.NewSession(&aws.Config{
		Region: aws.String(envCfg.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("create aws session: %w", err)
	}

	return r53svcsdk.New(s), nil
}

func createHealthCheck(name string, _ []*eventbridge.Tag) env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		routesdk, err := route53SDKClient()
		if err != nil {
			return ctx, fmt.Errorf("create route53 sdk client: %w", err)
		}

		cfg := r53svcsdk.HealthCheckConfig{
			Disabled:                 aws.Bool(false),
			EnableSNI:                aws.Bool(true),
			FailureThreshold:         aws.Int64(3),
			FullyQualifiedDomainName: aws.String("events.eu-west-1.amazonaws.com"),
			Inverted:                 aws.Bool(false),
			MeasureLatency:           aws.Bool(false),
			Port:                     aws.Int64(443),
			RequestInterval:          aws.Int64(10),
			Type:                     aws.String("HTTPS"),
		}
		req := r53svcsdk.CreateHealthCheckInput{
			CallerReference:   aws.String(name),
			HealthCheckConfig: &cfg,
		}
		resp, err := routesdk.CreateHealthCheckWithContext(ctx, &req)
		if err != nil {
			return nil, fmt.Errorf("create route53 health check: %w", err)
		}

		healthCheckID = *resp.HealthCheck.Id
		klog.V(1).Infof("created test route53 health check %q", healthCheckID)

		return ctx, nil
	}
}

func deleteHealthCheck() env.Func {
	return func(ctx context.Context, config *envconf.Config) (context.Context, error) {
		routesdk, err := route53SDKClient()
		if err != nil {
			return ctx, fmt.Errorf("create route53 sdk client: %w", err)
		}

		req := r53svcsdk.DeleteHealthCheckInput{HealthCheckId: aws.String(healthCheckID)}
		_, err = routesdk.DeleteHealthCheckWithContext(ctx, &req)
		if err != nil {
			return nil, fmt.Errorf("delete route53 health check: %w", err)
		}
		klog.V(1).Infof("deleted test route53 health check %q", healthCheckID)

		return ctx, nil
	}
}

func createSQSTestQueue(name string, tags []*eventbridge.Tag) env.Func {
	return func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
		sqssdk, err := sqsSDKClient()
		if err != nil {
			return ctx, fmt.Errorf("create sqs sdk client: %w", err)
		}

		sqstags := make(map[string]*string)
		for _, t := range tags {
			cp := *t
			sqstags[*cp.Key] = cp.Value
		}

		const sqsPolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Resource": "arn:aws:sqs:*",
      "Action": [
        "sqs:GetQueueAttributes",
        "sqs:GetQueueUrl",
        "sqs:SendMessage"
      ]
    }
  ]
}`

		resp, err := sqssdk.CreateQueue(&sqssvcsdk.CreateQueueInput{
			QueueName: aws.String(name),
			Tags:      sqstags,
			Attributes: map[string]*string{
				"Policy": aws.String(sqsPolicy),
			},
		})
		if err != nil {
			return ctx, fmt.Errorf("create sqs queue: %w", err)
		}
		queueURL = *resp.QueueUrl
		klog.V(1).Infof("created test sqs queue %q", *resp.QueueUrl)

		const arnKey = "QueueArn"
		attrs, err := sqssdk.GetQueueAttributes(&sqssvcsdk.GetQueueAttributesInput{
			AttributeNames: []*string{aws.String(arnKey)},
			QueueUrl:       resp.QueueUrl,
		})
		if err != nil {
			return ctx, fmt.Errorf("get sqs queue attributes: %w", err)
		}

		if arn, ok := attrs.Attributes[arnKey]; !ok {
			return ctx, fmt.Errorf("get sqs queue attributes: value for %q not found", arnKey)
		} else {
			queueARN = *arn
		}

		return ctx, nil
	}
}

func destroySQSTestQueue() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		sqssdk, err := sqsSDKClient()
		if err != nil {
			return ctx, fmt.Errorf("create sqs sdk client: %w", err)
		}

		_, err = sqssdk.DeleteQueue(&sqssvcsdk.DeleteQueueInput{
			QueueUrl: aws.String(queueURL),
		})
		if err != nil {
			return ctx, fmt.Errorf("destroy sqs queue: %w", err)
		}

		klog.V(1).Infof("destroyed test sqs queue %q", queueURL)
		return ctx, nil
	}
}

// createNSForFeature creates a random namespace with the runID as a prefix. It is stored in the context
// so that the deleteNSForFeature routine can look it up and delete it.
func createNSForFeature(ctx context.Context, cfg *envconf.Config, feature string) (context.Context, error) {
	ns := envconf.RandomName("ack-feature", 15)
	ctx = context.WithValue(ctx, namespaceKey, ns)

	klog.V(1).Infof("creating namespace %q for feature %q", ns, feature)
	nsObj := corev1.Namespace{}
	nsObj.Name = ns

	return ctx, cfg.Client().Resources().Create(ctx, &nsObj)
}

// deleteNSForFeature looks up the namespace corresponding to the given test and deletes it.
func deleteNSForFeature(ctx context.Context, cfg *envconf.Config, t *testing.T, feature string) (context.Context, error) {
	ns := getTestNamespaceFromContext(ctx, t)

	klog.V(1).Infof("deleting namespace %q for feature %q", ns, feature)

	nsObj := corev1.Namespace{}
	nsObj.Name = ns

	return ctx, cfg.Client().Resources().Delete(ctx, &nsObj)
}

func getTestNamespaceFromContext(ctx context.Context, t *testing.T) string {
	ns, ok := ctx.Value(namespaceKey).(string)
	assert.Equal(t, ok, true, "retrieve namespace from context: value not found for key %q", namespaceKey)
	return ns
}

func getTestBusArnFromContext(ctx context.Context, t *testing.T) string {
	arn, ok := ctx.Value(busArnCtxKey).(string)
	assert.Equal(t, ok, true, "retrieve test event bus arn from context: value not found for key %q", busArnCtxKey)
	return arn
}
