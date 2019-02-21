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
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
)

type FlattenListAdapter struct {
	printers.ResourcePrinter
}

func NewFlattenListAdapterPrinter(printer printers.ResourcePrinter) printers.ResourcePrinter {
	logrus.Debugf("Wrapping %T with FlattenListAdapterPrinter", printer)
	return &FlattenListAdapter{printer}
}

func (n *FlattenListAdapter) PrintObj(r runtime.Object, w io.Writer) error {
	if meta.IsListType(r) {
		items, err := meta.ExtractList(r)
		if err != nil {
			return errors.Wrap(err, "extract resource list")
		}
		for _, o := range items {
			if err := n.PrintObj(o, w); err != nil {
				return errors.Wrap(err, "print list item")
			}
		}
		return nil
	}

	return n.ResourcePrinter.PrintObj(r, w)
}

type ListAdapterPrinter struct {
	printers.ResourcePrinter
}

func NewListAdapterPrinter(printer printers.ResourcePrinter) printers.ResourcePrinter {
	logrus.Debugf("Wrapping %T with ListAdapterPrinter", printer)
	return &ListAdapterPrinter{printer}
}

func (n *ListAdapterPrinter) PrintObj(r runtime.Object, w io.Writer) error {
	if meta.IsListType(r) {
		r.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "List"})
	}

	return n.ResourcePrinter.PrintObj(r, w)
}
