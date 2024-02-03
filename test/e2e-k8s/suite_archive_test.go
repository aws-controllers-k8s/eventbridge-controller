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

func createArchive(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		if err != nil {
			t.Fail()
		}
		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		r.WithNamespace(namespace)
		archive := archiveFor(name, namespace, getTestBusArnFromContext(ctx, t), testEventPattern)
		err = r.Create(ctx, &archive)
		assert.NilError(t, err)

		return ctx
	}
}

func archiveSynced(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var archive v1alpha1.Archive
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &archive)
		assert.NilError(t, err)

		syncedCondition := conditions.New(r).ResourceMatch(&archive, func(archive k8s.Object) bool {
			for _, cond := range archive.(*v1alpha1.Archive).Status.Conditions {
				if cond.Type == ackcore.ConditionTypeResourceSynced && cond.Status == corev1.ConditionTrue {
					return true
				}
			}
			return false
		})

		err = wait.For(syncedCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		sdk := ebSDKClient(t)
		resp, err := sdk.DescribeArchiveWithContext(ctx, &eventbridge.DescribeArchiveInput{
			ArchiveName: aws.String(name),
		})
		assert.NilError(t, err)
		assert.Equal(t, *resp.ArchiveName, name, "compare archive: name mismatch")
		assert.DeepEqual(t, resp.RetentionDays, archive.Spec.RetentionDays)
		assert.DeepEqual(t, resp.Description, archive.Spec.Description)
		assert.DeepEqual(t, resp.EventPattern, archive.Spec.EventPattern)

		return ctx
	}
}

func updateArchive(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		err = v1alpha1.AddToScheme(r.GetScheme())
		assert.NilError(t, err)

		var archive v1alpha1.Archive
		r.WithNamespace(namespace)
		err = r.Get(ctx, name, namespace, &archive)
		assert.NilError(t, err)

		want := archive.Spec
		want.Description = aws.String("test description")
		want.RetentionDays = aws.Int64(5)
		want.EventPattern = aws.String(`{"source": ["some.source"]}`)

		archive.Spec = want
		err = r.Update(ctx, &archive)
		assert.NilError(t, err, "update archive: update kubernetes resource")

		sdk := ebSDKClient(t)

		archiveUpdated := func(ctx context.Context) (bool, error) {
			got, err := sdk.DescribeArchiveWithContext(ctx, &eventbridge.DescribeArchiveInput{
				ArchiveName: aws.String(name),
			})
			if err != nil {
				return false, fmt.Errorf("describe archive: %w", err)
			}

			if *got.Description != *want.Description {
				return false, nil
			}

			if *got.RetentionDays != *want.RetentionDays {
				return false, nil
			}

			if *got.EventPattern != *want.EventPattern {
				return false, nil
			}

			return true, nil
		}

		err = wait.For(archiveUpdated, wait.WithTimeout(time.Second*30))
		assert.NilError(t, err, "update archive: synchronization with backend")

		return ctx
	}
}

func deleteArchive(name string) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		namespace := getTestNamespaceFromContext(ctx, t)

		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		var archive v1alpha1.Archive
		err = r.Get(ctx, name, namespace, &archive)
		assert.NilError(t, err)

		err = r.Delete(ctx, &archive)
		assert.NilError(t, err)

		sdk := ebSDKClient(t)

		archiveDeleted := func(ctx context.Context) (bool, error) {
			resp, err := sdk.ListArchivesWithContext(ctx, &eventbridge.ListArchivesInput{
				NamePrefix: aws.String(name),
			})
			if err != nil {
				return false, fmt.Errorf("list archives: %w", err)
			}

			return len(resp.Archives) == 0, nil
		}

		waitTimeout := time.Second * 30
		err = wait.For(archiveDeleted, wait.WithTimeout(waitTimeout))
		assert.NilError(t, err, "delete archive: resources not cleaned up in service control plane")

		return ctx
	}
}
