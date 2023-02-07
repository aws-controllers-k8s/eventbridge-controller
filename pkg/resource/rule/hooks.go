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
}
