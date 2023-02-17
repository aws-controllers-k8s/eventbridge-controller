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
	"k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// wrapper around event bus create and has synced
func setupBus(name string, tags []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		steps := []features.Func{
			createEventBus(name, tags),
			eventBusSynced(name, tags),
		}

		for _, step := range steps {
			ctx = step(ctx, t, c)
		}

		return ctx
	}
}

// creates two identical event buses, with "-primary" and "-secondary" suffix
// respectively, in different regions to be used with global endpoints
func setupBuses(name string, primaryRegion string, secondaryRegion string, tags []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}
		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		r.WithNamespace(namespace)

		// create primary bus
		primary := eventBusFor(name, namespace, tags...)
		primary.Name = primary.Name + "-primary" // K8s object name must differ in same namespace, spec name is same
		annotations := primary.ObjectMeta.Annotations
		if annotations == nil {
			annotations = map[string]string{
				"services.k8s.aws/region": primaryRegion,
			}
		} else {
			annotations["services.k8s.aws/region"] = primaryRegion
		}
		primary.Annotations = annotations

		err = r.Create(ctx, &primary)
		assert.NilError(t, err)

		syncedCondition := conditions.New(r).ResourceMatch(&primary, func(bus k8s.Object) bool {
			for _, cond := range bus.(*v1alpha1.EventBus).Status.Conditions {
				if cond.Type == ackcore.ConditionTypeResourceSynced && cond.Status == v1.ConditionTrue {
					return true
				}
			}
			return false
		})

		err = wait.For(syncedCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		// create secondy bus
		secondary := eventBusFor(name, namespace, tags...)
		secondary.Name = secondary.Name + "-secondary"
		annotations = secondary.ObjectMeta.Annotations
		if annotations == nil {
			annotations = map[string]string{
				"services.k8s.aws/region": secondaryRegion,
			}
		} else {
			annotations["services.k8s.aws/region"] = secondaryRegion
		}
		secondary.Annotations = annotations

		err = r.Create(ctx, &secondary)
		assert.NilError(t, err)

		syncedCondition = conditions.New(r).ResourceMatch(&secondary, func(bus k8s.Object) bool {
			for _, cond := range bus.(*v1alpha1.EventBus).Status.Conditions {
				if cond.Type == ackcore.ConditionTypeResourceSynced && cond.Status == v1.ConditionTrue {
					return true
				}
			}
			return false
		})

		err = wait.For(syncedCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		return ctx
	}
}

func createEventBus(name string, tags []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}
		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		r.WithNamespace(namespace)
		bus := eventBusFor(name, namespace, tags...)
		err = r.Create(ctx, &bus)
		assert.NilError(t, err)

		return ctx
	}
}

func eventBusSynced(name string, tags []*v1alpha1.Tag) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var bus v1alpha1.EventBus
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &bus)
		assert.NilError(t, err)

		syncedCondition := conditions.New(r).ResourceMatch(&bus, func(bus k8s.Object) bool {
			for _, cond := range bus.(*v1alpha1.EventBus).Status.Conditions {
				if cond.Type == ackcore.ConditionTypeResourceSynced && cond.Status == v1.ConditionTrue {
					return true
				}
			}
			return false
		})

		err = wait.For(syncedCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		sdk := ebSDKClient(t)
		resp, err := sdk.DescribeEventBus(&eventbridge.DescribeEventBusInput{
			Name: aws.String(name),
		})
		assert.NilError(t, err)
		assert.Equal(t, *resp.Name, name, "compare bus: name mismatch")

		busArn := resp.Arn
		ctx = context.WithValue(ctx, busArnCtxKey, *busArn)

		listResp, err := sdk.ListTagsForResourceWithContext(ctx, &eventbridge.ListTagsForResourceInput{
			ResourceARN: busArn,
		})
		assert.NilError(t, err)

		serviceTags := make(map[string]string)
		for _, tag := range listResp.Tags {
			serviceTags[*tag.Key] = *tag.Value
		}

		for _, tag := range tags {
			v, ok := serviceTags[*tag.Key]
			assert.Equal(t, true, ok, "compare tags: tag not found")
			assert.Equal(t, *tag.Value, v, "compare tags: tag value mismatch")
		}

		return ctx
	}
}

// replaces existing tags with an array of new tags
func updateEventBus(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var bus v1alpha1.EventBus
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &bus)
		assert.NilError(t, err)

		// replace tags with three new tags
		newTags := make([]*v1alpha1.Tag, 3)
		for i := 0; i < 3; i++ {
			newTags[i] = &v1alpha1.Tag{
				Key:   aws.String(fmt.Sprintf("newtag-%d", i)),
				Value: aws.String(fmt.Sprintf("newvalue-%d", i)),
			}
		}
		bus.Spec.Tags = newTags

		err = r.Update(ctx, &bus)
		assert.NilError(t, err, "update event bus: update kubernetes resource tags")

		sdk := ebSDKClient(t)
		tagsSynced := func() (bool, error) {
			resp, err := sdk.DescribeEventBus(&eventbridge.DescribeEventBusInput{
				Name: aws.String(name),
			})
			if err != nil {
				return false, fmt.Errorf("describe event bus: %w", err)
			}

			listResp, err := sdk.ListTagsForResourceWithContext(ctx, &eventbridge.ListTagsForResourceInput{
				ResourceARN: resp.Arn,
			})
			if err != nil {
				return false, fmt.Errorf("list tags for event bus: %w", err)
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
		assert.NilError(t, err, "update event bus: tag synchronization with backend")
		return ctx
	}
}

func deleteBus(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		var bus v1alpha1.EventBus
		err = r.Get(ctx, name, namespace, &bus)
		assert.NilError(t, err)

		err = r.Delete(ctx, &bus)
		assert.NilError(t, err)

		sdk := ebSDKClient(t)

		busDeleted := func() (bool, error) {
			resp, err := sdk.ListEventBusesWithContext(ctx, &eventbridge.ListEventBusesInput{
				NamePrefix: aws.String(name), // ignore "default" bus
			})
			if err != nil {
				return false, fmt.Errorf("list event buses: %w", err)
			}

			return len(resp.EventBuses) == 0, nil
		}

		waitTimeout := time.Second * 30
		err = wait.For(busDeleted, wait.WithTimeout(waitTimeout))
		assert.NilError(t, err, "delete event bus: resources not cleaned up in service control plane")

		return ctx
	}
}

func deleteBuses(names []string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		var steps []features.Func
		for _, bus := range names {
			steps = append(steps, deleteBus(bus))
		}

		for _, step := range steps {
			ctx = step(ctx, t, c)
		}

		return ctx
	}
}
