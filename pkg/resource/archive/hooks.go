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

package archive

import (
	"errors"
	"fmt"

	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eventbridge"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// TerminalStatuses are the status strings that are terminal states for an
// Archive
var TerminalStatuses = []string{
	string(svcsdktypes.ArchiveStateCreateFailed),
	string(svcsdktypes.ArchiveStateUpdateFailed),
}

// archiveInTerminalState returns whether the supplied Archive is in a terminal
// state
func archiveInTerminalState(r *resource) bool {
	if r.ko.Status.State == nil {
		return false
	}
	state := *r.ko.Status.State
	for _, s := range TerminalStatuses {
		if state == s {
			return true
		}
	}
	return false
}

// archiveAvailable returns true if the supplied Archive is in an available
// status and can be modified
func archiveAvailable(r *resource) bool {
	if r.ko.Status.State == nil {
		return false
	}
	state := *r.ko.Status.State
	// Archive can be modified when ENABLED or DISABLED
	return state == string(svcsdktypes.ArchiveStateEnabled) ||
		state == string(svcsdktypes.ArchiveStateDisabled)
}

// archiveModifying returns true if the supplied Archive is in the process of
// being created
func archiveModifying(r *resource) bool {
	if r.ko.Status.State == nil {
		return false
	}
	state := *r.ko.Status.State
	return state == string(svcsdktypes.ArchiveStateCreating) || state == string(svcsdktypes.ArchiveStateUpdating)
}

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the Archive cannot be modified until it reaches an available status
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.State == nil {
		return nil
	}
	status := *r.ko.Status.State
	msg := fmt.Sprintf(
		"Archive is in status %q, cannot be modified.",
		status,
	)
	return ackrequeue.NeededAfter(
		errors.New(msg),
		ackrequeue.DefaultRequeueAfterDuration,
	)
}

// if an optional desired field value is nil explicitly unset it in the request
// input
func unsetRemovedSpecFields(
	spec v1alpha1.ArchiveSpec,
	input *svcsdk.UpdateArchiveInput,
) {
	if spec.EventPattern == nil {
		input.EventPattern = nil
	}

	if spec.Description == nil {
		input.Description = nil
	}

	if spec.RetentionDays == nil {
		input.RetentionDays = nil
	}
}
