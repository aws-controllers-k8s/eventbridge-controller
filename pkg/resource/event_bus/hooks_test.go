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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eventbridge"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackmetrics "github.com/aws-controllers-k8s/runtime/pkg/metrics"

	svcapitypes "github.com/aws-controllers-k8s/eventbridge-controller/apis/v1alpha1"
)

func Test_tagsDifference(t *testing.T) {
	tag := func(k, v string) *svcapitypes.Tag {
		return &svcapitypes.Tag{Key: &k, Value: &v}
	}

	tests := []struct {
		name        string
		desired     []*svcapitypes.Tag
		latest      []*svcapitypes.Tag
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
			desired:     []*svcapitypes.Tag{tag("akey", "avalue"), tag("bkey", "bvalue")},
			latest:      []*svcapitypes.Tag{tag("bkey", "bvalue"), tag("akey", "avalue")},
			wantAdded:   map[string]string{},
			wantRemoved: map[string]string{},
		},
		{
			name:        "add tags when latest is empty",
			desired:     []*svcapitypes.Tag{tag("akey", "avalue"), tag("bkey", "bvalue")},
			latest:      nil,
			wantAdded:   map[string]string{"akey": "avalue", "bkey": "bvalue"},
			wantRemoved: map[string]string{},
		},
		{
			name:        "remove tags when desired is empty",
			desired:     nil,
			latest:      []*svcapitypes.Tag{tag("akey", "avalue")},
			wantAdded:   map[string]string{},
			wantRemoved: map[string]string{"akey": "avalue"},
		},
		// GetTagsDifference has upsert semantics: a changed value appears in "added" with
		// the new value and NOT in "removed", because TagResource overwrites existing keys.
		// No UntagResource call is needed.
		{
			name:        "value change is in added and not in removed",
			desired:     []*svcapitypes.Tag{tag("akey", "avalue-new")},
			latest:      []*svcapitypes.Tag{tag("akey", "avalue-old")},
			wantAdded:   map[string]string{"akey": "avalue-new"},
			wantRemoved: map[string]string{},
		},
		{
			name:        "add, remove and value change together",
			desired:     []*svcapitypes.Tag{tag("akey", "avalue"), tag("ckey", "cvalue")},
			latest:      []*svcapitypes.Tag{tag("akey", "avalue-old"), tag("bkey", "bvalue")},
			wantAdded:   map[string]string{"akey": "avalue", "ckey": "cvalue"},
			wantRemoved: map[string]string{"bkey": "bvalue"},
		},
		{
			name: "nil key is skipped without panic",
			desired: []*svcapitypes.Tag{
				{Key: nil, Value: nil},
				tag("akey", "avalue"),
			},
			latest:      []*svcapitypes.Tag{{Key: nil, Value: nil}},
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

// capturedCall and callRecorder are duplicated in pkg/resource/rule/hooks_tags_test.go;
// keep both copies in sync if you change the fields.
type capturedCall struct {
	Target string
	Body   map[string]interface{}
}

type callRecorder struct {
	mu    sync.Mutex
	calls []capturedCall
}

func (r *callRecorder) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		r.mu.Lock()
		r.calls = append(r.calls, capturedCall{
			Target: req.Header.Get("X-Amz-Target"),
			Body:   payload,
		})
		r.mu.Unlock()
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}

func newTestTagRM(t *testing.T) (*resourceManager, *callRecorder) {
	rec := &callRecorder{}
	srv := httptest.NewServer(rec.handler())
	t.Cleanup(srv.Close)
	client := svcsdk.New(svcsdk.Options{
		Region:       "us-west-2",
		BaseEndpoint: aws.String(srv.URL),
		Credentials:  aws.AnonymousCredentials{},
	})
	return &resourceManager{
		metrics: ackmetrics.NewMetrics("eventbridge"),
		sdkapi:  client,
	}, rec
}

func mkEventBus(tags []*svcapitypes.Tag) *resource {
	arn := ackv1alpha1.AWSResourceName("arn:aws:events:us-west-2:123456789012:event-bus/test-bus")
	return &resource{ko: &svcapitypes.EventBus{
		Spec: svcapitypes.EventBusSpec{Tags: tags},
		Status: svcapitypes.EventBusStatus{
			ACKResourceMetadata: &ackv1alpha1.ResourceMetadata{ARN: &arn},
		},
	}}
}

func busTag(k, v string) *svcapitypes.Tag {
	return &svcapitypes.Tag{Key: aws.String(k), Value: aws.String(v)}
}

func Test_syncTags(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name          string
		desired       []*svcapitypes.Tag
		latest        []*svcapitypes.Tag
		wantTargets   []string // X-Amz-Target header values, in order; nil means no calls
		wantUntagKeys []string
		wantTagTags   map[string]string
	}{
		{
			name:        "no-op when tags are identical",
			desired:     []*svcapitypes.Tag{busTag("a", "1"), busTag("b", "2")},
			latest:      []*svcapitypes.Tag{busTag("b", "2"), busTag("a", "1")},
			wantTargets: nil,
		},
		{
			name:        "add only",
			desired:     []*svcapitypes.Tag{busTag("a", "1")},
			latest:      nil,
			wantTargets: []string{"AWSEvents.TagResource"},
			wantTagTags: map[string]string{"a": "1"},
		},
		{
			name:          "remove only",
			desired:       nil,
			latest:        []*svcapitypes.Tag{busTag("a", "1")},
			wantTargets:   []string{"AWSEvents.UntagResource"},
			wantUntagKeys: []string{"a"},
		},
		{
			name:        "value change only calls TagResource not UntagResource",
			desired:     []*svcapitypes.Tag{busTag("a", "new")},
			latest:      []*svcapitypes.Tag{busTag("a", "old")},
			wantTargets: []string{"AWSEvents.TagResource"},
			wantTagTags: map[string]string{"a": "new"},
		},
		{
			name:    "mixed add remove and value change",
			desired: []*svcapitypes.Tag{busTag("a", "new"), busTag("d", "d")},
			latest:  []*svcapitypes.Tag{busTag("a", "old"), busTag("b", "b")},
			// UntagResource for b, then TagResource for a (changed) and d (added)
			wantTargets:   []string{"AWSEvents.UntagResource", "AWSEvents.TagResource"},
			wantUntagKeys: []string{"b"},
			wantTagTags:   map[string]string{"a": "new", "d": "d"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rm, rec := newTestTagRM(t)
			// event_bus.syncTags signature: syncTags(ctx, desired, latest)
			err := rm.syncTags(ctx, mkEventBus(tc.desired), mkEventBus(tc.latest))
			if err != nil {
				t.Fatalf("syncTags() returned error: %v", err)
			}
			rec.mu.Lock()
			calls := rec.calls
			rec.mu.Unlock()

			var got []string
			for _, c := range calls {
				got = append(got, c.Target)
			}
			if !reflect.DeepEqual(tc.wantTargets, got) {
				t.Errorf("API calls: want %v, got %v", tc.wantTargets, got)
			}

			for _, c := range calls {
				switch c.Target {
				case "AWSEvents.UntagResource":
					if tc.wantUntagKeys != nil {
						rawKeys, _ := c.Body["TagKeys"].([]interface{})
						gotKeys := make([]string, len(rawKeys))
						for i, k := range rawKeys {
							gotKeys[i], _ = k.(string)
						}
						if !reflect.DeepEqual(tc.wantUntagKeys, gotKeys) {
							t.Errorf("UntagResource TagKeys: want %v, got %v", tc.wantUntagKeys, gotKeys)
						}
					}
				case "AWSEvents.TagResource":
					if tc.wantTagTags != nil {
						rawTags, _ := c.Body["Tags"].([]interface{})
						gotTags := make(map[string]string, len(rawTags))
						for _, rt := range rawTags {
							m, _ := rt.(map[string]interface{})
							k, _ := m["Key"].(string)
							v, _ := m["Value"].(string)
							gotTags[k] = v
						}
						if !reflect.DeepEqual(tc.wantTagTags, gotTags) {
							t.Errorf("TagResource Tags: want %v, got %v", tc.wantTagTags, gotTags)
						}
					}
				}
			}
		})
	}
}
