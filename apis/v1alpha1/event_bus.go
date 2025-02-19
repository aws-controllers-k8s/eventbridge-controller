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

package v1alpha1

import (
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventBusSpec defines the desired state of EventBus.
//
// An event bus receives events from a source, uses rules to evaluate them,
// applies any configured input transformation, and routes them to the appropriate
// target(s). Your account's default event bus receives events from Amazon Web
// Services services. A custom event bus can receive events from your custom
// applications and services. A partner event bus receives events from an event
// source created by an SaaS partner. These events come from the partners services
// or applications.
type EventBusSpec struct {

	// If you are creating a partner event bus, this specifies the partner event
	// source that the new event bus will be matched with.

	EventSourceName *string `json:"eventSourceName,omitempty"`
	// The name of the new event bus.
	//
	// Custom event bus names can't contain the / character, but you can use the
	// / character in partner event bus names. In addition, for partner event buses,
	// the name must exactly match the name of the partner event source that this
	// event bus is matched to.
	//
	// You can't use the name default for a custom event bus, as this name is already
	// used for your account's default event bus.

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable once set"
	// +kubebuilder:validation:Required

	Name *string `json:"name"`
	// Tags to associate with the event bus.

	Tags []*Tag `json:"tags,omitempty"`
}

// EventBusStatus defines the observed state of EventBus
type EventBusStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRs managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	// +kubebuilder:validation:Optional
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
}

// EventBus is the Schema for the EventBuses API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ARN",type=string,priority=1,JSONPath=`.status.ackResourceMetadata.arn`
// +kubebuilder:printcolumn:name="Synced",type="string",priority=0,JSONPath=".status.conditions[?(@.type==\"ACK.ResourceSynced\")].status"
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=eb;bus
type EventBus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              EventBusSpec   `json:"spec,omitempty"`
	Status            EventBusStatus `json:"status,omitempty"`
}

// EventBusList contains a list of EventBus
// +kubebuilder:object:root=true
type EventBusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventBus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EventBus{}, &EventBusList{})
}
