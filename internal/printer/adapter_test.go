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

package printer

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockObject struct {
	gvk     schema.GroupVersionKind
	content string
}

func (o *mockObject) GetObjectKind() schema.ObjectKind {
	return o
}

func (o *mockObject) String() string {
	return fmt.Sprintf("versionKind: %s/%s  content: %s\n", o.gvk.Version, o.gvk.Kind, o.content)
}

func (o *mockObject) DeepCopyObject() runtime.Object {
	clone := mockObject{
		gvk:     schema.GroupVersionKind{Group: o.gvk.Group, Version: o.gvk.Version, Kind: o.gvk.Kind},
		content: o.content,
	}
	return &clone
}

func (o *mockObject) GroupVersionKind() schema.GroupVersionKind {
	return o.gvk
}

func (o *mockObject) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	o.gvk = gvk
}

type mockList struct {
	mockObject
	Items []mockObject
}

type mockPrinter struct{}

func (*mockPrinter) PrintObj(r runtime.Object, w io.Writer) error {
	var message string
	switch r := r.(type) {
	case *mockList:
		message = r.String()
	case *mockObject:
		message = r.String()
	}

	_, err := io.WriteString(w, message)
	return err
}

func TestListAdapterPrinter_PrintObj(t *testing.T) {
	delegate := mockPrinter{}
	buffer := &bytes.Buffer{}
	printer := NewListAdapterPrinter(&delegate)

	o := &mockObject{content: "mock object"}
	err := printer.PrintObj(o, buffer)
	assert.NoError(t, err)
	assert.Equal(t, buffer.String(), "versionKind: /  content: mock object\n")

	buffer.Truncate(0)
	l := &mockList{mockObject: mockObject{content: "mock list"}}
	err = printer.PrintObj(l, buffer)
	assert.NoError(t, err)
	assert.Equal(t, buffer.String(), "versionKind: v1/List  content: mock list\n")
}

func TestFlattenListAdapter_PrintObj(t *testing.T) {
	delegate := mockPrinter{}
	buffer := &bytes.Buffer{}
	printer := NewFlattenListAdapterPrinter(&delegate)

	gvk := schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Test"}
	list := &mockList{
		Items: []mockObject{
			{gvk: gvk, content: "object 1"},
			{gvk: gvk, content: "object 2"},
			{gvk: gvk, content: "object 3"},
		},
	}

	err := printer.PrintObj(list, buffer)
	assert.NoError(t, err)
	assert.Equal(t, buffer.String(), `versionKind: v1/Test  content: object 1
versionKind: v1/Test  content: object 2
versionKind: v1/Test  content: object 3
`)
}
