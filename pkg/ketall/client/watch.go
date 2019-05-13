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
	"github.com/corneliusweig/ketall/pkg/ketall/color"
	"github.com/google/go-cmp/cmp"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"sync"

	"github.com/corneliusweig/ketall/pkg/ketall/constants"
	"github.com/corneliusweig/ketall/pkg/ketall/printer"
	"github.com/corneliusweig/tabwriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	watchtools "k8s.io/client-go/tools/watch"
)

type parPrinter struct {
	sync.Mutex

	w *tabwriter.Writer
	p printers.ResourcePrinter
}

func (pp *parPrinter) print(e watch.Event) error {
	pp.Lock()
	defer pp.Unlock()
	if err := pp.colorizeEventType(e.Type); err != nil {
		return err
	}
	if err := pp.p.PrintObj(e.Object, pp.w); err != nil {

		return err
	}
	if err := pp.w.Flush(); err != nil {
		return err
	}
	return nil
}

func (pp *parPrinter) colorizeEventType(e watch.EventType) error {
	var c color.Color
	switch e {
	case watch.Added:
		c = color.Green
	case watch.Modified:
		c = color.Blue
	case watch.Deleted:
		c = color.Red
	case watch.Error:
		c = color.Purple
	}
	if _, err := c.Fprint(pp.w, string(e)); err != nil {
		return err
	}
	if _, err := pp.w.Write([]byte("\t")); err != nil {
		return err
	}
	return nil
}
/*func (pp *parPrinter) colorizeEventType(e watch.EventType) error {
	var c color.Color
	var text string
	switch e {
	case watch.Added:
		c = color.Green
		text = " +"
	case watch.Modified:
		c = color.Blue
		text = " ○"
	case watch.Deleted:
		c = color.Red
		text = " -"
	case watch.Error:
		c = color.Purple
		text = " ✖"
	}
	if _, err := c.Fprint(pp.w, text); err != nil {
		return err
	}
	if _, err := pp.w.Write([]byte("\t")); err != nil {
		return err
	}
	return nil
}*/

func newParPrinter(out io.Writer) *parPrinter {
	tablePrinter := &printer.TablePrinter{}
	tabWriter := printer.GetNewTabWriter(out)
	_, _ = fmt.Fprint(tabWriter, "EVENT\t")
	_ = tablePrinter.PrintHeader(tabWriter)

	tabWriter.SetRememberedWidths([]int{8, 50, 20, 4})

	return &parPrinter{
		w: tabWriter,
		p: tablePrinter,
	}
}

func WatchAllServerResources(ctx context.Context, flags *genericclioptions.ConfigFlags, out io.Writer) error {
	useCache := viper.GetBool(constants.FlagUseCache)
	scope := viper.GetString(constants.FlagScope)

	grs, err := fetchAvailableGroupResources(useCache, scope, flags)
	if err != nil {
		return errors.Wrap(err, "fetch available group resources")
	}

	resources := extractRelevantResources(grs, getExclusions())

	watchMultipleResources(ctx, flags, out, resources)
	return nil
}

func watchMultipleResources(ctx context.Context, flags resource.RESTClientGetter, out io.Writer, resourceTypes []groupResource) {
	resourceNames := ToResourceTypes(resourceTypes)
	logrus.Debugf("Resources to watch: %s", resourceNames)

	p := newParPrinter(out)

	wg := &sync.WaitGroup{}

	for _, resourceName := range resourceNames {
		wg.Add(1)
		go func(r string) {
			defer wg.Done()
			watchSingleResource(ctx, flags, p, r)
		}(resourceName)
	}
	wg.Wait()

	return
}

type resourceHandle struct {
	name, namespace string
}

func watchSingleResource(ctx context.Context, flags resource.RESTClientGetter, p *parPrinter, resourceName string) {
	ns := viper.GetString(constants.FlagNamespace)
	selector := viper.GetString(constants.FlagSelector)
	fieldSelector := viper.GetString(constants.FlagFieldSelector)

	request := resource.NewBuilder(flags).
		Unstructured().
		ResourceTypes(resourceName).
		NamespaceParam(ns).DefaultNamespace().AllNamespaces(ns == "").
		LabelSelectorParam(selector).FieldSelectorParam(fieldSelector).SelectAllParam(selector == "" && fieldSelector == "").
		Latest().
		Do()

	obj, err := request.Object()
	if err != nil {
		logrus.Warnf("fetching resource %s failed: %s", resourceName, err)
		return
	}

	rv := "0"
	rv, err = meta.NewAccessor().ResourceVersion(obj)
	if err != nil {
		return
	}

	watcher, err := request.Watch(rv)
	if err != nil {
		logrus.Warnf("initializing watcher for %s failed: %s", resourceName, err)
		return
	}

	seen := make(map[resourceHandle]runtime.Object)

	_, _ = watchtools.UntilWithoutRetry(ctx, watcher, func(e watch.Event) (bool, error) {
		//if e.Type == watch.Modified {
		//	return false, nil
		//}

		err = p.print(e)

		accessor, _ := meta.Accessor(e.Object)
		h := resourceHandle{
			namespace: accessor.GetNamespace(),
			name: accessor.GetName(),
		}
		if last, ok := seen[h]; ok {
			fmt.Fprintf(os.Stdout, "%s\n", cmp.Diff(last, e.Object))
		}
		seen[h] = e.Object

		return false, err
	})
}
