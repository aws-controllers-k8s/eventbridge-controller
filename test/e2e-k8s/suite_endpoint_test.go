//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	ackcore "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

func createEndpoint(name, busNamePrefix, secondaryRegion string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}
		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		r.WithNamespace(namespace)

		// retrieve ARN from primary
		var bus v1alpha1.EventBus
		err = r.Get(ctx, busNamePrefix+"-primary", namespace, &bus)
		assert.NilError(t, err)
		primaryArn := string(*bus.Status.ACKResourceMetadata.ARN)

		// retrieve ARN from secondary
		err = r.Get(ctx, busNamePrefix+"-secondary", namespace, &bus)
		assert.NilError(t, err)
		secondaryARN := string(*bus.Status.ACKResourceMetadata.ARN)

		endpoint := endpointFor(name, namespace, primaryArn, secondaryARN, secondaryRegion, healthCheckID)
		err = r.Create(ctx, &endpoint)
		assert.NilError(t, err)

		return ctx
	}
}

func endpointSynced(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var endpoint v1alpha1.Endpoint
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &endpoint)
		assert.NilError(t, err)

		activeCondition := conditions.New(r).ResourceMatch(&endpoint, func(ep k8s.Object) bool {
			for _, cond := range ep.(*v1alpha1.Endpoint).Status.Conditions {
				if cond.Type == ackcore.ConditionTypeResourceSynced && cond.Status == corev1.ConditionTrue {
					break
				}
				return false
			}

			state := ep.(*v1alpha1.Endpoint).Status.State
			if state == nil || *state != "ACTIVE" {
				return false
			}

			return true
		})

		err = wait.For(activeCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		sdk := ebSDKClient(t)
		resp, err := sdk.DescribeEndpointWithContext(ctx, &eventbridge.DescribeEndpointInput{
			HomeRegion: aws.String(envCfg.Region),
			Name:       aws.String(name),
		})
		assert.NilError(t, err)
		assert.Equal(t, *resp.Name, name, "compare endpoint: name mismatch")

		return ctx
	}
}

func updateEndpoint(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var endpoint v1alpha1.Endpoint
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &endpoint)
		assert.NilError(t, err)

		want := aws.String("test description")
		endpoint.Spec.Description = want
		err = r.Update(ctx, &endpoint)
		assert.NilError(t, err, "update endpoint: update kubernetes resource")

		sdk := ebSDKClient(t)

		endpointUpdated := func() (bool, error) {
			got, err := sdk.DescribeEndpointWithContext(ctx, &eventbridge.DescribeEndpointInput{
				HomeRegion: aws.String(envCfg.Region),
				Name:       aws.String(name),
			})
			if err != nil {
				return false, fmt.Errorf("describe endpoint: %w", err)
			}

			if *got.Description != *want {
				return false, nil
			}

			return true, nil
		}

		err = wait.For(endpointUpdated, wait.WithTimeout(time.Second*30))
		assert.NilError(t, err, "update endpoint: synchronization with backend")

		return ctx
	}
}

func deleteEndpoint(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		var endpoint v1alpha1.Endpoint
		err = r.Get(ctx, name, namespace, &endpoint)
		assert.NilError(t, err)

		err = r.Delete(ctx, &endpoint)
		assert.NilError(t, err)

		sdk := ebSDKClient(t)

		endpointDeleted := func() (bool, error) {
			resp, err := sdk.ListEndpointsWithContext(ctx, &eventbridge.ListEndpointsInput{
				NamePrefix: aws.String(name),
			})
			if err != nil {
				return false, fmt.Errorf("list archives: %w", err)
			}

			return len(resp.Endpoints) == 0, nil
		}

		waitTimeout := time.Second * 30
		err = wait.For(endpointDeleted, wait.WithTimeout(waitTimeout))
		assert.NilError(t, err, "delete endpoint: resources not cleaned up in service control plane")

		return ctx
	}
}
