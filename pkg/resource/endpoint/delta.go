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

package endpoint

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
	customPreCompare(delta, a, b)

	if ackcompare.HasNilDifference(a.ko.Spec.Name, b.ko.Spec.Name) {
		delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
	} else if a.ko.Spec.Name != nil && b.ko.Spec.Name != nil {
		if *a.ko.Spec.Name != *b.ko.Spec.Name {
			delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.RoutingConfig, b.ko.Spec.RoutingConfig) {
		delta.Add("Spec.RoutingConfig", a.ko.Spec.RoutingConfig, b.ko.Spec.RoutingConfig)
	} else if a.ko.Spec.RoutingConfig != nil && b.ko.Spec.RoutingConfig != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.RoutingConfig.FailoverConfig, b.ko.Spec.RoutingConfig.FailoverConfig) {
			delta.Add("Spec.RoutingConfig.FailoverConfig", a.ko.Spec.RoutingConfig.FailoverConfig, b.ko.Spec.RoutingConfig.FailoverConfig)
		} else if a.ko.Spec.RoutingConfig.FailoverConfig != nil && b.ko.Spec.RoutingConfig.FailoverConfig != nil {
			if ackcompare.HasNilDifference(a.ko.Spec.RoutingConfig.FailoverConfig.Primary, b.ko.Spec.RoutingConfig.FailoverConfig.Primary) {
				delta.Add("Spec.RoutingConfig.FailoverConfig.Primary", a.ko.Spec.RoutingConfig.FailoverConfig.Primary, b.ko.Spec.RoutingConfig.FailoverConfig.Primary)
			} else if a.ko.Spec.RoutingConfig.FailoverConfig.Primary != nil && b.ko.Spec.RoutingConfig.FailoverConfig.Primary != nil {
				if ackcompare.HasNilDifference(a.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck, b.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck) {
					delta.Add("Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck", a.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck, b.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck)
				} else if a.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck != nil && b.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck != nil {
					if *a.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck != *b.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck {
						delta.Add("Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck", a.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck, b.ko.Spec.RoutingConfig.FailoverConfig.Primary.HealthCheck)
					}
				}
			}
			if ackcompare.HasNilDifference(a.ko.Spec.RoutingConfig.FailoverConfig.Secondary, b.ko.Spec.RoutingConfig.FailoverConfig.Secondary) {
				delta.Add("Spec.RoutingConfig.FailoverConfig.Secondary", a.ko.Spec.RoutingConfig.FailoverConfig.Secondary, b.ko.Spec.RoutingConfig.FailoverConfig.Secondary)
			} else if a.ko.Spec.RoutingConfig.FailoverConfig.Secondary != nil && b.ko.Spec.RoutingConfig.FailoverConfig.Secondary != nil {
				if ackcompare.HasNilDifference(a.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route, b.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route) {
					delta.Add("Spec.RoutingConfig.FailoverConfig.Secondary.Route", a.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route, b.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route)
				} else if a.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route != nil && b.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route != nil {
					if *a.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route != *b.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route {
						delta.Add("Spec.RoutingConfig.FailoverConfig.Secondary.Route", a.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route, b.ko.Spec.RoutingConfig.FailoverConfig.Secondary.Route)
					}
				}
			}
		}
	}

	return delta
}
