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

// RuleSpec defines the desired state of Rule.
//
// Contains information about a rule in Amazon EventBridge.
type RuleSpec struct {

	// A description of the rule.
	Description *string `json:"description,omitempty"`
	// The name or ARN of the event bus to associate with this rule. If you omit
	// this, the default event bus is used.
	EventBusName *string                                  `json:"eventBusName,omitempty"`
	EventBusRef  *ackv1alpha1.AWSResourceReferenceWrapper `json:"eventBusRef,omitempty"`
	// The event pattern. For more information, see Amazon EventBridge event patterns
	// (https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html)
	// in the Amazon EventBridge User Guide .
	EventPattern *string `json:"eventPattern,omitempty"`
	// The name of the rule that you are creating or updating.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// The Amazon Resource Name (ARN) of the IAM role associated with the rule.
	//
	// If you're setting an event bus in another account as the target and that
	// account granted permission to your account through an organization instead
	// of directly by the account ID, you must specify a RoleArn with proper permissions
	// in the Target structure, instead of here in this parameter.
	RoleARN *string `json:"roleARN,omitempty"`
	// The scheduling expression. For example, "cron(0 20 * * ? *)" or "rate(5 minutes)".
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
	// The state of the rule.
	//
	// Valid values include:
	//
	//   - DISABLED: The rule is disabled. EventBridge does not match any events
	//     against the rule.
	//
	//   - ENABLED: The rule is enabled. EventBridge matches events against the
	//     rule, except for Amazon Web Services management events delivered through
	//     CloudTrail.
	//
	//   - ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS: The rule is enabled for
	//     all events, including Amazon Web Services management events delivered
	//     through CloudTrail. Management events provide visibility into management
	//     operations that are performed on resources in your Amazon Web Services
	//     account. These are also known as control plane operations. For more information,
	//     see Logging management events (https://docs.aws.amazon.com/awscloudtrail/latest/userguide/logging-management-events-with-cloudtrail.html#logging-management-events)
	//     in the CloudTrail User Guide, and Filtering management events from Amazon
	//     Web Services services (https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-service-event.html#eb-service-event-cloudtrail)
	//     in the Amazon EventBridge User Guide . This value is only valid for rules
	//     on the default (https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-what-is-how-it-works-concepts.html#eb-bus-concepts-buses)
	//     event bus or custom event buses (https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-create-event-bus.html).
	//     It does not apply to partner event buses (https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-saas.html).
	State *string `json:"state,omitempty"`
	// The list of key-value pairs to associate with the rule.
	Tags    []*Tag    `json:"tags,omitempty"`
	Targets []*Target `json:"targets,omitempty"`
}

// RuleStatus defines the observed state of Rule
type RuleStatus struct {
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
}

// Rule is the Schema for the Rules API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ARN",type=string,priority=1,JSONPath=`.status.ackResourceMetadata.arn`
// +kubebuilder:printcolumn:name="Synced",type="string",priority=0,JSONPath=".status.conditions[?(@.type==\"ACK.ResourceSynced\")].status"
// +kubebuilder:printcolumn:name="Age",type="date",priority=0,JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:shortName=er
type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RuleSpec   `json:"spec,omitempty"`
	Status            RuleStatus `json:"status,omitempty"`
}

// RuleList contains a list of Rule
// +kubebuilder:object:root=true
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rule{}, &RuleList{})
}
