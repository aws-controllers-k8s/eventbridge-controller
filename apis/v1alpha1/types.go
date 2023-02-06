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
	"github.com/aws/aws-sdk-go/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = &aws.JSONValue{}
	_ = ackv1alpha1.AWSAccountID("")
)

// This structure specifies the VPC subnets and security groups for the task,
// and whether a public IP address is to be used. This structure is relevant
// only for ECS tasks that use the awsvpc network mode.
type AWSVPCConfiguration struct {
	AssignPublicIP *string   `json:"assignPublicIP,omitempty"`
	SecurityGroups []*string `json:"securityGroups,omitempty"`
	Subnets        []*string `json:"subnets,omitempty"`
}

// An Archive object that contains details about an archive.
type Archive struct {
	EventSourceARN *string `json:"eventSourceARN,omitempty"`
}

// The array properties for the submitted job, such as the size of the array.
// The array size can be between 2 and 10,000. If you specify array properties
// for a job, it becomes an array job. This parameter is used only if the target
// is an Batch job.
type BatchArrayProperties struct {
	Size *int64 `json:"size,omitempty"`
}

// The custom parameters to be used when the target is an Batch job.
type BatchParameters struct {
	// The array properties for the submitted job, such as the size of the array.
	// The array size can be between 2 and 10,000. If you specify array properties
	// for a job, it becomes an array job. This parameter is used only if the target
	// is an Batch job.
	ArrayProperties *BatchArrayProperties `json:"arrayProperties,omitempty"`
	JobDefinition   *string               `json:"jobDefinition,omitempty"`
	JobName         *string               `json:"jobName,omitempty"`
	// The retry strategy to use for failed jobs, if the target is an Batch job.
	// If you specify a retry strategy here, it overrides the retry strategy defined
	// in the job definition.
	RetryStrategy *BatchRetryStrategy `json:"retryStrategy,omitempty"`
}

// The retry strategy to use for failed jobs, if the target is an Batch job.
// If you specify a retry strategy here, it overrides the retry strategy defined
// in the job definition.
type BatchRetryStrategy struct {
	Attempts *int64 `json:"attempts,omitempty"`
}

// The details of a capacity provider strategy. To learn more, see CapacityProviderStrategyItem
// (https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_CapacityProviderStrategyItem.html)
// in the Amazon ECS API Reference.
type CapacityProviderStrategyItem struct {
	Base             *int64  `json:"base,omitempty"`
	CapacityProvider *string `json:"capacityProvider,omitempty"`
	Weight           *int64  `json:"weight,omitempty"`
}

// A JSON string which you can use to limit the event bus permissions you are
// granting to only accounts that fulfill the condition. Currently, the only
// supported condition is membership in a certain Amazon Web Services organization.
// The string must contain Type, Key, and Value fields. The Value field specifies
// the ID of the Amazon Web Services organization. Following is an example value
// for Condition:
//
// '{"Type" : "StringEquals", "Key": "aws:PrincipalOrgID", "Value": "o-1234567890"}'
type Condition struct {
	Key   *string `json:"key,omitempty"`
	Type  *string `json:"type_,omitempty"`
	Value *string `json:"value,omitempty"`
}

// Additional parameter included in the body. You can include up to 100 additional
// body parameters per request. An event payload cannot exceed 64 KB.
type ConnectionBodyParameter struct {
	IsValueSecret *bool   `json:"isValueSecret,omitempty"`
	Key           *string `json:"key,omitempty"`
	Value         *string `json:"value,omitempty"`
}

// Additional parameter included in the header. You can include up to 100 additional
// header parameters per request. An event payload cannot exceed 64 KB.
type ConnectionHeaderParameter struct {
	IsValueSecret *bool   `json:"isValueSecret,omitempty"`
	Value         *string `json:"value,omitempty"`
}

