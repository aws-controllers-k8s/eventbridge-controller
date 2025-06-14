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

// EndpointSpec defines the desired state of Endpoint.
//
// A global endpoint used to improve your application's availability by making
// it regional-fault tolerant. For more information about global endpoints,
// see Making applications Regional-fault tolerant with global endpoints and
// event replication (https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-global-endpoints.html)
// in the Amazon EventBridge User Guide .
type EndpointSpec struct {

	// A description of the global endpoint.
	//
	// Regex Pattern: `.*`
	Description *string `json:"description,omitempty"`
	// Define the event buses used.
	//
	// The names of the event buses must be identical in each Region.
	// +kubebuilder:validation:Required
	EventBuses []*EndpointEventBus `json:"eventBuses"`
	// The name of the global endpoint. For example, "Name":"us-east-2-custom_bus_A-endpoint".
	//
	// Regex Pattern: `^[\.\-_A-Za-z0-9]+$`
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable once set"
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// Enable or disable event replication. The default state is ENABLED which means
	// you must supply a RoleArn. If you don't have a RoleArn or you don't want
	// event replication enabled, set the state to DISABLED.
	ReplicationConfig *ReplicationConfig `json:"replicationConfig,omitempty"`
	// The ARN of the role used for replication.
	//
	// Regex Pattern: `^arn:aws[a-z-]*:iam::\d{12}:role\/[\w+=,.@/-]+$`
	RoleARN *string `json:"roleARN,omitempty"`
	// Configure the routing policy, including the health check and secondary Region..
	// +kubebuilder:validation:Required
	RoutingConfig *RoutingConfig `json:"routingConfig"`
}

// EndpointStatus defines the observed state of Endpoint
type EndpointStatus struct {
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
	// The state of the endpoint that was created by this request.
	// +kubebuilder:validation:Optional
	State *string `json:"state,omitempty"`
	// The reason the endpoint you asked for information about is in its current
	// state.
	//
	// Regex Pattern: `.*`
	// +kubebuilder:validation:Optional
	StateReason *string `json:"stateReason,omitempty"`
}

// Endpoint is the Schema for the Endpoints API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ARN",type=string,priority=1,JSONPath=`.status.ackResourceMetadata.arn`
// +kubebuilder:printcolumn:name="STATE",type=string,priority=0,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Synced",type="string",priority=0,JSONPath=".status.conditions[?(@.type==\"ACK.ResourceSynced\")].status"
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"
type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              EndpointSpec   `json:"spec,omitempty"`
	Status            EndpointStatus `json:"status,omitempty"`
}

// EndpointList contains a list of Endpoint
// +kubebuilder:object:root=true
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Endpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Endpoint{}, &EndpointList{})
}
