//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	ackcore "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/sqs"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
	ruleres "github.com/aws-controllers-k8s/eventbridge-controller/pkg/resource/rule"
)

const (
	testEventPattern = `{"detail-type": ["ack-e2e-testevent"]}`
)

func createRule(name, bus string, tags []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}
		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		targets := []*v1alpha1.Target{{
			ARN: aws.String(queueARN),
			ID:  aws.String(queueName),
		}}

		r.WithNamespace(namespace)
		rule := ruleFor(name, namespace, bus, testEventPattern, targets, tags...)
		err = r.Create(ctx, &rule)
		assert.NilError(t, err)

		return ctx
	}
}

func createInvalidRule(name, bus string, tags []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}
		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		r.WithNamespace(namespace)

		// rule without any pattern is invalid
		rule := v1alpha1.Rule{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1alpha1.RuleSpec{
				EventBusRef: &ackcore.AWSResourceReferenceWrapper{
					From: &ackcore.AWSResourceReference{
						Name: aws.String(bus),
					},
				},
				Name: aws.String(name),
				Tags: tags,
			},
		}
		err = r.Create(ctx, &rule)
		assert.NilError(t, err) // create succeeds because we do not have validation webhooks yet

		return ctx
	}
}

func ruleSynced(name, busName string, tags []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var rule v1alpha1.Rule
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &rule)
		assert.NilError(t, err)

		syncedCondition := conditions.New(r).ResourceMatch(&rule, func(rule k8s.Object) bool {
			for _, cond := range rule.(*v1alpha1.Rule).Status.Conditions {
				if cond.Type == ackcore.ConditionTypeResourceSynced && cond.Status == corev1.ConditionTrue {
					return true
				}
			}
			return false
		})

		err = wait.For(syncedCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		sdk := ebSDKClient(t)
		resp, err := sdk.DescribeRuleWithContext(ctx, &eventbridge.DescribeRuleInput{
			EventBusName: aws.String(busName),
			Name:         aws.String(name),
		})
		assert.NilError(t, err)
		assert.Equal(t, *resp.Name, name, "compare rule: name mismatch")

		targets, err := sdk.ListTargetsByRuleWithContext(ctx, &eventbridge.ListTargetsByRuleInput{
			EventBusName: aws.String(busName),
			Rule:         aws.String(name),
		})
		assert.NilError(t, err)
		assert.Equal(t, len(targets.Targets), len(rule.Spec.Targets), "compare rule targets: count mismatch")

		listResp, err := sdk.ListTagsForResourceWithContext(ctx, &eventbridge.ListTagsForResourceInput{
			ResourceARN: resp.Arn,
		})
		assert.NilError(t, err)

		tagMap := make(map[string]string)
		for _, tag := range listResp.Tags {
			tagMap[*tag.Key] = *tag.Value
		}

		for _, tag := range tags {
			v, ok := tagMap[*tag.Key]
			assert.Equal(t, true, ok, "compare tags: tag %q not found", *tag.Key)
			assert.Equal(t, *tag.Value, v, "compare tags: tag %q value mismatch", *tag.Key)
		}

		return ctx
	}
}

