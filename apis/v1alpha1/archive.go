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

// ArchiveSpec defines the desired state of Archive.
//
// An Archive object that contains details about an archive.
type ArchiveSpec struct {

	// The name for the archive to create.
	// +kubebuilder:validation:Required
	ArchiveName *string `json:"archiveName"`
	// A description for the archive.
	Description *string `json:"description,omitempty"`
	// An event pattern to use to filter events sent to the archive.
	EventPattern *string `json:"eventPattern,omitempty"`
	// The ARN of the event bus that sends events to the archive.
	EventSourceARN *string                                  `json:"eventSourceARN,omitempty"`
	EventSourceRef *ackv1alpha1.AWSResourceReferenceWrapper `json:"eventSourceRef,omitempty"`
	// The number of days to retain events for. Default value is 0. If set to 0,
	// events are retained indefinitely
	RetentionDays *int64 `json:"retentionDays,omitempty"`
}

// ArchiveStatus defines the observed state of Archive
type ArchiveStatus struct {
	// All CRs managed by ACK have a common `Status.ACKResourceMetadata` member
	// that is used to contain resource sync state, account ownership,
	// constructed ARN for the resource
	// +kubebuilder:validation:Optional
	ACKResourceMetadata *ackv1alpha1.ResourceMetadata `json:"ackResourceMetadata"`
	// All CRS managed by ACK have a common `Status.Conditions` member that
	// contains a collection of `ackv1alpha1.Condition` objects that describe
	// the various terminal states of the CR and its backend AWS service API
	// resource
	// +kubebuilder:validation:Optional
	Conditions []*ackv1alpha1.Condition `json:"conditions"`
	// The time at which the archive was created.
	// +kubebuilder:validation:Optional
	CreationTime *metav1.Time `json:"creationTime,omitempty"`
	// The state of the archive that was created.
	// +kubebuilder:validation:Optional
	State *string `json:"state,omitempty"`
	// The reason that the archive is in the state.
	// +kubebuilder:validation:Optional
	StateReason *string `json:"stateReason,omitempty"`
}

// Archive is the Schema for the Archives API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ARN",type=string,priority=0,JSONPath=`.status.ackResourceMetadata.arn`
// +kubebuilder:printcolumn:name="STATE",type=string,priority=0,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="SYNCED",type=string,priority=0,JSONPath=`.status.conditions[?(@.type=="ACK.ResourceSynced")].status`
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"
type Archive struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ArchiveSpec   `json:"spec,omitempty"`
	Status            ArchiveStatus `json:"status,omitempty"`
}

// ArchiveList contains a list of Archive
// +kubebuilder:object:root=true
type ArchiveList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Archive `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Archive{}, &ArchiveList{})
}