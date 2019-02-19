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
	"github.com/corneliusweig/ketall/pkg/constants"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"sort"
	"strings"
)

// groupResource contains the APIGroup and APIResource
type groupResource struct {
	APIGroup    string
	APIResource metav1.APIResource
}

// TODO rework client, so that it does not fail without cluster admin rights
func GetAllServerResources(flags *genericclioptions.ConfigFlags) (runtime.Object, error) {
	useCache := viper.GetBool(constants.FlagUseCache)
	scope := viper.GetString(constants.FlagScope)

	grs, err := FetchAvailableGroupResources(useCache, scope, flags)
	if err != nil {
		return nil, errors.Wrap(err, "fetch available group resources")
	}

	resNames := extractRelevantResourceNames(grs, viper.GetStringSlice(constants.FlagExclude))

	request := resource.NewBuilder(flags).
		Unstructured().
		SelectAllParam(true).
		ResourceTypes(resNames...).
		Latest()

	if ns := viper.GetString(constants.FlagNamespace); ns != "" {
		request.NamespaceParam(ns)
	} else {
		request.AllNamespaces(true)
	}

	response := request.Do()

	if infos, err := response.Infos(); err != nil {
		return nil, errors.Wrap(err, "request resources")
	} else if len(infos) == 0 {
		return nil, fmt.Errorf("No resources found")
	}

	return response.Object()
}

func FetchAvailableGroupResources(cache bool, scope string, flags *genericclioptions.ConfigFlags) ([]groupResource, error) {
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

	grs := []groupResource{}
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

func extractRelevantResourceNames(grs []groupResource, exclusions []string) []string {
	sort.Stable(sortableGroupResource(grs))
	forbidden := sets.NewString(exclusions...)

	result := []string{}
	for _, r := range grs {
		name := r.fullName()
		resourceIds := r.APIResource.ShortNames
		resourceIds = append(resourceIds, r.APIResource.Name)
		if forbidden.HasAny(resourceIds...) {
			logrus.Debugf("Excluding %s", name)
			continue
		}
		result = append(result, name)
	}

	logrus.Debugf("Resources to fetch: %s", result)
	return result
}

func getResourceScope(scope string) (skipCluster, skipNamespace bool, err error) {
	switch scope {
	case "":
		skipCluster = false
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

func (g groupResource) fullName() string {
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
