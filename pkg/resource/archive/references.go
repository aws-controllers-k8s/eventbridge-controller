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

// Code generated by ack-generate. DO NOT EDIT.

package archive

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	acktypes "github.com/aws-controllers-k8s/runtime/pkg/types"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// ResolveReferences finds if there are any Reference field(s) present
// inside AWSResource passed in the parameter and attempts to resolve
// those reference field(s) into target field(s).
// It returns an AWSResource with resolved reference(s), and an error if the
// passed AWSResource's reference field(s) cannot be resolved.
// This method also adds/updates the ConditionTypeReferencesResolved for the
// AWSResource.
func (rm *resourceManager) ResolveReferences(
	ctx context.Context,
	apiReader client.Reader,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	namespace := res.MetaObject().GetNamespace()
	ko := rm.concreteResource(res).ko.DeepCopy()
	err := validateReferenceFields(ko)
	if err == nil {
		err = resolveReferenceForEventSourceARN(ctx, apiReader, namespace, ko)
	}

	// If there was an error while resolving any reference, reset all the
	// resolved values so that they do not get persisted inside etcd
	if err != nil {
		ko = rm.concreteResource(res).ko.DeepCopy()
	}
	if hasNonNilReferences(ko) {
		return ackcondition.WithReferencesResolvedCondition(&resource{ko}, err)
	}
	return &resource{ko}, err
}

// validateReferenceFields validates the reference field and corresponding
// identifier field.
func validateReferenceFields(ko *svcapitypes.Archive) error {
	if ko.Spec.EventSourceRef != nil && ko.Spec.EventSourceARN != nil {
		return ackerr.ResourceReferenceAndIDNotSupportedFor("EventSourceARN", "EventSourceRef")
	}
	if ko.Spec.EventSourceRef == nil && ko.Spec.EventSourceARN == nil {
		return ackerr.ResourceReferenceOrIDRequiredFor("EventSourceARN", "EventSourceRef")
	}
	return nil
}

// hasNonNilReferences returns true if resource contains a reference to another
// resource
func hasNonNilReferences(ko *svcapitypes.Archive) bool {
	return false || (ko.Spec.EventSourceRef != nil)
}

// resolveReferenceForEventSourceARN reads the resource referenced
// from EventSourceRef field and sets the EventSourceARN
// from referenced resource
func resolveReferenceForEventSourceARN(
	ctx context.Context,
	apiReader client.Reader,
	namespace string,
	ko *svcapitypes.Archive,
) error {
	if ko.Spec.EventSourceRef != nil &&
		ko.Spec.EventSourceRef.From != nil {
		arr := ko.Spec.EventSourceRef.From
		if arr == nil || arr.Name == nil || *arr.Name == "" {
			return fmt.Errorf("provided resource reference is nil or empty")
		}
		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      *arr.Name,
		}
		obj := svcapitypes.EventBus{}
		err := apiReader.Get(ctx, namespacedName, &obj)
		if err != nil {
			return err
		}
		var refResourceSynced, refResourceTerminal bool
		for _, cond := range obj.Status.Conditions {
			if cond.Type == ackv1alpha1.ConditionTypeResourceSynced &&
				cond.Status == corev1.ConditionTrue {
				refResourceSynced = true
			}
			if cond.Type == ackv1alpha1.ConditionTypeTerminal &&
				cond.Status == corev1.ConditionTrue {
				refResourceTerminal = true
			}
		}
		if refResourceTerminal {
			return ackerr.ResourceReferenceTerminalFor(
				"EventBus",
				namespace, *arr.Name)
		}
		if !refResourceSynced {
			return ackerr.ResourceReferenceNotSyncedFor(
				"EventBus",
				namespace, *arr.Name)
		}
		if obj.Status.ACKResourceMetadata == nil || obj.Status.ACKResourceMetadata.ARN == nil {
			return ackerr.ResourceReferenceMissingTargetFieldFor(
				"EventBus",
				namespace, *arr.Name,
				"Status.ACKResourceMetadata.ARN")
		}
		referencedValue := string(*obj.Status.ACKResourceMetadata.ARN)
		ko.Spec.EventSourceARN = &referencedValue
	}
	return nil
}