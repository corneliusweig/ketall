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

package filter

import (
	"testing"
	"time"

	"github.com/flanksource/ketall/util"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type FakeV1Obj struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (*FakeV1Obj) DeepCopyObject() runtime.Object {
	panic("not supported")
}

func TestFilterByPredicate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		objects       runtime.Object
		givenMaxAge   string
		expectedNames []string
	}{
		{
			name: "two objects, one too old",
			objects: util.ToV1List([]runtime.Object{
				newFakeObj("o1", now),
				newFakeObj("o2", now.Add(-120*time.Second)),
			}),
			givenMaxAge:   "119s",
			expectedNames: []string{"o1"},
		},
		{
			name: "two objects, both too old",
			objects: util.ToV1List([]runtime.Object{
				newFakeObj("o1", now.Add(-60*time.Second)),
				newFakeObj("o2", now.Add(-120*time.Second)),
			}),
			givenMaxAge: "10s",
		},
		{
			name: "two objects, both match",
			objects: util.ToV1List([]runtime.Object{
				newFakeObj("o1", now.Add(-60*time.Second)),
				newFakeObj("o2", now.Add(-120*time.Second)),
			}),
			givenMaxAge:   "121s",
			expectedNames: []string{"o1", "o2"},
		},
		{
			name: "two objects, without duration",
			objects: util.ToV1List([]runtime.Object{
				newFakeObj("o1", now.Add(-60*time.Second)),
				newFakeObj("o2", now.Add(-120*time.Second)),
			}),
			givenMaxAge:   "",
			expectedNames: []string{"o1", "o2"},
		},
		{
			name: "two objects and empty lists, without duration",
			objects: util.ToV1List([]runtime.Object{
				util.ToV1List([]runtime.Object{}),
				newFakeObj("o1", now.Add(-60*time.Second)),
				util.ToV1List([]runtime.Object{}),
				newFakeObj("o2", now.Add(-120*time.Second)),
				util.ToV1List([]runtime.Object{}),
			}),
			givenMaxAge:   "",
			expectedNames: []string{"o1", "o2"},
		},
		{
			name: "two objects and empty lists, with duration",
			objects: util.ToV1List([]runtime.Object{
				util.ToV1List([]runtime.Object{}),
				newFakeObj("o1", now.Add(-60*time.Second)),
				util.ToV1List([]runtime.Object{}),
				newFakeObj("o2", now.Add(-120*time.Second)),
				util.ToV1List([]runtime.Object{}),
			}),
			givenMaxAge:   "90s",
			expectedNames: []string{"o1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filtered := ApplyFilter(test.objects, test.givenMaxAge)

			actualObjs, _ := meta.ExtractList(filtered)
			assert.Equal(t, test.expectedNames, toNames(actualObjs))
		})
	}
}

func newFakeObj(name string, age time.Time) *FakeV1Obj {
	o := &FakeV1Obj{}
	o.Name = name
	o.CreationTimestamp = metav1.Time{Time: age}
	return o
}

func toNames(in []runtime.Object) (names []string) {
	for _, o := range in {
		o, _ := meta.Accessor(o)
		names = append(names, o.GetName())
	}
	return
}

func TestParseHumanDuration(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  int64
		shouldErr bool
	}{
		{name: "one second", input: "1s", expected: 1},
		{name: "seconds with large value", input: "120s", expected: 120},
		{name: "one minute", input: "1m", expected: 60},
		{name: "one hour", input: "1h", expected: (60 * 60)},
		{name: "one day", input: "1d", expected: (24 * 60 * 60)},
		{name: "one year", input: "1y", expected: (365 * 24 * 60 * 60)},
		{name: "second and minute", input: "2m119s", expected: (2*60 + 119)},
		{name: "minute and hour", input: "3h21m", expected: (3*60*60 + 21*60)},
		{name: "hour and day", input: "4d7h", expected: (4*24*60*60 + 7*60*60)},
		{name: "day and year", input: "1y364d", expected: (365*24*60*60 + 364*24*60*60)},
		{name: "complex time", input: "1y1d1h1m1s", expected: 31626061},
		{name: "unknown unit", input: "7k", shouldErr: true},
		{name: "no value", input: "d", shouldErr: true},
		{name: "no value, several groups I", input: "2ys", shouldErr: true},
		{name: "no value, several groups II", input: "y2s", shouldErr: true},
		{name: "minute and seconds, swapped", input: "2s1m", shouldErr: true},
		{name: "same unit repeated", input: "4m2m", shouldErr: true},
		{name: "hours and minute, swapped", input: "2m1h", shouldErr: true},
		{name: "days and hours, swapped", input: "2h1d", shouldErr: true},
		{name: "years and days, swapped", input: "2d1y", shouldErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseHumanDuration(test.input)
			if test.shouldErr {
				assert.Error(t, err, test.input)
			} else {
				assert.Equal(t, time.Duration(test.expected*int64(time.Second)), actual)
			}
		})
	}
}
