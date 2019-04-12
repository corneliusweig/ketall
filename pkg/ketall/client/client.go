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
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/corneliusweig/ketall/pkg/ketall/constants"
	"github.com/corneliusweig/ketall/pkg/ketall/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
)

// groupResource contains the APIGroup and APIResource
type groupResource struct {
	APIGroup    string
	APIResource metav1.APIResource
}

func GetAllServerResources(flags *genericclioptions.ConfigFlags) (runtime.Object, error) {
	useCache := viper.GetBool(constants.FlagUseCache)
	scope := viper.GetString(constants.FlagScope)

	grs, err := fetchAvailableGroupResources(useCache, scope, flags)
	if err != nil {
		return nil, errors.Wrap(err, "fetch available group resources")
	}

	resources := extractRelevantResources(grs, viper.GetStringSlice(constants.FlagExclude))

	start := time.Now()
	response, err := fetchResourcesBulk(flags, resources...)
	logrus.Debugf("Initial fetchResourcesBulk done (%s)", duration.HumanDuration(time.Since(start)))
	if err == nil {
		return response, nil
	}

	return fetchResourcesIncremental(flags, resources...)
}

func fetchAvailableGroupResources(cache bool, scope string, flags *genericclioptions.ConfigFlags) ([]groupResource, error) {
	client, err := flags.ToDiscoveryClient()
	if err != nil {
		return nil, errors.Wrap(err, "discovery client")
	}

	if !cache {
		client.Invalidate()
	}

	skipCluster, skipNamespace, err := getResourceScope(scope)
	if err != nil {
		return nil, err
	}

	resources, err := client.ServerPreferredResources()
	if err != nil {
		return nil, errors.Wrap(err, "get preferred resources")
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

			if (r.Namespaced && skipNamespace) || (!r.Namespaced && skipCluster) {
				continue
			}

			// filter to resources that can be listed
			if len(r.Verbs) > 0 && !sets.NewString(r.Verbs...).HasAny("list", "get") {
				continue
			}

			grs = append(grs, groupResource{
				APIGroup:    gv.Group,
				APIResource: r,
			})
		}
	}

	return grs, nil
}

func extractRelevantResources(grs []groupResource, exclusions []string) []groupResource {
	sort.Stable(sortableGroupResource(grs))
	forbidden := sets.NewString(exclusions...)

	var result []groupResource
	for _, r := range grs {
		name := r.fullName()
		resourceIds := r.APIResource.ShortNames
		resourceIds = append(resourceIds, r.APIResource.Name)
		if forbidden.HasAny(resourceIds...) {
			logrus.Debugf("Excluding %s", name)
			continue
		}
		result = append(result, r)
	}

	return result
}

// Fetches all objects in bulk. This is much faster than incrementally but may fail due to missing rights
func fetchResourcesBulk(flags resource.RESTClientGetter, resourceTypes ...groupResource) (runtime.Object, error) {
	resourceNames := ToResourceTypes(resourceTypes)
	logrus.Debugf("Resources to fetch: %s", resourceNames)

	request := resource.NewBuilder(flags).
		Unstructured().
		SelectAllParam(true).
		ResourceTypes(resourceNames...).
		Latest()
	if ns := viper.GetString(constants.FlagNamespace); ns != "" {
		request.NamespaceParam(ns)
	} else {
		request.AllNamespaces(true)
	}

	return request.Do().Object()
}

// Fetches all objects of the given resources one-by-one. This can be used as a fallback when fetchResourcesBulk fails.
func fetchResourcesIncremental(flags resource.RESTClientGetter, rs ...groupResource) (runtime.Object, error) {
	logrus.Debug("Fetch resources incrementally")
	group := sync.WaitGroup{}

	objsChan := make(chan runtime.Object)
	for _, r := range rs {
		r := r
		group.Add(1)
		go func(sendObj chan<- runtime.Object) {
			defer group.Done()
			if o, e := fetchResourcesBulk(flags, r); e != nil {
				logrus.Warnf("Cannot fetch: %s", e)
			} else {
				sendObj <- o
			}
		}(objsChan)
	}

	go func() {
		start := time.Now()
		group.Wait()
		close(objsChan)
		logrus.Debugf("Requests done (elapsed %s)", duration.HumanDuration(time.Since(start)))
	}()

	var objs []runtime.Object
	for o := range objsChan {
		objs = append(objs, o)
	}

	if len(objs) == 0 {
		return nil, fmt.Errorf("not authorized to list any resources, try to narrow the scope with --namespace")
	}

	return util.ToV1List(objs), nil
}

func getResourceScope(scope string) (skipCluster, skipNamespace bool, err error) {
	switch scope {
	case "":
		skipCluster = viper.GetString(constants.FlagNamespace) != ""
		skipNamespace = false
	case "namespace":
		skipCluster = true
		skipNamespace = false
	case "cluster":
		skipCluster = false
		skipNamespace = true
	default:
		err = fmt.Errorf("%s is not a valid resource scope (must be one of 'cluster' or 'namespace')", scope)
	}
	return
}

// Extracts the full name including APIGroup, e.g. 'deployment.apps'
func (g groupResource) fullName() string {
	if g.APIGroup == "" {
		return g.APIResource.Name
	}
	return fmt.Sprintf("%s.%s", g.APIResource.Name, g.APIGroup)
}

type sortableGroupResource []groupResource

func ToResourceTypes(in []groupResource) []string {
	var result []string
	for _, r := range in {
		result = append(result, r.fullName())
	}
	return result
}

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
