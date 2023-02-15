package endpoint

import (
	"context"
	"fmt"
	"strings"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eventbridge"
)

// customUpdateEndpoint deals with the eventually consistent API of Endpoints in
// a generic manner. It will always requeue a resource after modification to
// sync its spec and status.
//
// note: there are several edge cases in this API which we don't cover e.g.,
// going from a DISABLED replication state to ENABLED with an invalid ARN. The
// controller currently has no logic to reconcile this against the eventual
// consistent nature and inconsistent behavior of the API (API replication config will remain
// DISABLED and state UPDATE_FAILED and does not reflect the desired state)
func (rm *resourceManager) customUpdateEndpoint(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkUpdate")
	defer func() {
		exit(err)
	}()
	if immutableFieldChanges := rm.getImmutableFieldChanges(delta); len(immutableFieldChanges) > 0 {
		msg := fmt.Sprintf("Immutable Spec fields have been modified: %s", strings.Join(immutableFieldChanges, ","))
		return nil, ackerr.NewTerminalError(fmt.Errorf(msg))
	}

	if err = validateEndpointSpec(delta, desired.ko.Spec); err != nil {
		return nil, ackerr.NewTerminalError(err)
	}

	if endpointInMutatingState(latest) {
		return latest, requeueWaitWhileUpdating
	}

	input, err := rm.newUpdateRequestPayload(ctx, desired)
	if err != nil {
		return nil, err
	}

	// we need to explicitly unset nil spec values
	unsetRemovedSpecFields(delta, desired.ko.Spec, input)

	_, err = rm.sdkapi.UpdateEndpointWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateEndpoint", err)
	if err != nil {
		return nil, err
	}
	return desired, ackrequeue.NeededAfter(nil, defaultRequeueDelay)
}

// newUpdateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Update API call for the resource
func (rm *resourceManager) newUpdateRequestPayload(
	_ context.Context,
	r *resource,
) (*svcsdk.UpdateEndpointInput, error) {
	res := &svcsdk.UpdateEndpointInput{}

	if r.ko.Spec.Description != nil {
		res.SetDescription(*r.ko.Spec.Description)
	}
	if r.ko.Spec.EventBuses != nil {
		var f1 []*svcsdk.EndpointEventBus
		for _, f1iter := range r.ko.Spec.EventBuses {
			f1elem := &svcsdk.EndpointEventBus{}
			if f1iter.EventBusARN != nil {
				f1elem.SetEventBusArn(*f1iter.EventBusARN)
			}
			f1 = append(f1, f1elem)
		}
		res.SetEventBuses(f1)
	}
	if r.ko.Spec.Name != nil {
		res.SetName(*r.ko.Spec.Name)
	}
	if r.ko.Spec.ReplicationConfig != nil {
		f3 := &svcsdk.ReplicationConfig{}
		if r.ko.Spec.ReplicationConfig.State != nil {
			f3.SetState(*r.ko.Spec.ReplicationConfig.State)
		}
		res.SetReplicationConfig(f3)
	}
	if r.ko.Spec.RoleARN != nil {
		res.SetRoleArn(*r.ko.Spec.RoleARN)
	}
	if r.ko.Spec.RoutingConfig != nil {
		f5 := &svcsdk.RoutingConfig{}
		if r.ko.Spec.RoutingConfig.FailoverConfig != nil {
			f5f0 := &svcsdk.FailoverConfig{}
			if r.ko.Spec.RoutingConfig.FailoverConfig.Primary != nil {
				f5f0f0 := &svcsdk.Primary{}
				if r.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck != nil {
					f5f0f0.SetHealthCheck(*r.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck)
				}
				f5f0.SetPrimary(f5f0f0)
			}
			if r.ko.Spec.RoutingConfig.FailoverConfig.Secondary != nil {
				f5f0f1 := &svcsdk.Secondary{}
				if r.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route != nil {
					f5f0f1.SetRoute(*r.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route)
				}
				f5f0.SetSecondary(f5f0f1)
			}
			f5.SetFailoverConfig(f5f0)
		}
		res.SetRoutingConfig(f5)
	}

	return res, nil
}
