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
	"testing"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gotest.tools/v3/assert"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"

	"github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

func Test_validateEndpointSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    v1alpha1.EndpointSpec
		delta   *ackcompare.Delta
		wantErr string
	}{
		{
			name: "no event buses specified",
			spec: v1alpha1.EndpointSpec{
				EventBuses: nil,
				Name:       aws.String("endpointspec"),
			},
			wantErr: "must contain exactly two event buses",
		},
		{
			name: "only one event bus specified",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
			},
			wantErr: "must contain exactly two event buses",
		},
		{
			name: "more than two event buses specified",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-3:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
			},
			wantErr: "must contain exactly two event buses",
		},
		{
			name: "two event buses one missing arn",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: nil},
				},
				Name: aws.String("endpointspec"),
			},
			wantErr: "event bus arn must be set",
		},
		{
			name: "two event buses one with invalid arn",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("invalid")},
				},
				Name: aws.String("endpointspec"),
			},
			wantErr: "invalid arn",
		},
		{
			name: "two event buses with different names",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:otherBus")},
				},
				Name: aws.String("endpointspec"),
			},
			wantErr: "event bus names must be identical",
		},
		{
			name: "routing config not set",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name:          aws.String("endpointspec"),
				RoutingConfig: nil,
			},
			wantErr: "spec.routingConfig.failoverConfig",
		},
		{
			name: "failover config not set",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name:          aws.String("endpointspec"),
				RoutingConfig: &v1alpha1.RoutingConfig{FailoverConfig: nil},
			},
			wantErr: "spec.routingConfig.failoverConfig",
		},
		{
			name: "valid spec during create",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &v1alpha1.RoutingConfig{FailoverConfig: &v1alpha1.FailoverConfig{
					Primary: &v1alpha1.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &v1alpha1.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
			},
			wantErr: "",
		},
		{
			name: "valid spec with new description during update",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &v1alpha1.RoutingConfig{FailoverConfig: &v1alpha1.FailoverConfig{
					Primary: &v1alpha1.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &v1alpha1.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
				Description: aws.String("some description"),
			},
			delta: &ackcompare.Delta{
				Differences: []*ackcompare.Difference{
					{
						Path: ackcompare.NewPath("Spec.Description"),
						A:    nil,
						B:    aws.String("some description"),
					},
				},
			},
			wantErr: "",
		},
		{
			name: "role unset during update",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &v1alpha1.RoutingConfig{FailoverConfig: &v1alpha1.FailoverConfig{
					Primary: &v1alpha1.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &v1alpha1.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
			},
			delta: &ackcompare.Delta{
				Differences: []*ackcompare.Difference{
					{
						Path: ackcompare.NewPath("Spec.RoleARN"),
						A:    aws.String("arn:aws:iam::1234567890:role/role"),
						B:    nil,
					},
				},
			},
			wantErr: "unsetting this field is not supported",
		},
		{
			name: "role and routing config added during update",
			spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &v1alpha1.RoutingConfig{FailoverConfig: &v1alpha1.FailoverConfig{
					Primary: &v1alpha1.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &v1alpha1.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
				RoleARN: aws.String("arn:aws:iam::1234567890:role/role"),
			},
			delta: &ackcompare.Delta{
				Differences: []*ackcompare.Difference{
					{
						Path: ackcompare.NewPath("Spec.RoleARN"),
						A:    nil,
						B:    aws.String("arn:aws:iam::1234567890:role/role"),
					},
					{
						Path: ackcompare.NewPath("Spec.RoutingConfig"),
						A:    nil,
						B: &v1alpha1.RoutingConfig{FailoverConfig: &v1alpha1.FailoverConfig{
							Primary: &v1alpha1.Primary{
								HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
							},
							Secondary: &v1alpha1.Secondary{
								Route: aws.String("eu-central-1"),
							},
						}},
					},
				},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEndpointSpec(tt.delta, tt.spec)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func Test_unsetRemovedSpecFields(t *testing.T) {
	emtpyString := ""

	tests := []struct {
		name      string
		spec      v1alpha1.EndpointSpec
		input     *eventbridge.UpdateEndpointInput
		delta     *ackcompare.Delta
		wantInput *eventbridge.UpdateEndpointInput
	}{
		{
			name: "description removed",
			spec: v1alpha1.EndpointSpec{
				Description: nil,
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &v1alpha1.RoutingConfig{FailoverConfig: &v1alpha1.FailoverConfig{
					Primary: &v1alpha1.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &v1alpha1.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
			},
			input: &eventbridge.UpdateEndpointInput{
				EventBuses: []svcsdktypes.EndpointEventBus{
					{EventBusArn: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusArn: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &svcsdktypes.RoutingConfig{FailoverConfig: &svcsdktypes.FailoverConfig{
					Primary: &svcsdktypes.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &svcsdktypes.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
			},
			delta: &ackcompare.Delta{
				Differences: []*ackcompare.Difference{
					{
						Path: ackcompare.NewPath("Spec.Description"),
						A:    aws.String("some description"),
						B:    nil,
					},
				},
			},
			wantInput: &eventbridge.UpdateEndpointInput{
				Description: &emtpyString,
				EventBuses: []svcsdktypes.EndpointEventBus{
					{EventBusArn: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusArn: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &svcsdktypes.RoutingConfig{FailoverConfig: &svcsdktypes.FailoverConfig{
					Primary: &svcsdktypes.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &svcsdktypes.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
			},
		},
		{
			name: "replication config removed",
			spec: v1alpha1.EndpointSpec{
				ReplicationConfig: nil,
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusARN: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &v1alpha1.RoutingConfig{FailoverConfig: &v1alpha1.FailoverConfig{
					Primary: &v1alpha1.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &v1alpha1.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
			},
			input: &eventbridge.UpdateEndpointInput{
				EventBuses: []svcsdktypes.EndpointEventBus{
					{EventBusArn: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusArn: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &svcsdktypes.RoutingConfig{FailoverConfig: &svcsdktypes.FailoverConfig{
					Primary: &svcsdktypes.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &svcsdktypes.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
			},
			delta: &ackcompare.Delta{
				Differences: []*ackcompare.Difference{
					{
						Path: ackcompare.NewPath("Spec.ReplicationConfig"),
						A:    &v1alpha1.ReplicationConfig{State: aws.String("ENABLED")},
						B:    nil,
					},
				},
			},
			wantInput: &eventbridge.UpdateEndpointInput{
				EventBuses: []svcsdktypes.EndpointEventBus{
					{EventBusArn: aws.String("arn:aws:events:us-east-1:123456789012:myApplicationBus")},
					{EventBusArn: aws.String("arn:aws:events:us-east-2:123456789012:myApplicationBus")},
				},
				Name: aws.String("endpointspec"),
				RoutingConfig: &svcsdktypes.RoutingConfig{FailoverConfig: &svcsdktypes.FailoverConfig{
					Primary: &svcsdktypes.Primary{
						HealthCheck: aws.String("arn:aws:route53:::healthcheck/1dc6d4f8-5ec8-4089-8b2d-692eef46316b"),
					},
					Secondary: &svcsdktypes.Secondary{
						Route: aws.String("eu-central-1"),
					},
				}},
				ReplicationConfig: &svcsdktypes.ReplicationConfig{State: svcsdktypes.ReplicationStateEnabled},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unsetRemovedSpecFields(tt.delta, tt.spec, tt.input)
			normalizeNilStrings(tt.input)
			normalizeNilStrings(tt.wantInput)

			if diff := cmp.Diff(tt.wantInput, tt.input,
				cmpopts.IgnoreUnexported(
					svcsdktypes.EndpointEventBus{},
					svcsdktypes.ReplicationConfig{},
					svcsdktypes.RoutingConfig{},
					svcsdktypes.FailoverConfig{},
					svcsdktypes.Primary{},
					svcsdktypes.Secondary{},
				),
				cmp.AllowUnexported(eventbridge.UpdateEndpointInput{}),
			); diff != "" {
				t.Errorf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_customPreCompare(t *testing.T) {
	aResource := &resource{
		ko: &v1alpha1.Endpoint{
			Spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:bus:1")},
					{EventBusARN: aws.String("arn:bus:2")},
				},
				Name:              aws.String("bus"),
				ReplicationConfig: &v1alpha1.ReplicationConfig{State: aws.String("ENABLED")},
				RoleARN:           aws.String("arn:role:1"),
				RoutingConfig: &v1alpha1.RoutingConfig{
					FailoverConfig: &v1alpha1.FailoverConfig{
						Primary:   &v1alpha1.Primary{HealthCheck: aws.String("arn:healthcheck:1")},
						Secondary: &v1alpha1.Secondary{Route: aws.String("eu-central-1")},
					},
				},
			},
		},
	}
	type args struct {
		a func() *resource
		b func() *resource
	}
	tests := []struct {
		name            string
		args            args
		wantDifferences []string
	}{
		{
			name: "equal resources",
			args: args{
				a: func() *resource {
					return aResource
				},
				b: func() *resource {
					return aResource
				},
			},
			wantDifferences: nil,
		},
		{
			name: "equal resources with nil and empty description",
			args: args{
				a: func() *resource {
					return aResource
				},
				b: func() *resource {
					b := aResource.ko.DeepCopy()
					b.Spec.Description = aws.String("")
					return &resource{ko: b}
				},
			},
			wantDifferences: nil,
		},
		{
			name: "equal resources with nil and empty roleARN",
			args: args{
				a: func() *resource {
					a := aResource.ko.DeepCopy()
					a.Spec.RoleARN = nil
					return &resource{ko: a}
				},
				b: func() *resource {
					b := aResource.ko.DeepCopy()
					b.Spec.RoleARN = aws.String("")
					return &resource{ko: b}
				},
			},
			wantDifferences: nil,
		},
		{
			name: "equal resources with event buses in different order",
			args: args{
				a: func() *resource {
					return aResource
				},
				b: func() *resource {
					b := aResource.ko.DeepCopy()
					b.Spec.EventBuses = []*v1alpha1.EndpointEventBus{
						{EventBusARN: aws.String("arn:bus:2")},
						{EventBusARN: aws.String("arn:bus:1")},
					}
					return &resource{ko: b}
				},
			},
			wantDifferences: nil,
		},
		{
			name: "equal resources with ENABLED and nil replication config setting",
			args: args{
				a: func() *resource {
					// we need to manipulate desired because latest always returns a setting
					a := aResource.ko.DeepCopy()
					a.Spec.ReplicationConfig.State = nil
					return &resource{ko: a}
				},
				b: func() *resource {
					return aResource
				},
			},
			wantDifferences: nil,
		},
		{
			name: "different roleARN",
			args: args{
				a: func() *resource {
					return aResource
				},
				b: func() *resource {
					b := aResource.ko.DeepCopy()
					b.Spec.RoleARN = aws.String("arn:role:2")
					return &resource{ko: b}
				},
			},
			wantDifferences: []string{"Spec.RoleARN"},
		},
		{
			name: "different event buses",
			args: args{
				a: func() *resource {
					return aResource
				},
				b: func() *resource {
					b := aResource.ko.DeepCopy()
					b.Spec.EventBuses = []*v1alpha1.EndpointEventBus{
						{EventBusARN: aws.String("arn:bus:2")},
						{EventBusARN: aws.String("arn:bus:3")},
					}
					return &resource{ko: b}
				},
			},
			wantDifferences: []string{"Spec.EventBuses"},
		},
		{
			name: "different replication configs",
			args: args{
				a: func() *resource {
					return aResource
				},
				b: func() *resource {
					b := aResource.ko.DeepCopy()
					b.Spec.ReplicationConfig.State = aws.String("DISABLED")
					return &resource{ko: b}
				},
			},
			wantDifferences: []string{"Spec.ReplicationConfig"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ackcompare.NewDelta()
			customPreCompare(d, tt.args.a(), tt.args.b())

			if tt.wantDifferences != nil {
				for _, path := range tt.wantDifferences {
					if !d.DifferentAt(path) {
						t.Errorf("expected difference at %q", path)
					}
				}
			} else {
				assert.DeepEqual(t, d.Differences, []*ackcompare.Difference{})
			}
		})
	}
}
func normalizeNilStrings(input *eventbridge.UpdateEndpointInput) *eventbridge.UpdateEndpointInput {
	if input.Description == nil {
		emptyStr := ""
		input.Description = &emptyStr
	}
	return input
}

// Test_customPreCompare_eventBusesDeltaPayload asserts the Spec.EventBuses
// delta records the event bus values rather than some other field's.
//
// Test_customPreCompare above only asserts the delta *path*, which is why this
// went unnoticed: the entry was raised on the right condition but carried the
// ReplicationConfig values. The delta log is the evidence used to diagnose a
// reference-form mismatch on eventBusARN, so a wrong payload there sends a
// reader chasing the wrong field.
func Test_customPreCompare_eventBusesDeltaPayload(t *testing.T) {
	desired := &resource{
		ko: &v1alpha1.Endpoint{
			Spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:bus:1")},
					{EventBusARN: aws.String("arn:bus:2")},
				},
				Name:              aws.String("bus"),
				ReplicationConfig: &v1alpha1.ReplicationConfig{State: aws.String("ENABLED")},
			},
		},
	}
	latest := &resource{
		ko: &v1alpha1.Endpoint{
			Spec: v1alpha1.EndpointSpec{
				EventBuses: []*v1alpha1.EndpointEventBus{
					{EventBusARN: aws.String("arn:bus:3")},
					{EventBusARN: aws.String("arn:bus:4")},
				},
				Name:              aws.String("bus"),
				ReplicationConfig: &v1alpha1.ReplicationConfig{State: aws.String("ENABLED")},
			},
		},
	}

	d := ackcompare.NewDelta()
	customPreCompare(d, desired, latest)

	assert.Assert(t, d.DifferentAt("Spec.EventBuses"))

	var found *ackcompare.Difference
	for _, diff := range d.Differences {
		if diff.Path.Contains("Spec.EventBuses") {
			found = diff
			break
		}
	}
	assert.Assert(t, found != nil, "expected a Spec.EventBuses difference")

	a, ok := found.A.([]*v1alpha1.EndpointEventBus)
	assert.Assert(t, ok, "Spec.EventBuses delta A must carry the event buses, got %T", found.A)
	b, ok := found.B.([]*v1alpha1.EndpointEventBus)
	assert.Assert(t, ok, "Spec.EventBuses delta B must carry the event buses, got %T", found.B)

	assert.Equal(t, len(a), 2)
	assert.Equal(t, len(b), 2)
	assert.Equal(t, *a[0].EventBusARN, "arn:bus:1")
	assert.Equal(t, *b[0].EventBusARN, "arn:bus:3")
}

// Test_equalEventBusConfigs_nilSafe covers an entry whose ARN is unset. That is
// reachable now that eventBusARN can be supplied as an eventBusRef companion
// instead: the concrete field stays nil until the reference resolves, and delta
// computation runs before the spec validation that rejects a missing ARN. An
// unguarded dereference here would panic the controller.
func Test_equalEventBusConfigs_nilSafe(t *testing.T) {
	withARN := func(s string) *v1alpha1.EndpointEventBus {
		return &v1alpha1.EndpointEventBus{EventBusARN: aws.String(s)}
	}

	tests := []struct {
		name  string
		a     []*v1alpha1.EndpointEventBus
		b     []*v1alpha1.EndpointEventBus
		equal bool
	}{
		{
			name:  "both entries unresolved",
			a:     []*v1alpha1.EndpointEventBus{{}, {}},
			b:     []*v1alpha1.EndpointEventBus{{}, {}},
			equal: true,
		}, {
			name:  "one side unresolved",
			a:     []*v1alpha1.EndpointEventBus{{}, {}},
			b:     []*v1alpha1.EndpointEventBus{withARN("arn:bus:1"), withARN("arn:bus:2")},
			equal: false,
		}, {
			name:  "nil entry in the slice",
			a:     []*v1alpha1.EndpointEventBus{nil, withARN("arn:bus:1")},
			b:     []*v1alpha1.EndpointEventBus{nil, withARN("arn:bus:1")},
			equal: true,
		}, {
			name: "ref-only entry compares equal to the same ref-only entry",
			a: []*v1alpha1.EndpointEventBus{
				{EventBusRef: &ackv1alpha1.AWSResourceReferenceWrapper{
					From: &ackv1alpha1.AWSResourceReference{Name: aws.String("my-bus")},
				}},
				withARN("arn:bus:1"),
			},
			b: []*v1alpha1.EndpointEventBus{
				{EventBusRef: &ackv1alpha1.AWSResourceReferenceWrapper{
					From: &ackv1alpha1.AWSResourceReference{Name: aws.String("my-bus")},
				}},
				withARN("arn:bus:1"),
			},
			equal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Must not panic.
			assert.Equal(t, equalEventBusConfigs(tt.a, tt.b), tt.equal)
		})
	}
}
