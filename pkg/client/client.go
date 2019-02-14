package client

import (
	"fmt"
	"github.com/corneliusweig/ketall/pkg/options"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
	"k8s.io/cli-runtime/pkg/genericclioptions/resource"
	"os"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

type GAClient struct {
	resource.RESTClient
}

// groupResource contains the APIGroup and APIResource
type groupResource struct {
	APIGroup    string
	APIResource metav1.APIResource
}

func PrintAllServerResources(gaOptions *options.CmdOptions) error {
	flags := gaOptions.GenericCliFlags

	resNames, err := FetchAvailableResourceNames(flags)
	if err != nil {
		return errors.Wrap(err, "fetch available resources")
	}

	r := resource.NewBuilder(flags).
		Unstructured().
		AllNamespaces(true).
		SelectAllParam(true).
		ResourceTypes(resNames...).
		Latest().
		Do()

	if infos, err := r.Infos(); err != nil {
		return errors.Wrap(err, "request resources")
	} else if len(infos) == 0 {
		logrus.Warn("No resources found")
		return nil
	}

	allObjects, _ := meta.ExtractList(must(r.Object()))

	if err := writeTable(allObjects); err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func must(o runtime.Object, e error) runtime.Object {
	if e != nil {
		logrus.Fatal(e)
	}
	return o
}

func writeNames(objs []runtime.Object, flags *genericclioptions.PrintFlags) error {
	printer, err := flags.ToPrinter()
	if err != nil {
		return err
	}

	for _, objToPrint := range objs {
		if err := printer.PrintObj(objToPrint, os.Stdout); err != nil {
			return fmt.Errorf("unable to output the provided object: %v", err)
		}
	}

	return nil
}

func writeTable(objs []runtime.Object) error {
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	defer w.Flush()

	if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", "NAME", "NAMESPACE", "AGE"); err != nil {
		return err
	}

	for _, o := range objs {

		if printers.InternalObjectPreventer.IsForbidden(reflect.Indirect(reflect.ValueOf(o)).Type().PkgPath()) {
			return fmt.Errorf(printers.InternalObjectPrinterErr)
		}

		if o.GetObjectKind().GroupVersionKind().Empty() {
			return fmt.Errorf("missing apiVersion or kind; try GetObjectKind().SetGroupVersionKind() if you know the type")
		}

		if meta.IsListType(o) {
			objs, err := meta.ExtractList(o)
			if err != nil {
				return err
			}
			for _, o := range objs {
				if err := printObj(w, o); err != nil {
					return err
				}
			}

		} else {
			if err := printObj(w, o); err != nil {
				return err
			}
		}
	}

	return nil
}

func printObj(w io.Writer, o runtime.Object) error {
	groupKind := GetObjectGroupKind(o)

	acc, err := meta.Accessor(o)
	if err != nil {
		return err
	}

	name := fullName(acc.GetName(), groupKind)
	timestamp := acc.GetCreationTimestamp()
	if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", name, acc.GetNamespace(), translateTimestampSince(timestamp)); err != nil {
		return err
	}
	return nil
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

func fullName(name string, groupKind schema.GroupKind) string {
	if len(groupKind.Kind) == 0 {
		return name
	}

	if len(groupKind.Group) == 0 {
		return fmt.Sprintf("%s/%s", strings.ToLower(groupKind.Kind), name)
	}

	return fmt.Sprintf("%s.%s/%s", strings.ToLower(groupKind.Kind), groupKind.Group, name)
}

func GetObjectGroupKind(obj runtime.Object) schema.GroupKind {
	if obj == nil {
		return schema.GroupKind{Kind: "<unknown>"}
	}
	groupVersionKind := obj.GetObjectKind().GroupVersionKind()
	if len(groupVersionKind.Kind) > 0 {
		return groupVersionKind.GroupKind()
	}

	if uns, ok := obj.(*unstructured.Unstructured); ok {
		if len(uns.GroupVersionKind().Kind) > 0 {
			return uns.GroupVersionKind().GroupKind()
		}
	}

	return schema.GroupKind{Kind: "<unknown>"}
}

func FetchAvailableResourceNames(flags *genericclioptions.ConfigFlags) ([]string, error) {
	client, err := flags.ToDiscoveryClient()

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
