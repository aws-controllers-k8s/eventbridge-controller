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
	"errors"
	"reflect"

	"github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	ackutil "github.com/aws-controllers-k8s/runtime/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eventbridge"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
	pkgtags "github.com/aws-controllers-k8s/eventbridge-controller/pkg/tags"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// TODO(embano1): add more input validation
func validateTargets(targets []*svcapitypes.Target) error {
	seen := make(map[string]bool)

	for _, t := range targets {
		if pkgtags.EqualZeroString(t.ID) || pkgtags.EqualZeroString(t.ARN) {
			return errors.New("invalid target: target ID and ARN must be specified")
		}

		if seen[*t.ID] {
			return errors.New("invalid target: unique target ID is already used")
		}

		seen[*t.ID] = true
	}

	return nil
}

// getTags retrieves a resource list of tags.
func (rm *resourceManager) getTargets(ctx context.Context, rule, bus string) (targets []*svcapitypes.Target, err error) {
	rlog := log.FromContext(ctx)
	exit := rlog.Trace("rm.getTargets")
	defer func() { exit(err) }()

	var listTargetsResponse *svcsdk.ListTargetsByRuleOutput
	listTargetsResponse, err = rm.sdkapi.ListTargetsByRule(
		ctx,
		&svcsdk.ListTargetsByRuleInput{
			EventBusName: aws.String(bus),
			Rule:         aws.String(rule),
		},
	)
	rm.metrics.RecordAPICall("GET", "ListTargetsByRule", err)
	if err != nil {
		return nil, err
	}
	// Convert []Target to []*Target
	sdkTargets := make([]*svcsdktypes.Target, len(listTargetsResponse.Targets))
	for i := range listTargetsResponse.Targets {
		sdkTargets[i] = &listTargetsResponse.Targets[i]
	}
	return resourceTargetsFromSDKTargets(sdkTargets), nil
}

// syncTargets synchronizes rule targets
func (rm *resourceManager) syncTargets(
	ctx context.Context,
	ruleName *string,
	eventBus *string, // name or arn
	desired, latest []*v1alpha1.Target,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncTargets")
	defer func() { exit(err) }()

	added, removed := computeTargetsDelta(latest, desired)

	if len(removed) > 0 {
		// Convert []*string to []string
		tagKeys := make([]string, len(removed))
		for i, key := range removed {
			tagKeys[i] = *key
		}

		_, err = rm.sdkapi.RemoveTargets(
			ctx,
			&svcsdk.RemoveTargetsInput{
				// NOTE(a-hilaly,embano1): we might need to force the removal, in some cases?
				// thinking annotations... terminal conditions...
				Rule:         ruleName,
				EventBusName: eventBus,
				Ids:          tagKeys, // Use converted slice
			})
		rm.metrics.RecordAPICall("UPDATE", "RemoveTargets", err)
		if err != nil {
			return err
		}
	}

	if len(added) > 0 {
		sdkTargets, err := sdkTargetsFromResourceTargets(added)
		if err != nil {
			return err
		}

		// Convert []*svcsdktypes.Target to []svcsdktypes.Target
		targets := make([]svcsdktypes.Target, len(sdkTargets))
		for i, t := range sdkTargets {
			targets[i] = *t
		}

		_, err = rm.sdkapi.PutTargets(
			ctx,
			&svcsdk.PutTargetsInput{
				Rule:         ruleName,
				EventBusName: eventBus,
				Targets:      targets,
			})
		rm.metrics.RecordAPICall("UPDATE", "PutTargets", err)
		if err != nil {
			return err
		}
	}
	return nil
}

// computeTargetsDelta computes the delta between the specified targets and
// returns added and removed targets
func computeTargetsDelta(
	a []*svcapitypes.Target,
	b []*svcapitypes.Target,
) (added []*svcapitypes.Target, removed []*string) {
	var visitedIndexes []string
mainLoop:
	for _, aElement := range a {
		visitedIndexes = append(visitedIndexes, *aElement.ID)
		for _, bElement := range b {
			if pkgtags.EqualStrings(aElement.ID, bElement.ID) {
				if !reflect.DeepEqual(aElement, bElement) {
					added = append(added, bElement)
				}
				continue mainLoop
			}
		}
		removed = append(removed, aElement.ID)
	}
	for _, bElement := range b {
		if !ackutil.InStrings(*bElement.ID, visitedIndexes) {
			added = append(added, bElement)
		}
	}
	return added, removed
}

// equalTargets returns true if two Tag arrays are equal regardless of the order
// of their elements.
func equalTargets(
	a []*svcapitypes.Target,
	b []*svcapitypes.Target,
) bool {
	added, removed := computeTargetsDelta(a, b)
	return len(added) == 0 && len(removed) == 0
}
