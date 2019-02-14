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

func GetAllServerResources(ketallOptions *options.KetallOptions) (runtime.Object, error) {
	flags := ketallOptions.GenericCliFlags

	resNames, err := FetchAvailableResourceNames(flags)
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

	r := request.Do()

	if infos, err := r.Infos(); err != nil {
		return nil, errors.Wrap(err, "request resources")
	} else if len(infos) == 0 {
		return nil, fmt.Errorf("No resources found")
	}

	return r.Object()
}

func FetchAvailableResourceNames(flags *genericclioptions.ConfigFlags) ([]string, error) {
	client, err := flags.ToDiscoveryClient()
	if err != nil {
		return nil, errors.Wrap(err, "discovery client")
	}

	// TODO add option to disable caching
	/*if !o.Cached {
		// Always request fresh data from the server
		discoveryclient.Invalidate()
	}*/

	resources, err := client.ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("ERR: %s", err)
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
			// filter to resources that support the specified verbs
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

func getApiGroups(flags *genericclioptions.ConfigFlags) ([]string, error) {
	client, err := flags.ToDiscoveryClient()
	apiGroupList, err := client.ServerGroups()
	if err != nil {
		return nil, errors.Wrap(err, "getting api groups")
	}
	apiVersions := metav1.ExtractGroupVersions(apiGroupList)

	sort.Strings(apiVersions)
	return apiVersions, nil
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
