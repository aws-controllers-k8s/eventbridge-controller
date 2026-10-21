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

package endpoint

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

// A nil requeue typed as *RequeueNeededAfter is still a non-nil error for the
// runtime, which reads a zero duration from it and drops the resource from the
// queue, so the helper must return a real requeue whatever the state is.
func TestRequeueWaitUntilCanModifyNeverReturnsNil(t *testing.T) {
	for _, state := range []*string{nil, aws.String("UPDATING")} {
		r := &resource{ko: &svcapitypes.Endpoint{}}
		r.ko.Status.State = state
		rq := requeueWaitUntilCanModify(r)
		if rq == nil {
			t.Fatalf("state %v: requeue is nil", aws.ToString(state))
		}
		if rq.Duration() <= 0 {
			t.Errorf("state %v: duration = %v, want a positive delay", aws.ToString(state), rq.Duration())
		}
		if want := endpointState(r); !strings.Contains(rq.Error(), want) {
			t.Errorf("state %v: message %q does not name the state %q", aws.ToString(state), rq.Error(), want)
		}
	}
}

func TestRequeueWaitAfterUpdateHasReason(t *testing.T) {
	if requeueWaitAfterUpdate.Error() == "" {
		t.Error("the requeue carries no reason for the log line")
	}
	if requeueWaitAfterUpdate.Duration() != defaultRequeueDelay {
		t.Errorf("duration = %v, want %v", requeueWaitAfterUpdate.Duration(), defaultRequeueDelay)
	}
}
