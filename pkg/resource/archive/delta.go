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

package archive

import (
	"bytes"
	"reflect"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	acktags "github.com/aws-controllers-k8s/runtime/pkg/tags"
)

// Hack to avoid import errors during build...
var (
	_ = &bytes.Buffer{}
	_ = &reflect.Method{}
	_ = &acktags.Tags{}
)

// newResourceDelta returns a new `ackcompare.Delta` used to compare two
// resources
func newResourceDelta(
	a *resource,
	b *resource,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if (a == nil && b != nil) ||
		(a != nil && b == nil) {
		delta.Add("", a, b)
		return delta
	}

	if ackcompare.HasNilDifference(a.ko.Spec.ArchiveName, b.ko.Spec.ArchiveName) {
		delta.Add("Spec.ArchiveName", a.ko.Spec.ArchiveName, b.ko.Spec.ArchiveName)
	} else if a.ko.Spec.ArchiveName != nil && b.ko.Spec.ArchiveName != nil {
		if *a.ko.Spec.ArchiveName != *b.ko.Spec.ArchiveName {
			delta.Add("Spec.ArchiveName", a.ko.Spec.ArchiveName, b.ko.Spec.ArchiveName)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Description, b.ko.Spec.Description) {
		delta.Add("Spec.Description", a.ko.Spec.Description, b.ko.Spec.Description)
	} else if a.ko.Spec.Description != nil && b.ko.Spec.Description != nil {
		if *a.ko.Spec.Description != *b.ko.Spec.Description {
			delta.Add("Spec.Description", a.ko.Spec.Description, b.ko.Spec.Description)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.EventPattern, b.ko.Spec.EventPattern) {
		delta.Add("Spec.EventPattern", a.ko.Spec.EventPattern, b.ko.Spec.EventPattern)
	} else if a.ko.Spec.EventPattern != nil && b.ko.Spec.EventPattern != nil {
		if *a.ko.Spec.EventPattern != *b.ko.Spec.EventPattern {
			delta.Add("Spec.EventPattern", a.ko.Spec.EventPattern, b.ko.Spec.EventPattern)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.EventSourceARN, b.ko.Spec.EventSourceARN) {
		delta.Add("Spec.EventSourceARN", a.ko.Spec.EventSourceARN, b.ko.Spec.EventSourceARN)
	} else if a.ko.Spec.EventSourceARN != nil && b.ko.Spec.EventSourceARN != nil {
		if *a.ko.Spec.EventSourceARN != *b.ko.Spec.EventSourceARN {
			delta.Add("Spec.EventSourceARN", a.ko.Spec.EventSourceARN, b.ko.Spec.EventSourceARN)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.EventSourceRef, b.ko.Spec.EventSourceRef) {
		delta.Add("Spec.EventSourceRef", a.ko.Spec.EventSourceRef, b.ko.Spec.EventSourceRef)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.RetentionDays, b.ko.Spec.RetentionDays) {
		delta.Add("Spec.RetentionDays", a.ko.Spec.RetentionDays, b.ko.Spec.RetentionDays)
	} else if a.ko.Spec.RetentionDays != nil && b.ko.Spec.RetentionDays != nil {
		if *a.ko.Spec.RetentionDays != *b.ko.Spec.RetentionDays {
			delta.Add("Spec.RetentionDays", a.ko.Spec.RetentionDays, b.ko.Spec.RetentionDays)
		}
	}

	return delta
}
