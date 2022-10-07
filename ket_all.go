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

package ketall

import (
	"github.com/flanksource/ketall/client"
	"github.com/flanksource/ketall/filter"
	"github.com/flanksource/ketall/options"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func KetAll(ketallOptions *options.KetallOptions) []*unstructured.Unstructured {
	all, err := client.GetAllServerResources(ketallOptions)
	if err != nil {
		klog.Fatal(err)
	}

	filtered := filter.ApplyFilter(all, ketallOptions.Since)

	items := filtered.(*v1.List).Items
	var unstructuredItems []*unstructured.Unstructured
	for _, item := range items {
		unstrucrureItem := item.Object.(*unstructured.Unstructured)
		unstructuredItems = append(unstructuredItems, unstrucrureItem)
	}
	if filtered == nil {
		return nil
	}
	return unstructuredItems
}
