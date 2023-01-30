package event_bus

import (
	"context"
	"fmt"
	"strings"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/aws-controllers-k8s/runtime/pkg/errors"
	"github.com/aws-controllers-k8s/runtime/pkg/util"
	svcsdk "github.com/aws/aws-sdk-go/service/eventbridge"

	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// setResourceAdditionalFields will set the fields that are not returned by
// DescribeEventBus calls
func (rm *resourceManager) setResourceAdditionalFields(
	ctx context.Context,
	ko *svcapitypes.EventBus,
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
	}
	return nil
}

// getTags retrieves a resource list of tags.
func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN string,
) (tags []*svcapitypes.Tag, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.getTags")
	defer func() { exit(err) }()

	var listTagsResponse *svcsdk.ListTagsForResourceOutput
	listTagsResponse, err = rm.sdkapi.ListTagsForResourceWithContext(
		ctx,
		&svcsdk.ListTagsForResourceInput{
			ResourceARN: &resourceARN,
		},
	)
	rm.metrics.RecordAPICall("GET", "ListTagsForResource", err)
	if err != nil {
		return nil, err
	}
	for _, tag := range listTagsResponse.Tags {
		tags = append(tags, &svcapitypes.Tag{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	return tags, nil
}

// customUpdate implements a custom logic for handling EventBus resource
// updates.
func (rm *resourceManager) customUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	var err error
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateEventBus")
	defer func() { exit(err) }()

	if immutableFieldChanges := rm.getImmutableFieldChanges(delta); len(immutableFieldChanges) > 0 {
		msg := fmt.Sprintf("Immutable Spec fields have been modified: %s", strings.Join(immutableFieldChanges, ","))
		return nil, errors.NewTerminalError(fmt.Errorf(msg))
	}

	if delta.DifferentAt("Spec.Tags") {
		err = rm.syncTags(ctx, latest, desired)
		if err != nil {
			return nil, err
		}
	}
	return desired, nil
}

// syncTags updates event bus tags
func (rm *resourceManager) syncTags(
	ctx context.Context,
	latest *resource,
	desired *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncEventBusTags")
	defer func() { exit(err) }()

	missing, extra := computeTagsDelta(desired.ko.Spec.Tags, latest.ko.Spec.Tags)

	arn := (*string)(latest.ko.Status.ACKResourceMetadata.ARN)
	if len(extra) > 0 {
		_, err = rm.sdkapi.UntagResourceWithContext(
			ctx,
			&svcsdk.UntagResourceInput{
				ResourceARN: arn,
				TagKeys:     sdkTagStringsFromResourceTags(extra),
			})

		rm.metrics.RecordAPICall("UPDATE", "UntagResource", err)
		if err != nil {
			return err
		}
	}

	if len(missing) > 0 {
		_, err = rm.sdkapi.TagResourceWithContext(
			ctx,
			&svcsdk.TagResourceInput{
				ResourceARN: arn,
				Tags:        sdkTagsFromResourceTags(missing),
			})

		rm.metrics.RecordAPICall("UPDATE", "TagResource", err)
		if err != nil {
			return err
		}
	}
	return nil
}

// computeTagsDelta compares two Tag arrays and return two different lists
// containing the added and removed tags.
// The removed tags list only contains the Key of tags
func computeTagsDelta(
	desired []*svcapitypes.Tag,
	latest []*svcapitypes.Tag,
) (missing, extra []*svcapitypes.Tag) {
	var visitedIndexes []string
mainLoop:
	for _, le := range latest {
		visitedIndexes = append(visitedIndexes, *le.Key)
		for _, de := range desired {
			if equalStrings(le.Key, de.Key) {
				if !equalStrings(le.Value, de.Value) {
					missing = append(missing, de)
				}
				continue mainLoop
			}
		}
		extra = append(extra, le)
	}
	for _, de := range desired {
		if !util.InStrings(*de.Key, visitedIndexes) {
			missing = append(missing, de)
		}
	}
	return missing, extra
}

func equalStrings(a, b *string) bool {
	if a == nil {
		return b == nil || *b == ""
	}

	if a != nil && b == nil {
		return false
	}

	return (*a == "" && b == nil) || *a == *b
}

// sdkTagsFromResourceTags transforms a *svcapitypes.Tag array to a *svcsdk.Tag array.
func sdkTagsFromResourceTags(rTags []*svcapitypes.Tag) []*svcsdk.Tag {
	tags := make([]*svcsdk.Tag, len(rTags))
	for i := range rTags {
		tags[i] = &svcsdk.Tag{
			Key:   rTags[i].Key,
			Value: rTags[i].Value,
		}
	}
	return tags
}

// sdkTagStringsFromResourceTags transforms a *svcapitypes.Tag array to a string array.
func sdkTagStringsFromResourceTags(rTags []*svcapitypes.Tag) []*string {
	tags := make([]*string, len(rTags))
	for i := range rTags {
		tags[i] = rTags[i].Key
	}
	return tags
}

// compareTags is a custom comparison function for comparing lists of Tag
// structs where the order of the structs in the list is not important.
func compareTags(
	delta *ackcompare.Delta,
	desired *resource,
	latest *resource,
) {
	if len(desired.ko.Spec.Tags) != len(latest.ko.Spec.Tags) {
		delta.Add("Spec.Tags", desired.ko.Spec.Tags, latest.ko.Spec.Tags)
		return
	}
	if !equalTags(desired.ko.Spec.Tags, latest.ko.Spec.Tags) {
		delta.Add("Spec.Tags", desired.ko.Spec.Tags, latest.ko.Spec.Tags)
	}
}

// equalTags returns true if two Tag arrays are equal regardless of the order
// of their elements.
func equalTags(
	desired []*svcapitypes.Tag,
	latest []*svcapitypes.Tag,
) bool {
	addedOrUpdated, removed := computeTagsDelta(desired, latest)
	return len(addedOrUpdated) == 0 && len(removed) == 0
}