// Additional query string parameter for the connection. You can include up
// to 100 additional query string parameters per request. Each additional parameter
// counts towards the event payload size, which cannot exceed 64 KB.
type ConnectionQueryStringParameter struct {
	IsValueSecret *bool   `json:"isValueSecret,omitempty"`
	Value         *string `json:"value,omitempty"`
}

// A DeadLetterConfig object that contains information about a dead-letter queue
// configuration.
type DeadLetterConfig struct {
	ARN *string `json:"arn,omitempty"`
}

// The custom parameters to be used when the target is an Amazon ECS task.
type ECSParameters struct {
	CapacityProviderStrategy []*CapacityProviderStrategyItem `json:"capacityProviderStrategy,omitempty"`
	EnableECSManagedTags     *bool                           `json:"enableECSManagedTags,omitempty"`
	EnableExecuteCommand     *bool                           `json:"enableExecuteCommand,omitempty"`
	Group                    *string                         `json:"group,omitempty"`
	LaunchType               *string                         `json:"launchType,omitempty"`
	// This structure specifies the network configuration for an ECS task.
	NetworkConfiguration *NetworkConfiguration  `json:"networkConfiguration,omitempty"`
	PlacementConstraints []*PlacementConstraint `json:"placementConstraints,omitempty"`
	PlacementStrategy    []*PlacementStrategy   `json:"placementStrategy,omitempty"`
	PlatformVersion      *string                `json:"platformVersion,omitempty"`
	PropagateTags        *string                `json:"propagateTags,omitempty"`
	ReferenceID          *string                `json:"referenceID,omitempty"`
	Tags                 []*Tag                 `json:"tags,omitempty"`
	TaskCount            *int64                 `json:"taskCount,omitempty"`
	TaskDefinitionARN    *string                `json:"taskDefinitionARN,omitempty"`
}

// An event bus receives events from a source and routes them to rules associated
// with that event bus. Your account's default event bus receives events from
// Amazon Web Services services. A custom event bus can receive events from
// your custom applications and services. A partner event bus receives events
// from an event source created by an SaaS partner. These events come from the
// partners services or applications.
type EventBus_SDK struct {
	ARN    *string `json:"arn,omitempty"`
	Name   *string `json:"name,omitempty"`
	Policy *string `json:"policy,omitempty"`
}

// A partner event source is created by an SaaS partner. If a customer creates
// a partner event bus that matches this event source, that Amazon Web Services
// account can receive events from the partner's applications or services.
type EventSource struct {
	ARN       *string `json:"arn,omitempty"`
	CreatedBy *string `json:"createdBy,omitempty"`
	Name      *string `json:"name,omitempty"`
}

// These are custom parameter to be used when the target is an API Gateway REST
// APIs or EventBridge ApiDestinations. In the latter case, these are merged
// with any InvocationParameters specified on the Connection, with any values
// from the Connection taking precedence.
type HTTPParameters struct {
	HeaderParameters      map[string]*string `json:"headerParameters,omitempty"`
	PathParameterValues   []*string          `json:"pathParameterValues,omitempty"`
	QueryStringParameters map[string]*string `json:"queryStringParameters,omitempty"`
}

// Contains the parameters needed for you to provide custom input to a target
// based on one or more pieces of data extracted from the event.
type InputTransformer struct {
	InputPathsMap map[string]*string `json:"inputPathsMap,omitempty"`
	InputTemplate *string            `json:"inputTemplate,omitempty"`
}

// This object enables you to specify a JSON path to extract from the event
// and use as the partition key for the Amazon Kinesis data stream, so that
// you can control the shard to which the event goes. If you do not include
// this parameter, the default is to use the eventId as the partition key.
type KinesisParameters struct {
	PartitionKeyPath *string `json:"partitionKeyPath,omitempty"`
}

// This structure specifies the network configuration for an ECS task.
type NetworkConfiguration struct {
	// This structure specifies the VPC subnets and security groups for the task,
	// and whether a public IP address is to be used. This structure is relevant
	// only for ECS tasks that use the awsvpc network mode.
	AWSvpcConfiguration *AWSVPCConfiguration `json:"awsvpcConfiguration,omitempty"`
}

