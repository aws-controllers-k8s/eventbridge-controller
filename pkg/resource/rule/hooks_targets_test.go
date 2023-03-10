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
	"errors"
	"fmt"
	"testing"

	"github.com/aws-controllers-k8s/runtime/pkg/metrics"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/eventbridge/eventbridgeiface"
	"gotest.tools/v3/assert"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

const (
	ruleName  = "test-rule"
	busName   = "test-bus"
	arnFormat = "arn:service:%d"
	idFormat  = "id-%d"
)

type ebAPIMockTargetsClient struct {
	eventbridgeiface.EventBridgeAPI
	putInput    *eventbridge.PutTargetsInput
	removeInput *eventbridge.RemoveTargetsInput
	calls       int
}

func (eb *ebAPIMockTargetsClient) PutTargetsWithContext(_ aws.Context, input *eventbridge.PutTargetsInput, _ ...request.Option) (*eventbridge.PutTargetsOutput, error) {
	eb.calls++
	eb.putInput = input

	if len(input.Targets) > 5 {
		return nil, errors.New("the requested resource exceeds the maximum number allowed")
	}

	return nil, nil
}

func (eb *ebAPIMockTargetsClient) RemoveTargetsWithContext(_ aws.Context, input *eventbridge.RemoveTargetsInput, _ ...request.Option) (*eventbridge.RemoveTargetsOutput, error) {
	eb.calls++
	eb.removeInput = input

	return nil, nil
}

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

