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
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/ketall/options"
	"github.com/flanksource/ketall/util"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/klog/v2"
)

var errEmpty = errors.New("no resources found")

// groupResource contains the APIGroup and APIResource
type groupResource struct {
	APIGroup    string
	APIResource metav1.APIResource
}

func GetAllServerResources(ketalloptions *options.KetallOptions) (runtime.Object, error) {

	grs, err := groupResources(ketalloptions)
	if err != nil {
		return nil, errors.Wrap(err, "fetch available group resources")
	}

	start := time.Now()
	response, err := fetchResourcesBulk(ketalloptions.Namespace, ketalloptions.Selector, ketalloptions.FieldSelector, ketalloptions.GenericCliFlags, grs...)
	klog.V(2).Infof("Initial fetchResourcesBulk done (%s)", duration.HumanDuration(time.Since(start)))
	if err == nil {
		return response, nil
	}

	return fetchResourcesIncremental(context.TODO(), ketalloptions, grs...)
}

func getExclusions(exclusions []string, selector, fieldSelector string) []string {
	// This is a workaround for a k8s bug where componentstatus is reported even though the selector does not apply
	if selector != "" || fieldSelector != "" {
		exclusions = append(exclusions, "componentstatuses")
	}

	return exclusions
}

func groupResources(ketalloptions *options.KetallOptions) ([]groupResource, error) {
	client, err := ketalloptions.GenericCliFlags.ToDiscoveryClient()
	if err != nil {
		return nil, errors.Wrap(err, "discovery client")
	}

	if !ketalloptions.UseCache {
		client.Invalidate()
	}

	scopeCluster, scopeNamespace, err := getResourceScope(ketalloptions.Scope, ketalloptions.Namespace)
	if err != nil {
		return nil, err
	}

	resources, err := client.ServerPreferredResources()
	if err != nil {
		if resources == nil || !ketalloptions.AllowIncomplete {
			return nil, errors.Wrap(err, "get preferred resources")
		}
		klog.Warningf("Could not fetch complete list of API resources, results will be incomplete: %s", err)
	}

	var grs []groupResource
	for _, list := range resources {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, r := range list.APIResources {
			if len(r.Verbs) == 0 {
				continue
			}

			if !((r.Namespaced && scopeNamespace) || (!r.Namespaced && scopeCluster)) {
				// The resource scope was disabled.
				continue
			}

			// filter to resources that can be listed
			if !sets.NewString(r.Verbs...).HasAny("list", "get") {
				continue
			}

			grs = append(grs, groupResource{
				APIGroup:    gv.Group,
				APIResource: r,
			})
		}
	}

	sort.Stable(sortableGroupResource(grs))
	blocked := sets.NewString(getExclusions(ketalloptions.Exclusions, ketalloptions.Selector, ketalloptions.FieldSelector)...)

	ret := grs[:0]
	for _, r := range grs {
		name := r.String()
		resourceIds := r.APIResource.ShortNames
		resourceIds = append(resourceIds, r.APIResource.Name)
		resourceIds = append(resourceIds, r.APIResource.Kind)
		resourceIds = append(resourceIds, name)
		if blocked.HasAny(resourceIds...) {
			klog.V(2).Infof("Excluding %s", name)
			continue
		}
		ret = append(ret, r)
	}
	return ret, nil
}

// Fetches all objects in bulk. This is much faster than incrementally but may fail due to missing rights
func fetchResourcesBulk(ns, selector, fieldSelector string, flags resource.RESTClientGetter, grs ...groupResource) (runtime.Object, error) {
	var resources []string
	for _, gr := range grs {
		resources = append(resources, gr.String())
	}
	klog.V(2).Infof("Resources to fetch: %s", resources)
	request := resource.NewBuilder(flags).
		Unstructured().
		ResourceTypes(resources...).
		NamespaceParam(ns).DefaultNamespace().AllNamespaces(ns == "").
		LabelSelectorParam(selector).FieldSelectorParam(fieldSelector).SelectAllParam(selector == "" && fieldSelector == "").
		Flatten().
		Latest()

	return request.Do().Object()
}

// Fetches all objects of the given resources one-by-one. This can be used as a fallback when fetchResourcesBulk fails.
func fetchResourcesIncremental(ctx context.Context, ketalloptions *options.KetallOptions, grs ...groupResource) (runtime.Object, error) {
	// TODO(corneliusweig): this needs to properly pass ctx around
	klog.V(2).Info("Fetch resources incrementally")
	start := time.Now()

	sem := semaphore.NewWeighted(ketalloptions.MaxInflight) // restrict parallelism to 64 inflight requests

	var mu sync.Mutex // mu guards ret
	var ret []runtime.Object

	var wg sync.WaitGroup
	for _, gr := range grs {
		wg.Add(1)
		go func(gr groupResource) {
			defer wg.Done()
			if err := sem.Acquire(ctx, 1); err != nil {
				return // context cancelled
			}
			defer sem.Release(1)
			obj, err := fetchResourcesBulk(ketalloptions.Namespace, ketalloptions.Selector, ketalloptions.FieldSelector, ketalloptions.GenericCliFlags, gr)
			if err != nil {
				klog.Warningf("Cannot fetch: %v", err)
				return
			}
			mu.Lock()
			ret = append(ret, obj)
			mu.Unlock()
		}(gr)
	}
	wg.Wait()
	klog.V(2).Infof("Requests done (elapsed %s)", duration.HumanDuration(time.Since(start)))

	if len(ret) == 0 {
		klog.Warningf("No resources found, are you authorized? Try to narrow the scope with --namespace.")
		return nil, errEmpty
	}

	return util.ToV1List(ret), nil
}

func getResourceScope(scope, ns string) (cluster, namespace bool, err error) {
	switch scope {
	case "":
		cluster = ns == ""
		namespace = true
	case "namespace":
		cluster = false
		namespace = true
	case "cluster":
		cluster = true
		namespace = false
	default:
		err = fmt.Errorf("%s is not a valid resource scope (must be one of 'cluster' or 'namespace')", scope)
	}
	return
}

// String returns the canonical full name of the groupResource.
func (g groupResource) String() string {
	if g.APIGroup == "" {
		return g.APIResource.Name
	}
	return fmt.Sprintf("%s.%s", g.APIResource.Name, g.APIGroup)
}

type sortableGroupResource []groupResource

func (s sortableGroupResource) Len() int      { return len(s) }
func (s sortableGroupResource) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortableGroupResource) Less(i, j int) bool {
	ret := strings.Compare(s[i].APIGroup, s[j].APIGroup)
	if ret > 0 {
		return false
	} else if ret == 0 {
		return strings.Compare(s[i].APIResource.Name, s[j].APIResource.Name) < 0
	}
	return true
}
