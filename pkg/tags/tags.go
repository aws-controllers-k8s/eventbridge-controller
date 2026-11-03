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
	"github.com/aws/aws-sdk-go-v2/aws"
)

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
