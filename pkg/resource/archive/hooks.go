package archive

import (
	"errors"
	"fmt"

	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	svcsdk "github.com/aws/aws-sdk-go/service/eventbridge"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// TerminalStatuses are the status strings that are terminal states for an
// Archive
var TerminalStatuses = []string{
	svcsdk.ArchiveStateCreateFailed,
	svcsdk.ArchiveStateUpdateFailed,
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
// status
func archiveAvailable(r *resource) bool {
	if r.ko.Status.State == nil {
		return false
	}
	state := *r.ko.Status.State
	return state == svcsdk.ArchiveStateEnabled || state == svcsdk.ArchiveStateDisabled
}

// archiveCreating returns true if the supplied Archive is in the process of
// being created
func archiveCreating(r *resource) bool {
	if r.ko.Status.State == nil {
		return false
	}
	state := *r.ko.Status.State
	return state == svcsdk.ArchiveStateCreating
}

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the Archive cannot be modified until it reaches an available
// status.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.State == nil {
		return nil
	}
	status := *r.ko.Status.State
	msg := fmt.Sprintf(
		"Archive in '%s' state, cannot be modified.",
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
		input.SetEventPattern("")
	}

	if spec.Description == nil {
		input.SetDescription("")
	}

	if spec.RetentionDays == nil {
		input.SetRetentionDays(0)
	}
}
