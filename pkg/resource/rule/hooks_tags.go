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

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eventbridge"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// getTags retrieves a resource list of tags.
func (rm *resourceManager) getTags(
	ctx context.Context,
	resourceARN string,
) (tags []*svcapitypes.Tag, err error) {
	rlog := log.FromContext(ctx)
	exit := rlog.Trace("rm.getTags")
	defer func() { exit(err) }()

	var listTagsResponse *svcsdk.ListTagsForResourceOutput
	listTagsResponse, err = rm.sdkapi.ListTagsForResource(
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

// syncTags synchronizes rule tags
func (rm *resourceManager) syncTags(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := log.FromContext(ctx)
	exit := rlog.Trace("rm.syncTags")
	defer func() { exit(err) }()

	from, fromOrder := convertToOrderedACKTags(latest.ko.Spec.Tags)
	to, toOrder := convertToOrderedACKTags(desired.ko.Spec.Tags)
	added, _, removed := ackcompare.GetTagsDifference(from, to)
	toAdd := fromACKTags(added, toOrder)
	toDelete := fromACKTags(removed, fromOrder)

	if len(toAdd) == 0 && len(toDelete) == 0 {
		return nil
	}

	arn := (*string)(latest.ko.Status.ACKResourceMetadata.ARN)
	if len(toDelete) > 0 {
		_, err = rm.sdkapi.UntagResource(
			ctx,
			&svcsdk.UntagResourceInput{
				ResourceARN: arn,
				TagKeys:     sdkTagStringsFromResourceTags(toDelete),
			})

		rm.metrics.RecordAPICall("UPDATE", "UntagResource", err)
		if err != nil {
			return err
		}
	}

	if len(toAdd) > 0 {
		_, err = rm.sdkapi.TagResource(
			ctx,
			&svcsdk.TagResourceInput{
				ResourceARN: arn,
				Tags:        sdkTagsFromResourceTags(toAdd),
			})

		rm.metrics.RecordAPICall("UPDATE", "TagResource", err)
		if err != nil {
			return err
		}
	}
	return nil
}

// sdkTagsFromResourceTags transforms a *svcapitypes.Tag array to a *svcsdk.Tag array.
func sdkTagsFromResourceTags(rTags []*svcapitypes.Tag) []svcsdktypes.Tag {
	tags := make([]svcsdktypes.Tag, len(rTags))
	for i := range rTags {
		tags[i] = svcsdktypes.Tag{
			Key:   rTags[i].Key,
			Value: rTags[i].Value,
		}
	}
	return tags
}

// sdkTagStringsFromResourceTags transforms a *svcapitypes.Tag array to a string array.
func sdkTagStringsFromResourceTags(rTags []*svcapitypes.Tag) []string {
	tags := make([]string, len(rTags))
	for i := range rTags {
		tags[i] = *rTags[i].Key
	}
	return tags
}
