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

package event_bus

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
	pkgtags "github.com/aws-controllers-k8s/eventbridge-controller/pkg/tags"
)

func Test_computeTagsDelta(t *testing.T) {
	type args struct {
		desired []*svcapitypes.Tag
		latest  []*svcapitypes.Tag
	}
	tests := []struct {
		name        string
		args        args
		wantMissing []*svcapitypes.Tag
		wantExtra   []*svcapitypes.Tag
	}{
		{
			name: "nil values on desired and latest tags",
			args: args{
				desired: nil,
				latest:  nil,
			},
			wantMissing: nil,
			wantExtra:   nil,
		},
		{
			name: "desired tags nil, latest with one tag",
			args: args{
				desired: nil,
				latest: []*svcapitypes.Tag{
					{
						Key:   aws.String("akey"),
						Value: aws.String("avalue"),
					},
				},
			},
			wantMissing: nil,
			wantExtra: []*svcapitypes.Tag{{
				Key:   aws.String("akey"),
				Value: aws.String("avalue"),
			}},
		},
		{
			name: "desired with two tags, latest with one tag with different value",
			args: args{
				desired: []*svcapitypes.Tag{
					{
						Key:   aws.String("akey"),
						Value: aws.String("avalue"),
					},
					{
						Key:   aws.String("bkey"),
						Value: aws.String("bvalue"),
					},
				},
				latest: []*svcapitypes.Tag{
					{
						Key:   aws.String("akey"),
						Value: aws.String("avalue-old"),
					},
				},
			},
			wantMissing: []*svcapitypes.Tag{
				{
					Key:   aws.String("akey"),
					Value: aws.String("avalue"),
				},
				{
					Key:   aws.String("bkey"),
					Value: aws.String("bvalue"),
				},
			},
			wantExtra: nil,
		},
		{
			name: "desired with three tags, latest with two tags one with same value one with different value",
			args: args{
				desired: []*svcapitypes.Tag{
					{
						Key:   aws.String("akey"),
						Value: aws.String("avalue"),
					},
					{
						Key:   aws.String("bkey"),
						Value: aws.String("bvalue"),
					},
					{
						Key:   aws.String("ckey"),
						Value: aws.String("cvalue"),
					},
				},
				latest: []*svcapitypes.Tag{
					{
						Key:   aws.String("akey"),
						Value: aws.String("avalue"),
					},
					{
						Key:   aws.String("bkey"),
						Value: aws.String("bvalue-old"),
					},
				},
			},
			wantMissing: []*svcapitypes.Tag{
				{
					Key:   aws.String("bkey"),
					Value: aws.String("bvalue"),
				},
				{
					Key:   aws.String("ckey"),
					Value: aws.String("cvalue"),
				},
			},
			wantExtra: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMissing, gotExtra := pkgtags.ComputeTagsDelta(tt.args.desired, tt.args.latest)
			if !reflect.DeepEqual(gotMissing, tt.wantMissing) {
				t.Errorf("computeTagsDelta() gotMissing = %v, want %v", gotMissing, tt.wantMissing)
			}
			if !reflect.DeepEqual(gotExtra, tt.wantExtra) {
				t.Errorf("computeTagsDelta() gotExtra = %v, want %v", gotExtra, tt.wantExtra)
			}
		})
	}
}
