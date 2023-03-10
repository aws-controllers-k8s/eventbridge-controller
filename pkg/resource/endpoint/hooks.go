// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package endpoint

import (
	"errors"
	"fmt"
	"time"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"golang.org/x/exp/slices"

	svcsdk "github.com/aws/aws-sdk-go/service/eventbridge"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/eventbridge-controller/pkg/tags"
)

const (
	defaultRequeueDelay = time.Second * 5
)

var (
	requeueWaitWhileCreating = ackrequeue.NeededAfter(
		fmt.Errorf("endpoint in status %q, requeueing", svcsdk.EndpointStateCreating),
		defaultRequeueDelay,
	)

	requeueWaitWhileUpdating = ackrequeue.NeededAfter(
		fmt.Errorf("endpoint in status %q, cannot be modified or deleted", svcsdk.EndpointStateUpdating),
		defaultRequeueDelay,
	)

	requeueWaitWhileDeleting = ackrequeue.NeededAfter(
		fmt.Errorf("endpoint in status %q, cannot be modified or deleted", svcsdk.EndpointStateDeleting),
		defaultRequeueDelay,
	)
)

type validationError struct {
	field   string
	message string
}

func (v validationError) Error() string {
	return fmt.Sprintf("invalid Spec: %q: %s", v.field, v.message)
}

func newValidationError(field, message string) validationError {
	return validationError{
		field:   field,
		message: message,
	}
}

func validateEndpointSpec(delta *ackcompare.Delta, spec v1alpha1.EndpointSpec) error {
	if err := validateEventBus(spec); err != nil {
		return err
	}

	if spec.RoutingConfig == nil || spec.RoutingConfig.FailoverConfig == nil {
		return newValidationError("spec.routingConfig.failoverConfig", "must be set")
	}

	if delta != nil && delta.DifferentAt("Spec.RoleARN") {
		roleARN := spec.RoleARN
		if roleARN == nil || *roleARN == "" {
			return newValidationError("spec.roleARN", "unsetting this field is not supported")
		}
	}

	return nil
}

func validateEventBus(spec v1alpha1.EndpointSpec) error {
	if len(spec.EventBuses) != 2 {
		return newValidationError("spec.eventBuses", "must contain exactly two event buses")
	}

	// event bus names must be identical
	arns := make([]string, 2)
	for i, b := range spec.EventBuses {
		if b.EventBusARN == nil {
			return newValidationError("spec.eventBuses", "event bus arn must be set")
		}
		arnInfo, err := arn.Parse(*b.EventBusARN)
		if err != nil {
			return newValidationError("spec.eventBuses", fmt.Sprintf("invalid arn %q", *b.EventBusARN))
		}
		arns[i] = arnInfo.Resource
	}

	if arns[0] != arns[1] {
		return newValidationError("spec.eventBuses", "event bus names must be identical")
	}
	return nil
}

// endpointAvailable returns true if the supplied Endpoint is in an available
// status
func endpointAvailable(r *resource) bool {
	if r.ko.Status.State == nil {
		return false
	}
	state := *r.ko.Status.State
	return state == svcsdk.EndpointStateActive
}

// endpointInMutatingState returns true if the supplied Endpoint is in the process of
// being created
func endpointInMutatingState(r *resource) bool {
	if r.ko.Status.State == nil {
		return false
	}
	state := *r.ko.Status.State
	return state == svcsdk.EndpointStateCreating || state == svcsdk.EndpointStateUpdating || state == svcsdk.EndpointStateDeleting
}

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the Endpoint cannot be modified until it reaches an available
// status.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.State == nil {
		return nil
	}
	msg := fmt.Sprintf("Endpoint is in status %q, cannot be modified.", *r.ko.Status.State)
	return ackrequeue.NeededAfter(
		errors.New(msg),
		defaultRequeueDelay,
	)
}

// if an optional desired field value is nil explicitly unset it in the request
// input
func unsetRemovedSpecFields(
	delta *ackcompare.Delta,
	spec v1alpha1.EndpointSpec,
	input *eventbridge.UpdateEndpointInput,
) {
	if delta.DifferentAt("Spec.Description") {
		if spec.Description == nil {
			input.SetDescription("")
		}
	}

	if delta.DifferentAt("Spec.ReplicationConfig") {
		if spec.ReplicationConfig == nil {
			input.SetReplicationConfig(&eventbridge.ReplicationConfig{State: aws.String("ENABLED")})
		}
	}
}

func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	aDescr := a.ko.Spec.Description
	bDescr := b.ko.Spec.Description

	if !tags.EqualStrings(aDescr, bDescr) {
		delta.Add("Spec.Description", aDescr, bDescr)
	}

	aRole := a.ko.Spec.RoleARN
	bRole := b.ko.Spec.RoleARN

	if !tags.EqualStrings(aRole, bRole) {
		delta.Add("Spec.RoleARN", aRole, bRole)
	}

	aReplCfg := a.ko.Spec.ReplicationConfig
	bReplCfg := b.ko.Spec.ReplicationConfig

	if !equalReplicationConfigs(aReplCfg, bReplCfg) {
		delta.Add("Spec.ReplicationConfig", aReplCfg, bReplCfg)
	}

	aBusCfg := a.ko.Spec.EventBuses
	bBusCfg := b.ko.Spec.EventBuses

	if !equalEventBusConfigs(aBusCfg, bBusCfg) {
		delta.Add("Spec.EventBuses", aReplCfg, bReplCfg)
	}
}

func equalEventBusConfigs(a, b []*v1alpha1.EndpointEventBus) bool {
	sortFn := func(a, b *v1alpha1.EndpointEventBus) bool { return *a.EventBusARN < *b.EventBusARN }
	slices.SortFunc(a, sortFn)
	slices.SortFunc(b, sortFn)

	equalFn := func(a, b *v1alpha1.EndpointEventBus) bool { return *a.EventBusARN == *b.EventBusARN }
	return slices.EqualFunc(a, b, equalFn)
}

func equalReplicationConfigs(a, b *v1alpha1.ReplicationConfig) bool {
	// assumes API always returns replication config
	if (a == nil || a.State == nil || *a.State == "" || *a.State == "ENABLED") && *b.State == "ENABLED" {
		return true
	}

	if a != nil && a.State != nil && *a.State == "DISABLED" && *b.State == "DISABLED" {
		return true
	}

	return false
}
