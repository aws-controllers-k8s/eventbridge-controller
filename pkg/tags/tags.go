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

package tags

import (
	"github.com/aws-controllers-k8s/runtime/pkg/util"
	"github.com/aws/aws-sdk-go-v2/aws"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// ComputeTagsDelta compares two Tag arrays and return two different lists
// containing the added and removed tags. The removed tags list only contains
// the Key of tags
func ComputeTagsDelta(
	desired []*svcapitypes.Tag,
	latest []*svcapitypes.Tag,
) (missing, extra []*svcapitypes.Tag) {
	var visitedIndexes []string
mainLoop:
	for _, le := range latest {
		visitedIndexes = append(visitedIndexes, *le.Key)
		for _, de := range desired {
			if EqualStrings(le.Key, de.Key) {
				if !EqualStrings(le.Value, de.Value) {
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

// EqualTags returns true if two Tag arrays are equal regardless of the order of
// their elements.
func EqualTags(
	desired []*svcapitypes.Tag,
	latest []*svcapitypes.Tag,
) bool {
	addedOrUpdated, removed := ComputeTagsDelta(desired, latest)
	return len(addedOrUpdated) == 0 && len(removed) == 0
}

// EqualStrings returns true if two strings are equal e.g., both are nil, one is
// nil and the other is empty string, or both non-zero strings are equal.
// TODO (@embano1): needs additional case -> a == "" && b == nil return true
func EqualStrings(a, b *string) bool {
	if a == nil {
		return b == nil || *b == ""
	}

	if a != nil && b == nil {
		return false
	}

	return (*a == "" && b == nil) || *a == *b
}

func EqualZeroString(a *string) bool {
	return EqualStrings(a, aws.String(""))
}