func updateRule(name, busName string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var rule v1alpha1.Rule
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &rule)
		assert.NilError(t, err)

		// replace tags with three new tags
		newTags := make([]*v1alpha1.Tag, 3)
		for i := 0; i < 3; i++ {
			newTags[i] = &v1alpha1.Tag{
				Key:   aws.String(fmt.Sprintf("newtag-%d", i)),
				Value: aws.String(fmt.Sprintf("newvalue-%d", i)),
			}
		}
		rule.Spec.Tags = newTags

		// parse account id and region to create a valid target
		arnInfo, err := arn.Parse(queueARN)
		assert.NilError(t, err, "update rule: parse arn")

		targetARN := fmt.Sprintf("arn:aws:lambda:%s:%s:function:MyFunction", arnInfo.Region, arnInfo.AccountID)
		newTargets := []*v1alpha1.Target{{
			ARN: aws.String(targetARN),
			ID:  aws.String("newtarget"),
			InputTransformer: &v1alpha1.InputTransformer{
				InputPathsMap: map[string]*string{
					"instance": aws.String("$.detail.instance"),
					"status":   aws.String("$.detail.status"),
				},
				InputTemplate: aws.String("\"<instance> is in state <status>\""), // quotes needed for valid input
			},
			RetryPolicy: &v1alpha1.RetryPolicy{
				MaximumRetryAttempts: aws.Int64(0),
			},
			SQSParameters: &v1alpha1.SQSParameters{
				MessageGroupID: aws.String("someid"),
			},
		}}

		// replace semantics to test add/remove paths
		rule.Spec.Targets = newTargets

		err = r.Update(ctx, &rule)
		assert.NilError(t, err, "update rule: update kubernetes resource targets and tags")

		sdk := ebSDKClient(t)

		// assert tag synchronization
		tagsSynced := func() (bool, error) {
			resp, err := sdk.DescribeRuleWithContext(ctx, &eventbridge.DescribeRuleInput{
				EventBusName: aws.String(busName),
				Name:         aws.String(name),
			})
			if err != nil {
				return false, fmt.Errorf("describe rule: %w", err)
			}

			listResp, err := sdk.ListTagsForResourceWithContext(ctx, &eventbridge.ListTagsForResourceInput{
				ResourceARN: resp.Arn,
			})
			if err != nil {
				return false, fmt.Errorf("list tags for rule: %w", err)
			}

			serviceTags := make(map[string]string)
			for _, tag := range listResp.Tags {
				serviceTags[*tag.Key] = *tag.Value
			}

			matched := 0
			for _, tag := range newTags {
				v, ok := serviceTags[*tag.Key]
				if !ok {
					continue
				}

				if v == *tag.Value {
					matched++
				}
			}

			return matched == len(newTags), nil
		}

		err = wait.For(tagsSynced, wait.WithTimeout(time.Second*30))
		assert.NilError(t, err, "update rule: tag synchronization with backend")

		// assert target synchronization
		targetsSynced := func() (bool, error) {
			resp, err := sdk.ListTargetsByRuleWithContext(ctx, &eventbridge.ListTargetsByRuleInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(name),
			})
			if err != nil {
				return false, fmt.Errorf("list targets for rule: %w", err)
			}

			wantTargets := ruleres.SdkTargetsFromResourceTargets(newTargets)
			if ok := cmp.DeepEqual(resp.Targets, wantTargets)().Success(); !ok {
				klog.V(1).Infof("targets differ: got=%+v want=%+v", resp.Targets, wantTargets)
				return false, nil
			}
			return true, nil
		}

		err = wait.For(targetsSynced, wait.WithTimeout(time.Second*30))
		assert.NilError(t, err, "update rule: target synchronization with backend")

		return ctx
	}
}

func ruleInTerminalState(name, busName string, _ []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var rule v1alpha1.Rule
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &rule)
		assert.NilError(t, err)

		terminalCondition := conditions.New(r).ResourceMatch(&rule, func(rule k8s.Object) bool {
			for _, cond := range rule.(*v1alpha1.Rule).Status.Conditions {
				if cond.Type == ackcore.ConditionTypeTerminal && cond.Status == corev1.ConditionTrue {
					return true
				}
			}
			return false
		})

		err = wait.For(terminalCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		// no rule should be created in backend
		sdk := ebSDKClient(t)
		resp, err := sdk.ListRulesWithContext(ctx, &eventbridge.ListRulesInput{
			EventBusName: aws.String(busName),
			NamePrefix:   aws.String(name),
		})
		assert.NilError(t, err)
		assert.Assert(t, len(resp.Rules) == 0, "list rules: length mismatch")

		return ctx
	}
}

