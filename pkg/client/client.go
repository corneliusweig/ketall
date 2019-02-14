package client

import (
	"fmt"
	"github.com/corneliusweig/ketall/pkg/options"
	"github.com/pkg/errors"
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
func GetAllServerResources(ketallOptions *options.KetallOptions) (runtime.Object, error) {
	flags := ketallOptions.GenericCliFlags

	resNames, err := FetchAvailableResourceNames(ketallOptions.UseCache, ketallOptions.Scope, flags)
	if err != nil {
		return nil, errors.Wrap(err, "fetch available resources")
	}

	request := resource.NewBuilder(flags).
		Unstructured().
		SelectAllParam(true).
		ResourceTypes(resNames...).
		Latest()

	ns := ketallOptions.GenericCliFlags.Namespace
	if ns != nil && *ns != "" {
		request.NamespaceParam(*ns)
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

func FetchAvailableResourceNames(cache bool, scope string, flags *genericclioptions.ConfigFlags) ([]string, error) {
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

	sort.Stable(sortableGroupResource(grs))
	result := []string{}
	for _, r := range grs {
		result = append(result, r.APIResource.Name)
	}

	return result, nil
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
