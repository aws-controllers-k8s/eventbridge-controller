//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const (
	// naive approach excludes all yaml files starting with "k"
	// TODO(embano1): hack to exclude kustomization files
	filterPattern      = "[^k]*.yaml"
	controllerFilePath = "../../config/controller"
	rbacFilePath       = "../../config/rbac"
	controllerName     = "ack-eventbridge-controller"
	ackNamespace       = "ack-system"

	// go 1.20 binary coverage
	coverDirPath = "/coverdata"
)

func createController() features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		klog.V(1).Info("creating eventbridge controller")
		r, err := resources.New(c.Client().RESTConfig())
		assert.NilError(t, err)

		envs := []v1.EnvVar{
			{
				Name:  "AWS_REGION",
				Value: envCfg.Region,
			},
			{
				Name:  "AWS_ACCESS_KEY_ID",
				Value: envCfg.AccessKey,
			},
			{
				Name:  "AWS_SECRET_ACCESS_KEY",
				Value: envCfg.SecretKey,
			},
			{
				Name:  "AWS_SESSION_TOKEN",
				Value: envCfg.SessionToken,
			},
			{
				Name:  "GOCOVERDIR",
				Value: coverDirPath,
			},
		}

		// controller manifests
		err = decoder.DecodeEachFile(
			ctx, os.DirFS(controllerFilePath), filterPattern,
			decoder.CreateHandler(r),
			mutateController(envCfg.CtrlImage, envs), // update manifest values for test
		)
		assert.NilError(t, err)

		// rbac manifests
		err = decoder.DecodeEachFile(
			ctx, os.DirFS(rbacFilePath), filterPattern,
			decoder.CreateHandler(r),
		)
		assert.NilError(t, err)

		return ctx
	}
}

func controllerRunning() features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		client, err := c.NewClient()
		assert.NilError(t, err)

		var d appsv1.Deployment
		err = client.Resources().Get(ctx, controllerName, ackNamespace, &d)
		assert.NilError(t, err)

		readyCondition := conditions.New(client.Resources()).DeploymentConditionMatch(&d, appsv1.DeploymentAvailable, v1.ConditionTrue)
		err = wait.For(readyCondition, wait.WithTimeout(time.Minute))
		assert.NilError(t, err)

		return ctx
	}
}
