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
	"reflect"
	"testing"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

func Test_validateRuleSpec(t *testing.T) {
	type args struct {
		spec v1alpha1.RuleSpec
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid state",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String("invalid"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid state (empty string)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String(""),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid target (missing arn)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:        aws.String("ENABLED"),
					EventPattern: aws.String(`{"some":"pattern"}`),
					Targets: []*v1alpha1.Target{
						{
							ARN: nil,
							ID:  aws.String("some-id"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid target (missing id)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:        aws.String("ENABLED"),
					EventPattern: aws.String(`{"some":"pattern"}`),
					Targets: []*v1alpha1.Target{
						{
							ARN: aws.String("some-arn"),
							ID:  nil,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid state (lower case)",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String("enabled"),
				},
			},
			wantErr: true,
		},
		{
			name: "valid state, pattern missing",
			args: args{
				spec: v1alpha1.RuleSpec{
					State: aws.String("ENABLED"),
				},
			},
			wantErr: true,
		},
		{
			name: "valid state, rule and schedule pattern specified",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:              aws.String("ENABLED"),
					EventPattern:       aws.String(`{"some":"pattern"}`),
					ScheduleExpression: aws.String(`{"someschedule}`),
				},
			},
			wantErr: false,
		},
		{
			name: "valid state and rule pattern",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:        aws.String("ENABLED"),
					EventPattern: aws.String(`{"some":"pattern"}`), // we don't verify rule syntax
				},
			},
			wantErr: false,
		},
		{
			name: "valid state and schedule pattern",
			args: args{
				spec: v1alpha1.RuleSpec{
					State:              aws.String("ENABLED"),
					ScheduleExpression: aws.String(`{"someschedule"}`), // we don't verify rule syntax
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRuleSpec(tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("validateRuleSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_equalScheduleExpression(t *testing.T) {
	var (
		emptyString = ""
		someString  = "test-string"
		otherString = "test-other-string"
	)

	type args struct {
		desiredExpression *string
		latestExpression  *string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal: both values are nil",
			args: args{
				desiredExpression: nil,
				latestExpression:  nil,
			},
			want: true,
		}, {
			name: "equal: desired value is nil, latest is empty string",
			args: args{
				desiredExpression: nil,
				latestExpression:  &emptyString,
			},
			want: true,
		}, {
			name: "equal: desired value is empty string, latest is nil",
			args: args{
				desiredExpression: &emptyString,
				latestExpression:  nil,
			},
			want: true,
		}, {
			name: "equal: desired value is empty string, latest is nil",
			args: args{
				desiredExpression: &emptyString,
				latestExpression:  nil,
			},
			want: true,
		}, {
			name: "not equal: desired value is empty string, latest has value",
			args: args{
				desiredExpression: &emptyString,
				latestExpression:  &someString,
			},
			want: false,
		}, {
			name: "not equal: desired has value, latest nil",
			args: args{
				desiredExpression: &someString,
				latestExpression:  nil,
			},
			want: false,
		}, {
			name: "not equal: desired has value, latest has different value",
			args: args{
				desiredExpression: &someString,
				latestExpression:  &otherString,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalScheduleExpression(tt.args.desiredExpression, tt.args.latestExpression); got != tt.want {
				t.Errorf("equalScheduleExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_equalEventBusName(t *testing.T) {
	var (
		defaultEventBusName = "default"
		emptyString         = ""
		customEventBusName  = "custom"
	)
	type args struct {
		desiredExpression *string
		latestExpression  *string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal: desired nil, latest default",
			args: args{
				desiredExpression: nil,
				latestExpression:  &defaultEventBusName,
			},
			want: true,
		}, {
			name: "equal: desired default, latest nil",
			args: args{
				desiredExpression: nil,
				latestExpression:  &defaultEventBusName,
			},
			want: true,
		}, {
			name: "equal: desired empty string, latest nil",
			args: args{
				desiredExpression: &emptyString,
				latestExpression:  nil,
			},
			want: true,
		}, {
			name: "equal: desired is default, latest empty string",
			args: args{
				desiredExpression: &defaultEventBusName,
				latestExpression:  &emptyString,
			},
			want: true,
		}, {
			name: "equal: both nil",
			args: args{
				desiredExpression: nil,
				latestExpression:  nil,
			},
			want: true,
		}, {
			name: "equal: both same default value",
			args: args{
				desiredExpression: &defaultEventBusName,
				latestExpression:  &defaultEventBusName,
			},
			want: true,
		}, {
			name: "equal: both same custom value",
			args: args{
				desiredExpression: &customEventBusName,
				latestExpression:  &customEventBusName,
			},
			want: true,
		}, {
			name: "not equal: desired nil, latest custom value",
			args: args{
				desiredExpression: nil,
				latestExpression:  &customEventBusName,
			},
			want: false,
		}, {
			name: "not equal: desired default, latest custom value",
			args: args{
				desiredExpression: &defaultEventBusName,
				latestExpression:  &customEventBusName,
			},
			want: false,
		}, {
			name: "not equal: desired custom value, latest default",
			args: args{
				desiredExpression: &defaultEventBusName,
				latestExpression:  &customEventBusName,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalEventBusName(tt.args.desiredExpression, tt.args.latestExpression); got != tt.want {
				t.Errorf("equalEventBusName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tagsDifference(t *testing.T) {
	tag := func(k, v string) *v1alpha1.Tag {
		return &v1alpha1.Tag{Key: &k, Value: &v}
	}

	tests := []struct {
		name        string
		desired     []*v1alpha1.Tag
		latest      []*v1alpha1.Tag
		wantAdded   map[string]string
		wantRemoved map[string]string
	}{
		{
			name:        "nil desired and latest tags",
			desired:     nil,
			latest:      nil,
			wantAdded:   map[string]string{},
			wantRemoved: map[string]string{},
		},
		{
			name:        "identical tags in different order",
			desired:     []*v1alpha1.Tag{tag("akey", "avalue"), tag("bkey", "bvalue")},
			latest:      []*v1alpha1.Tag{tag("bkey", "bvalue"), tag("akey", "avalue")},
			wantAdded:   map[string]string{},
			wantRemoved: map[string]string{},
		},
		{
			name:        "add tags when latest is empty",
			desired:     []*v1alpha1.Tag{tag("akey", "avalue"), tag("bkey", "bvalue")},
			latest:      nil,
			wantAdded:   map[string]string{"akey": "avalue", "bkey": "bvalue"},
			wantRemoved: map[string]string{},
		},
		{
			name:        "remove tags when desired is empty",
			desired:     nil,
			latest:      []*v1alpha1.Tag{tag("akey", "avalue")},
			wantAdded:   map[string]string{},
			wantRemoved: map[string]string{"akey": "avalue"},
		},
		// GetTagsDifference has upsert semantics: a changed value appears in "added" with
		// the new value and NOT in "removed", because TagResource overwrites existing keys.
		// No UntagResource call is needed.
		{
			name:        "value change is in added and not in removed",
			desired:     []*v1alpha1.Tag{tag("akey", "avalue-new")},
			latest:      []*v1alpha1.Tag{tag("akey", "avalue-old")},
			wantAdded:   map[string]string{"akey": "avalue-new"},
			wantRemoved: map[string]string{},
		},
		{
			name:        "add, remove and value change together",
			desired:     []*v1alpha1.Tag{tag("akey", "avalue"), tag("ckey", "cvalue")},
			latest:      []*v1alpha1.Tag{tag("akey", "avalue-old"), tag("bkey", "bvalue")},
			wantAdded:   map[string]string{"akey": "avalue", "ckey": "cvalue"},
			wantRemoved: map[string]string{"bkey": "bvalue"},
		},
		{
			name: "nil key is skipped without panic",
			desired: []*v1alpha1.Tag{
				{Key: nil, Value: nil},
				tag("akey", "avalue"),
			},
			latest:      []*v1alpha1.Tag{{Key: nil, Value: nil}},
			wantAdded:   map[string]string{"akey": "avalue"},
			wantRemoved: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, _ := convertToOrderedACKTags(tt.latest)
			to, _ := convertToOrderedACKTags(tt.desired)
			added, _, removed := ackcompare.GetTagsDifference(from, to)
			if !reflect.DeepEqual(map[string]string(added), tt.wantAdded) {
				t.Errorf("GetTagsDifference() added = %v, want %v", added, tt.wantAdded)
			}
			if !reflect.DeepEqual(map[string]string(removed), tt.wantRemoved) {
				t.Errorf("GetTagsDifference() removed = %v, want %v", removed, tt.wantRemoved)
			}
		})
	}
}
