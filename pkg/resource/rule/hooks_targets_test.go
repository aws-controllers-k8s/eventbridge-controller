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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"gotest.tools/v3/assert"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

const (
	ruleName  = "test-rule"
	busName   = "test-bus"
	arnFormat = "arn:service:%d"
	idFormat  = "id-%d"
)

func Test_validateTargets(t *testing.T) {
	tests := []struct {
		name    string
		targets []*svcapitypes.Target
		wantErr string
	}{
		{
			name:    "empty list of targets",
			targets: nil,
			wantErr: "",
		}, {
			name: "two targets, one without id",
			targets: []*svcapitypes.Target{
				{
					ARN: aws.String("arn:1"),
					ID:  nil,
				}, {
					ARN: aws.String("arn:2"),
					ID:  aws.String("id2"),
				},
			},
			wantErr: "invalid target: target ID and ARN must be specified",
		}, {
			name: "two targets, one without arn",
			targets: []*svcapitypes.Target{
				{
					ARN: aws.String("arn:1"),
					ID:  aws.String("id1"),
				}, {
					ARN: nil,
					ID:  aws.String("id2"),
				},
			},
			wantErr: "invalid target: target ID and ARN must be specified",
		}, {
			name: "two targets, duplicate ids",
			targets: []*svcapitypes.Target{
				{
					ARN: aws.String("arn:1"),
					ID:  aws.String("id1"),
				}, {
					ARN: aws.String("arn:2"),
					ID:  aws.String("id1"),
				},
			},
			wantErr: "invalid target: unique target ID is already used",
		}, {
			name: "two valid targets, different ids same arn",
			targets: []*svcapitypes.Target{
				{
					ARN: aws.String("arn:1"),
					ID:  aws.String("id1"),
				}, {
					ARN: aws.String("arn:1"),
					ID:  aws.String("id2"),
				},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTargets(tt.targets)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}