func eventReceived(busName string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		now := time.Now().UTC()
		payload := map[string]interface{}{
			"message":       "test event from ack e2e suite for eventbridge",
			"sentTimestamp": now,
		}

		payloadbytes, err := json.Marshal(payload)
		assert.NilError(t, err)

		testEvent := eventbridge.PutEventsInput{
			Entries: []*eventbridge.PutEventsRequestEntry{{
				Detail:       aws.String(string(payloadbytes)),
				DetailType:   aws.String("ack-e2e-testevent"),
				EventBusName: aws.String(busName),
				Resources:    []*string{&namespace, &testBusName, &testRuleName},
				Source:       aws.String("kubernetes.io/ack-e2e"),
				Time:         aws.Time(time.Now().UTC()),
			}},
		}

		receiveTimeout := time.Minute // rule pattern sync is eventually consistent
		timeoutctx, cancel := context.WithTimeout(ctx, receiveTimeout)
		defer cancel()

		// event sender
		go func() {
			ebsdk := ebSDKClient(t)
			ticker := time.NewTicker(time.Second * 5)
			defer ticker.Stop()

			attempts := 0
			for {
				attempts++

				select {
				case <-ticker.C:
					klog.V(1).Infof("sending test event: attempt %d", attempts)

					resp, err := ebsdk.PutEventsWithContext(ctx, &testEvent)
					assert.NilError(t, err)
					assert.Equal(t, *resp.FailedEntryCount, int64(0), "send test event: failed entry count is not 0")
				case <-timeoutctx.Done():
					return
				}
			}
		}()

		sqssdk, err := sqsSDKClient()
		assert.NilError(t, err)

		// event receiver
		var msgs []*sqs.Message
		received := func() (done bool, err error) {
			rcvResp, err := sqssdk.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
				AttributeNames:  []*string{aws.String("All")},
				QueueUrl:        aws.String(queueURL),
				WaitTimeSeconds: aws.Int64(3),
			})
			if err != nil {
				return false, fmt.Errorf("receive sqs message: %w", err)
			}

			if len(rcvResp.Messages) > 0 {
				klog.V(1).Infof("received new messages from sqs")
				msgs = rcvResp.Messages
				return true, nil
			}
			return false, nil
		}

		klog.V(1).Infof("waiting for messages from sqs")
		err = wait.For(received, wait.WithTimeout(receiveTimeout))
		assert.NilError(t, err)
		assert.Assert(t, len(msgs) > 0, "receive sqs message: no messages received")

		msgbody := msgs[0].Body
		assert.Assert(t, msgbody != nil, "receive sqs message: body is nil")

		var ebevent events.CloudWatchEvent
		err = json.Unmarshal([]byte(*msgbody), &ebevent)
		assert.NilError(t, err, "receive sqs message: unmarshal body")

		assert.Equal(
			t,
			string(ebevent.Detail),
			string(payloadbytes),
			"receive sqs message: compare send and receive payloads",
		)

		return ctx
	}
}

func deleteRule(name, busName string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		// delete rule
		var rule v1alpha1.Rule
		err = r.Get(ctx, name, namespace, &rule)
		assert.NilError(t, err)

		err = r.Delete(ctx, &rule)
		assert.NilError(t, err)

		sdk := ebSDKClient(t)

		ruleDeleted := func() (bool, error) {
			resp, err := sdk.ListRulesWithContext(ctx, &eventbridge.ListRulesInput{
				EventBusName: aws.String(busName),
			})
			if err != nil {
				return false, fmt.Errorf("list rules: %w", err)
			}

			return len(resp.Rules) == 0, nil
		}

		waitTimeout := time.Second * 30
		err = wait.For(ruleDeleted, wait.WithTimeout(waitTimeout))
		assert.NilError(t, err, "delete rule: resources not cleaned up in service control plane")

		return ctx
	}
}
