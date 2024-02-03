//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"

	ebv1alpha1 "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

type envConfig struct {
	// aws credentials
	Region       string `envconfig:"AWS_DEFAULT_REGION" required:"true"`
	AccessKey    string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	SecretKey    string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	SessionToken string `envconfig:"AWS_SESSION_TOKEN" required:"true"`

	// global endpoints overwrite
	SecondaryRegion string `envconfig:"AWS_SECONDARY_REGION" required:"true" default:"eu-west-1"`

	// kind configuration
	KindCluster string `envconfig:"KIND_CLUSTER_NAME" required:"true" default:"ack"`
	CtrlImage   string `envconfig:"ACK_CONTROLLER_IMAGE" required:"true"`
}

type (
	namespaceCtxKey string
	testbusCtxKey   string
)

const (
	baseCRDPath   = "../../config/crd/bases"
	commonCRDPath = "../../config/crd/common/bases"

	namespaceKey = namespaceCtxKey("featureNamespace")
	busArnCtxKey = testbusCtxKey("testBusArn")
)

var (
	testEnv env.Environment
	envCfg  envConfig

	// test queue & route 53 healthcheck
	queueName       string
	queueARN        string
	queueURL        string
	healthCheckName string
	healthCheckID   string

	// common tags
	tags []*ebv1alpha1.Tag
)

func TestMain(m *testing.M) {
	envconfig.MustProcess("", &envCfg)

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		klog.Fatalf("environment variable parsing failed: %s", err)
	}

	if envCfg.Region == envCfg.SecondaryRegion {
		klog.Fatalf("%q and %q must be different", "AWS_DEFAULT_REGION", "AWS_SECONDARY_REGION")
	}

	testEnv = env.NewWithConfig(cfg)
	queueName = envconf.RandomName("ack-e2e-queue", 20)
	healthCheckName = envconf.RandomName("ack-e2e-hc", 20)

	tags = []*ebv1alpha1.Tag{{
		Key:   aws.String("ack-e2e"),
		Value: aws.String("true"),
	}}

	klog.V(1).Infof("setting up test environment with kind cluster %q", envCfg.KindCluster)
	testEnv.Setup(
		createHealthCheck(healthCheckName, tags),
		createSQSTestQueue(queueName, tags),
		envfuncs.CreateKindCluster(envCfg.KindCluster),
		envfuncs.SetupCRDs(baseCRDPath, "*"),
		envfuncs.SetupCRDs(commonCRDPath, "*"),
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(ackNamespace),
		deleteHealthCheck(),
		destroySQSTestQueue(),
		envfuncs.TeardownCRDs(baseCRDPath, "*"),
		envfuncs.TeardownCRDs(commonCRDPath, "*"),
	)

	// create/delete namespace per feature
	testEnv.BeforeEachFeature(func(ctx context.Context, cfg *envconf.Config, _ *testing.T, f features.Feature) (context.Context, error) {
		return createNSForFeature(ctx, cfg, f.Name())
	})
	testEnv.AfterEachFeature(func(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature) (context.Context, error) {
		return deleteNSForFeature(ctx, cfg, t, f.Name())
	})

	os.Exit(testEnv.Run(m))
}
