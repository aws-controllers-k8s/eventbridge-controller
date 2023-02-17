//go:build e2e

package e2e

import (
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	testBusName      = envconf.RandomName("ack-bus-e2e", 20)
	testRuleName     = envconf.RandomName("ack-rule-e2e", 20)
	testArchiveName  = envconf.RandomName("ack-archive-e2e", 20)
	testEndpointName = envconf.RandomName("ack-endpoint-e2e", 20)
)

func TestSuite(t *testing.T) {
	// required for other features
	ctrl := features.New("EventBridge Controller").
		WithLabel("feature", "ctrl").
		Setup(createController()).
		Assess("controller running without leader election", controllerRunning()).
		Feature()

	bus := features.New("EventBridge Event Bus CRUD").
		WithLabel("feature", "bus").
		Assess("create event bus", createEventBus(testBusName, tags)).
		Assess("event bus has synced", eventBusSynced(testBusName, tags)).
		Assess("update event bus", updateEventBus(testBusName)).
		Assess("delete event bus", deleteBus(testBusName)).
		Feature()

	rule := features.New("EventBridge Rule CRUD").
		WithLabel("feature", "rule").
		Setup(setupBus(testBusName, tags)).
		Assess("create rule", createRule(testRuleName, testBusName, tags)).
		Assess("rule has synced", ruleSynced(testRuleName, testBusName, tags)).
		Assess("update rule", updateRule(testRuleName, testBusName)).
		Assess("delete rule", deleteRule(testRuleName, testBusName)).
		Teardown(deleteBus(testBusName)).
		Feature()

	invalidRule := features.New("EventBridge Rule invalid in terminal state").
		WithLabel("feature", "rule").
		Setup(setupBus(testBusName, tags)).
		Assess("create invalid rule", createInvalidRule(testRuleName, testBusName, tags)).
		Assess("rule is in terminal state", ruleInTerminalState(testRuleName, testBusName, tags)).
		Assess("delete rule", deleteRule(testRuleName, testBusName)).
		Teardown(deleteBus(testBusName)).
		Feature()

	archive := features.New("EventBridge Archive CRUD").
		WithLabel("feature", "archive").
		Setup(setupBus(testBusName, tags)).
		Assess("create archive", createArchive(testArchiveName)).
		Assess("archive has synced", archiveSynced(testArchiveName)).
		Assess("update archive", updateArchive(testArchiveName)).
		Assess("delete archive", deleteArchive(testArchiveName)).
		Teardown(deleteBus(testBusName)).
		Feature()

	endpoint := features.New("EventBridge Endpoint CRUD").
		WithLabel("feature", "endpoint").
		Setup(setupBuses(testBusName, envCfg.Region, envCfg.SecondaryRegion, tags)).
		Assess("create Endpoint", createEndpoint(testEndpointName, testBusName, envCfg.SecondaryRegion)).
		Assess("Endpoint has synced", endpointSynced(testEndpointName)).
		Assess("update Endpoint", updateEndpoint(testEndpointName)).
		Assess("delete Endpoint", deleteEndpoint(testEndpointName)).
		Teardown(deleteBuses([]string{
			testBusName + "-primary",
			testBusName + "-secondary",
		})).
		Feature()

	e2e := features.New("EventBridge E2E").
		WithLabel("feature", "e2e").
		Setup(setupBus(testBusName, tags)).
		Assess("create rule", createRule(testRuleName, testBusName, tags)).
		Assess("rule has synced", ruleSynced(testRuleName, testBusName, tags)).
		Assess("event received in sqs", eventReceived(testBusName)).
		Assess("delete rule", deleteRule(testRuleName, testBusName)).
		Teardown(deleteBus(testBusName)).
		Feature()

	testEnv.Test(t, ctrl, bus, rule, invalidRule, archive, endpoint, e2e)
}
