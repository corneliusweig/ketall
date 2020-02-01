/*
Copyright 2019 Cornelius Weig

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractRelevantResourceNames(t *testing.T) {
	var tests = []struct {
		testName  string
		resources []v1.APIResource
		groups    []string
		exclude   []string
		expected  []string
	}{
		{
			testName:  "a single resource",
			resources: []v1.APIResource{{Name: "foo"}},
			groups:    []string{"group"},
			expected:  []string{"foo.group"},
		},
		{
			testName:  "two resources, without group",
			resources: []v1.APIResource{{Name: "foo"}, {Name: "bar"}},
			groups:    []string{"group", ""},
			expected:  []string{"bar", "foo.group"},
		},
		{
			testName:  "two resources, same group",
			resources: []v1.APIResource{{Name: "foo"}, {Name: "bar"}},
			groups:    []string{"group", "group"},
			expected:  []string{"bar.group", "foo.group"},
		},
		{
			testName:  "two filtered by Name",
			resources: []v1.APIResource{{Name: "foo"}, {Name: "bar"}},
			groups:    []string{"group", "puorg"},
			exclude:   []string{"bar"},
			expected:  []string{"foo.group"},
		},
		{
			testName:  "two filtered by ShortName",
			resources: []v1.APIResource{{Name: "foo", ShortNames: []string{"baz"}}, {Name: "bar"}},
			groups:    []string{"group", "puorg"},
			exclude:   []string{"baz"},
			expected:  []string{"bar.puorg"},
		},
		{
			testName:  "two filtered by fully-qualified resource name",
			resources: []v1.APIResource{{Name: "foo"}, {Name: "bar"}},
			groups:    []string{"group", "puorg"},
			exclude:   []string{"foo.group"},
			expected:  []string{"bar.puorg"},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			var grs []groupResource
			for i := range test.resources {
				grs = append(grs, groupResource{
					APIGroup:    test.groups[i],
					APIResource: test.resources[i],
				})
			}

			names := extractRelevantResources(grs, test.exclude)
			assert.Equal(t, test.expected, ToResourceTypes(names))
		})
	}
}