// A partner event source is created by an SaaS partner. If a customer creates
// a partner event bus that matches this event source, that Amazon Web Services
// account can receive events from the partner's applications or services.
type PartnerEventSource struct {
	ARN  *string `json:"arn,omitempty"`
	Name *string `json:"name,omitempty"`
}

// An object representing a constraint on task placement. To learn more, see
// Task Placement Constraints (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-placement-constraints.html)
// in the Amazon Elastic Container Service Developer Guide.
type PlacementConstraint struct {
	Expression *string `json:"expression,omitempty"`
	Type       *string `json:"type_,omitempty"`
}

// The task placement strategy for a task or service. To learn more, see Task
// Placement Strategies (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-placement-strategies.html)
// in the Amazon Elastic Container Service Service Developer Guide.
type PlacementStrategy struct {
	Field *string `json:"field,omitempty"`
	Type  *string `json:"type_,omitempty"`
}

// Represents an event to be submitted.
type PutEventsRequestEntry struct {
	Detail     *string `json:"detail,omitempty"`
	DetailType *string `json:"detailType,omitempty"`
	Source     *string `json:"source,omitempty"`
}

// The details about an event generated by an SaaS partner.
type PutPartnerEventsRequestEntry struct {
	Detail     *string `json:"detail,omitempty"`
	DetailType *string `json:"detailType,omitempty"`
	Source     *string `json:"source,omitempty"`
}

// Represents a target that failed to be added to a rule.
type PutTargetsResultEntry struct {
	TargetID *string `json:"targetID,omitempty"`
}

// These are custom parameters to be used when the target is a Amazon Redshift
// cluster to invoke the Amazon Redshift Data API ExecuteStatement based on
// EventBridge events.
type RedshiftDataParameters struct {
	Database         *string `json:"database,omitempty"`
	DBUser           *string `json:"dbUser,omitempty"`
	SecretManagerARN *string `json:"secretManagerARN,omitempty"`
	Sql              *string `json:"sql,omitempty"`
	StatementName    *string `json:"statementName,omitempty"`
	WithEvent        *bool   `json:"withEvent,omitempty"`
}

// Represents a target that failed to be removed from a rule.
type RemoveTargetsResultEntry struct {
	TargetID *string `json:"targetID,omitempty"`
}

// A Replay object that contains details about a replay.
type Replay struct {
	EventSourceARN *string `json:"eventSourceARN,omitempty"`
}

// A ReplayDestination object that contains details about a replay.
type ReplayDestination struct {
	ARN *string `json:"arn,omitempty"`
}

// A RetryPolicy object that includes information about the retry policy settings.
type RetryPolicy struct {
	MaximumEventAgeInSeconds *int64 `json:"maximumEventAgeInSeconds,omitempty"`
	MaximumRetryAttempts     *int64 `json:"maximumRetryAttempts,omitempty"`
}

// Contains information about a rule in Amazon EventBridge.
type Rule_SDK struct {
	ARN                *string `json:"arn,omitempty"`
	Description        *string `json:"description,omitempty"`
	EventBusName       *string `json:"eventBusName,omitempty"`
	EventPattern       *string `json:"eventPattern,omitempty"`
	ManagedBy          *string `json:"managedBy,omitempty"`
	Name               *string `json:"name,omitempty"`
	RoleARN            *string `json:"roleARN,omitempty"`
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
	State              *string `json:"state,omitempty"`
}

// This parameter contains the criteria (either InstanceIds or a tag) used to
// specify which EC2 instances are to be sent the command.
type RunCommandParameters struct {
	RunCommandTargets []*RunCommandTarget `json:"runCommandTargets,omitempty"`
}

// Information about the EC2 instances that are to be sent the command, specified
// as key-value pairs. Each RunCommandTarget block can include only one key,
// but this key may specify multiple values.
type RunCommandTarget struct {
	Key    *string   `json:"key,omitempty"`
	Values []*string `json:"values,omitempty"`
}

