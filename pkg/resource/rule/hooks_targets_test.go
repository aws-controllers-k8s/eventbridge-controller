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

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"

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

// Test_computeTargetsDelta_ignoresReferenceFields covers the interaction
// between the hand-written targets delta and the cross-resource reference
// companion fields the code-generator adds for a Targets.* `references` block.
//
// The desired target carries the companion (the user set roleRef); the latest
// target never can, because it is rebuilt from the ListTargetsByRule response.
// If the comparison did not ignore the companion, every one of these cases
// would report a difference, the rule would never converge, and syncTargets
// would re-issue PutTargets on every reconcile.
func Test_computeTargetsDelta_ignoresReferenceFields(t *testing.T) {
	roleARN := "arn:aws:iam::123456789012:role/my-target-role"

	// desired: user supplied roleRef, which the runtime has already resolved
	// into the concrete RoleARN field before the delta is computed.
	desiredWithRef := func() *svcapitypes.Target {
		return &svcapitypes.Target{
			ID:      aws.String("id-1"),
			ARN:     aws.String("arn:service:1"),
			RoleARN: aws.String(roleARN),
			RoleRef: &ackv1alpha1.AWSResourceReferenceWrapper{
				From: &ackv1alpha1.AWSResourceReference{
					Name: aws.String("my-target-role"),
				},
			},
		}
	}

	// latest: read back from AWS, concrete value only, never a companion.
	latestConcrete := func() *svcapitypes.Target {
		return &svcapitypes.Target{
			ID:      aws.String("id-1"),
			ARN:     aws.String("arn:service:1"),
			RoleARN: aws.String(roleARN),
		}
	}

	tests := []struct {
		name        string
		latest      []*svcapitypes.Target
		desired     []*svcapitypes.Target
		wantAdded   int
		wantRemoved int
	}{
		{
			name:        "resolved reference vs concrete readback is not a delta",
			latest:      []*svcapitypes.Target{latestConcrete()},
			desired:     []*svcapitypes.Target{desiredWithRef()},
			wantAdded:   0,
			wantRemoved: 0,
		}, {
			name:   "a real difference is still detected alongside a reference",
			latest: []*svcapitypes.Target{latestConcrete()},
			desired: []*svcapitypes.Target{
				func() *svcapitypes.Target {
					tgt := desiredWithRef()
					tgt.Input = aws.String(`{"changed":true}`)
					return tgt
				}(),
			},
			wantAdded:   1,
			wantRemoved: 0,
		}, {
			name:   "a differing resolved role ARN is still detected",
			latest: []*svcapitypes.Target{latestConcrete()},
			desired: []*svcapitypes.Target{
				func() *svcapitypes.Target {
					tgt := desiredWithRef()
					tgt.RoleARN = aws.String("arn:aws:iam::123456789012:role/other-role")
					return tgt
				}(),
			},
			wantAdded:   1,
			wantRemoved: 0,
		}, {
			name:   "companion on a nested member is ignored too",
			latest: []*svcapitypes.Target{latestConcrete()},
			desired: []*svcapitypes.Target{
				func() *svcapitypes.Target {
					tgt := desiredWithRef()
					tgt.DeadLetterConfig = &svcapitypes.DeadLetterConfig{
						ARN: aws.String("arn:aws:sqs:us-west-2:123456789012:dlq"),
					}
					return tgt
				}(),
			},
			// The nested concrete ARN is a genuine difference: latest has no
			// DeadLetterConfig at all. This asserts the walk does not flatten
			// real nested changes away.
			wantAdded:   1,
			wantRemoved: 0,
		}, {
			name:        "target present in desired only is added",
			latest:      nil,
			desired:     []*svcapitypes.Target{desiredWithRef()},
			wantAdded:   1,
			wantRemoved: 0,
		}, {
			name:        "target present in latest only is removed",
			latest:      []*svcapitypes.Target{latestConcrete()},
			desired:     nil,
			wantAdded:   0,
			wantRemoved: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Argument order matches syncTargets: a=latest, b=desired, so
			// `added` carries the desired elements to send to PutTargets.
			added, removed := computeTargetsDelta(tt.latest, tt.desired)
			assert.Equal(t, len(added), tt.wantAdded)
			assert.Equal(t, len(removed), tt.wantRemoved)

			for _, a := range added {
				assert.Assert(t, a != nil)
			}
		})
	}
}

// Test_targetWithoutReferences_doesNotMutateInput guards the deep copy: the
// delta runs against the live desired resource, so clearing companions in place
// would strip roleRef from the spec the controller is about to write back.
func Test_targetWithoutReferences_doesNotMutateInput(t *testing.T) {
	in := &svcapitypes.Target{
		ID:      aws.String("id-1"),
		ARN:     aws.String("arn:service:1"),
		RoleARN: aws.String("arn:aws:iam::123456789012:role/my-target-role"),
		RoleRef: &ackv1alpha1.AWSResourceReferenceWrapper{
			From: &ackv1alpha1.AWSResourceReference{
				Name: aws.String("my-target-role"),
			},
		},
	}

	out := targetWithoutReferences(in)

	assert.Assert(t, in.RoleRef != nil, "input companion must be preserved")
	assert.Equal(t, *in.RoleRef.From.Name, "my-target-role")
	assert.Assert(t, out.RoleRef == nil, "copy companion must be cleared")
	// The concrete field must survive the clearing, since it is what carries
	// the resolved value into the comparison.
	assert.Equal(t, *out.RoleARN, *in.RoleARN)

	assert.Assert(t, targetWithoutReferences(nil) == nil)
}