func Test_resourceManager_syncTargets(t *testing.T) {
	tests := []struct {
		name            string
		rule            string
		bus             string
		latest          func() []*svcapitypes.Target
		desired         func() []*svcapitypes.Target
		wantCalls       int
		wantPutInput    *eventbridge.PutTargetsInput
		wantRemoveInput *eventbridge.RemoveTargetsInput
		wantErr         string
	}{
		{
			name:   "fails when adding more than 5 targets to an existing rule without targets",
			rule:   ruleName,
			bus:    busName,
			latest: func() []*svcapitypes.Target { return nil },
			desired: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 6)
				for i := 0; i < 6; i++ {
					targets[i] = createTarget(i)
				}
				return targets
			},
			wantCalls: 1,
			wantPutInput: &eventbridge.PutTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Targets: []*eventbridge.Target{
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 0)),
						Id:  aws.String(fmt.Sprintf(idFormat, 0)),
					},
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 1)),
						Id:  aws.String(fmt.Sprintf(idFormat, 1)),
					},
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 2)),
						Id:  aws.String(fmt.Sprintf(idFormat, 2)),
					},
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 3)),
						Id:  aws.String(fmt.Sprintf(idFormat, 3)),
					},
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 4)),
						Id:  aws.String(fmt.Sprintf(idFormat, 4)),
					},
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 5)),
						Id:  aws.String(fmt.Sprintf(idFormat, 5)),
					},
				},
			},
			wantRemoveInput: nil,
			wantErr:         "the requested resource exceeds the maximum number allowed",
		}, {
			name:            "no change to rule with no targets",
			rule:            ruleName,
			bus:             busName,
			latest:          func() []*svcapitypes.Target { return nil },
			desired:         func() []*svcapitypes.Target { return nil },
			wantCalls:       0,
			wantPutInput:    nil,
			wantRemoveInput: nil,
			wantErr:         "",
		}, {
			name:   "add two targets to rule with no targets",
			rule:   ruleName,
			bus:    busName,
			latest: func() []*svcapitypes.Target { return nil },
			desired: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 2)
				for i := 0; i < 2; i++ {
					targets[i] = createTarget(i)
				}
				return targets
			},
			wantCalls: 1,
			wantPutInput: &eventbridge.PutTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Targets: []*eventbridge.Target{
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 0)),
						Id:  aws.String(fmt.Sprintf(idFormat, 0)),
					},
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 1)),
						Id:  aws.String(fmt.Sprintf(idFormat, 1)),
					},
				},
			},
			wantRemoveInput: nil,
			wantErr:         "",
		}, {
			name: "add one, remove one from existing rule with two targets",
			rule: ruleName,
			bus:  busName,
			latest: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 2)
				for i := 0; i < 2; i++ {
					targets[i] = createTarget(i)
				}
				return targets
			},
			desired: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 2)
				targets[0] = createTarget(1) // means first target removed
				targets[1] = createTarget(2) // added target

				return targets
			},
			wantCalls: 2,
			wantPutInput: &eventbridge.PutTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Targets: []*eventbridge.Target{
					{
						Arn: aws.String(fmt.Sprintf(arnFormat, 2)),
						Id:  aws.String(fmt.Sprintf(idFormat, 2)),
					},
				},
			},
			wantRemoveInput: &eventbridge.RemoveTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Ids:          []*string{aws.String("id-0")},
			},
			wantErr: "",
		}, {
			name: "remove all from existing rule with two targets",
			rule: ruleName,
			bus:  busName,
			latest: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 2)
				for i := 0; i < 2; i++ {
					targets[i] = createTarget(i)
				}
				return targets
			},
			desired:      func() []*svcapitypes.Target { return nil },
			wantCalls:    1,
			wantPutInput: nil,
			wantRemoveInput: &eventbridge.RemoveTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Ids:          []*string{aws.String("id-0"), aws.String("id-1")},
			},
			wantErr: "",
		}, {
			name: "update one target from rule with three targets",
			rule: ruleName,
			bus:  busName,
			latest: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 3)
				for i := 0; i < 3; i++ {
					targets[i] = createTarget(i)
				}
				return targets
			},
			desired: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 3)
				for i := 0; i < 3; i++ {
					targets[i] = createTarget(i)
				}

				// update first target
				targets[0].Input = aws.String("some input")
				return targets
			},
			wantCalls: 1,
			wantPutInput: &eventbridge.PutTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Targets: []*eventbridge.Target{
					{
						Arn:   aws.String(fmt.Sprintf(arnFormat, 0)),
						Id:    aws.String(fmt.Sprintf(idFormat, 0)),
						Input: aws.String("some input"),
					},
				},
			},
			wantRemoveInput: nil,
			wantErr:         "",
		}, {
			name: "add one, update one, remove one target from rule with two targets",
			rule: ruleName,
			bus:  busName,
			latest: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 2)
				for i := 0; i < 2; i++ {
					targets[i] = createTarget(i)
				}
				return targets
			},
			desired: func() []*svcapitypes.Target {
				targets := make([]*svcapitypes.Target, 2)
				targets[0] = createTarget(1) // means first target removed
				targets[1] = createTarget(2) // added target

				// update first target
				targets[0].Input = aws.String("some input")
				return targets
			},
			wantCalls: 2,
			wantPutInput: &eventbridge.PutTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Targets: []*eventbridge.Target{
					{
						Arn:   aws.String(fmt.Sprintf(arnFormat, 1)),
						Id:    aws.String(fmt.Sprintf(idFormat, 1)),
						Input: aws.String("some input"),
					}, {
						Arn: aws.String(fmt.Sprintf(arnFormat, 2)),
						Id:  aws.String(fmt.Sprintf(idFormat, 2)),
					},
				},
			},
			wantRemoveInput: &eventbridge.RemoveTargetsInput{
				EventBusName: aws.String(busName),
				Rule:         aws.String(ruleName),
				Ids:          []*string{aws.String("id-0")},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ebClient := ebAPIMockTargetsClient{}
			rm := &resourceManager{
				metrics: metrics.NewMetrics("eventbridge"),
				sdkapi:  &ebClient,
			}

			err := rm.syncTargets(context.TODO(), &tt.rule, &tt.bus, tt.desired(), tt.latest())
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
			}

			assert.Equal(t, ebClient.calls, tt.wantCalls)
			assert.DeepEqual(t, ebClient.putInput, tt.wantPutInput)
			assert.DeepEqual(t, ebClient.removeInput, tt.wantRemoveInput)
		})
	}
}

func createTarget(id int) *svcapitypes.Target {
	return &svcapitypes.Target{
		ARN: aws.String(fmt.Sprintf(arnFormat, id)),
		ID:  aws.String(fmt.Sprintf(idFormat, id)),
	}
}