// This structure includes the custom parameter to be used when the target is
// an SQS FIFO queue.
type SQSParameters struct {
	MessageGroupID *string `json:"messageGroupID,omitempty"`
}

// Name/Value pair of a parameter to start execution of a SageMaker Model Building
// Pipeline.
type SageMakerPipelineParameter struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

// These are custom parameters to use when the target is a SageMaker Model Building
// Pipeline that starts based on EventBridge events.
type SageMakerPipelineParameters struct {
	PipelineParameterList []*SageMakerPipelineParameter `json:"pipelineParameterList,omitempty"`
}

// A key-value pair associated with an Amazon Web Services resource. In EventBridge,
// rules and event buses support tagging.
type Tag struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// Targets are the resources to be invoked when a rule is triggered. For a complete
// list of services and resources that can be set as a target, see PutTargets
// (https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutTargets.html).
//
// If you are setting the event bus of another account as the target, and that
// account granted permission to your account through an organization instead
// of directly by the account ID, then you must specify a RoleArn with proper
// permissions in the Target structure. For more information, see Sending and
// Receiving Events Between Amazon Web Services Accounts (https://docs.aws.amazon.com/eventbridge/latest/userguide/eventbridge-cross-account-event-delivery.html)
// in the Amazon EventBridge User Guide.
type Target struct {
	ARN *string `json:"arn,omitempty"`
	// The custom parameters to be used when the target is an Batch job.
	BatchParameters *BatchParameters `json:"batchParameters,omitempty"`
	// A DeadLetterConfig object that contains information about a dead-letter queue
	// configuration.
	DeadLetterConfig *DeadLetterConfig `json:"deadLetterConfig,omitempty"`
	// The custom parameters to be used when the target is an Amazon ECS task.
	ECSParameters *ECSParameters `json:"ecsParameters,omitempty"`
	// These are custom parameter to be used when the target is an API Gateway REST
	// APIs or EventBridge ApiDestinations. In the latter case, these are merged
	// with any InvocationParameters specified on the Connection, with any values
	// from the Connection taking precedence.
	HTTPParameters *HTTPParameters `json:"httpParameters,omitempty"`
	ID             *string         `json:"id,omitempty"`
	Input          *string         `json:"input,omitempty"`
	InputPath      *string         `json:"inputPath,omitempty"`
	// Contains the parameters needed for you to provide custom input to a target
	// based on one or more pieces of data extracted from the event.
	InputTransformer *InputTransformer `json:"inputTransformer,omitempty"`
	// This object enables you to specify a JSON path to extract from the event
	// and use as the partition key for the Amazon Kinesis data stream, so that
	// you can control the shard to which the event goes. If you do not include
	// this parameter, the default is to use the eventId as the partition key.
	KinesisParameters *KinesisParameters `json:"kinesisParameters,omitempty"`
	// These are custom parameters to be used when the target is a Amazon Redshift
	// cluster to invoke the Amazon Redshift Data API ExecuteStatement based on
	// EventBridge events.
	RedshiftDataParameters *RedshiftDataParameters `json:"redshiftDataParameters,omitempty"`
	// A RetryPolicy object that includes information about the retry policy settings.
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
	RoleARN     *string      `json:"roleARN,omitempty"`
	// This parameter contains the criteria (either InstanceIds or a tag) used to
	// specify which EC2 instances are to be sent the command.
	RunCommandParameters *RunCommandParameters `json:"runCommandParameters,omitempty"`
	// These are custom parameters to use when the target is a SageMaker Model Building
	// Pipeline that starts based on EventBridge events.
	SageMakerPipelineParameters *SageMakerPipelineParameters `json:"sageMakerPipelineParameters,omitempty"`
	// This structure includes the custom parameter to be used when the target is
	// an SQS FIFO queue.
	SQSParameters *SQSParameters `json:"sqsParameters,omitempty"`
}
