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

package rule

import (
	"context"
	"fmt"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eventbridge"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
	pkgtags "github.com/aws-controllers-k8s/eventbridge-controller/pkg/tags"
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

func validateRuleSpec(spec v1alpha1.RuleSpec) error {
	var match bool
	if s := spec.State; s != nil {
		allowedValues := svcsdk.RuleState_Values()
		for _, v := range allowedValues {
			if *s == v {
				match = true
			}
		}
		if !match {
			return newValidationError(
				"spec.state",
				fmt.Sprintf("supported states: %v", allowedValues),
			)
		}
	}

	emptyPattern := spec.EventPattern == nil || *spec.EventPattern == ""
	emptySchedule := spec.ScheduleExpression == nil || *spec.ScheduleExpression == ""

	if emptySchedule && emptyPattern {
		return newValidationError(
			"spec",
			fmt.Sprintf("at least one of %q or %q must be specified",
				"spec.eventPattern", "spec.scheduleExpression"),
		)
	}

	// TODO (@embano1): until code-gen can generate required markers for custom_field
	for _, t := range spec.Targets {
		arn := t.ARN
		id := t.ID

		if arn == nil || *arn == "" || id == nil || *id == "" {
			return newValidationError(
				"spec.targets",
				fmt.Sprintf("%q and %q must be specified for each target", "arn", "id"),
			)
		}
	}

	return nil
}

// setResourceAdditionalFields will set the fields that are not returned by
// DescribeRule calls
func (rm *resourceManager) setResourceAdditionalFields(
	ctx context.Context,
	ko *svcapitypes.Rule,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setResourceAdditionalFields")
	defer func() { exit(err) }()

	if ko.Status.ACKResourceMetadata != nil && ko.Status.ACKResourceMetadata.ARN != nil &&
		*ko.Status.ACKResourceMetadata.ARN != "" {
		// Set event data store tags
		ko.Spec.Tags, err = rm.getTags(ctx, string(*ko.Status.ACKResourceMetadata.ARN))
		if err != nil {
			return err
		}

		ko.Spec.Targets, err = rm.getTargets(ctx, *ko.Spec.Name, *ko.Spec.EventBusName)
		if err != nil {
			return err
		}
	}

	return nil
}

func customPreCompare(
	delta *ackcompare.Delta,
	desired *resource,
	latest *resource,
) {
	if len(desired.ko.Spec.Tags) != len(latest.ko.Spec.Tags) {
		delta.Add("Spec.Tags", desired.ko.Spec.Tags, latest.ko.Spec.Tags)
	}

	if !pkgtags.EqualTags(desired.ko.Spec.Tags, latest.ko.Spec.Tags) {
		delta.Add("Spec.Tags", desired.ko.Spec.Tags, latest.ko.Spec.Tags)
	}

	if len(desired.ko.Spec.Targets) != len(latest.ko.Spec.Targets) {
		delta.Add("Spec.Targets", desired.ko.Spec.Targets, latest.ko.Spec.Targets)
	}

	if !equalTargets(desired.ko.Spec.Targets, latest.ko.Spec.Targets) {
		delta.Add("Spec.Targets", desired.ko.Spec.Targets, latest.ko.Spec.Targets)
	}

	// ideally EqualStrings should do but we're missing one case there (see
	// EqualStrings function comment)
	desiredExpression := desired.ko.Spec.ScheduleExpression
	latestExpression := latest.ko.Spec.ScheduleExpression
	if !equalScheduleExpression(desiredExpression, latestExpression) {
		delta.Add("Spec.ScheduleExpression", desiredExpression, latestExpression)
	}

	desiredBusName := desired.ko.Spec.EventBusName
	latestBusName := latest.ko.Spec.EventBusName
	if !equalEventBusName(desiredBusName, latestBusName) {
		delta.Add("Spec.EventBusName", desiredBusName, latestBusName)
	}
}

func equalScheduleExpression(desiredExpression, latestExpression *string) bool {
	// fast pass: empty/nil string equality (supersedes HasNilDifference)
	if pkgtags.EqualZeroString(desiredExpression) && pkgtags.EqualZeroString(latestExpression) {
		return true
	}

	if ackcompare.HasNilDifference(desiredExpression, latestExpression) {
		return false
	} else if desiredExpression != nil && latestExpression != nil {
		if *desiredExpression != *latestExpression {
			return false
		}
	}

	return true
}

// equalEventBusName is a helper function comparing the provided event bus
// names. A "default" and nil value is treated as equal.
// @embano1: fixes #aws-controllers-k8s/community/issues/1989
func equalEventBusName(desiredBus, latestBus *string) bool {
	isDefaultBus := func(name *string) bool {
		return name == nil || *name == "default" || *name == ""
	}

	if isDefaultBus(desiredBus) && isDefaultBus(latestBus) {
		return true
	}

	if ackcompare.HasNilDifference(desiredBus, latestBus) {
		return false
	} else if desiredBus != nil && latestBus != nil {
		if *desiredBus != *latestBus {
			return false
		}
	}
	return true
}

// unsetScheduleExpression is a helper function to unset the ScheduleExpression
// if the spec field value is an empty string
// @embano1: fixes #aws-controllers-k8s/community/issues/1984
func unsetScheduleExpression(spec v1alpha1.RuleSpec, input *svcsdk.PutRuleInput) {
	if pkgtags.EqualZeroString(spec.ScheduleExpression) {
		input.ScheduleExpression = nil
	}
}
