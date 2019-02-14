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
	"github.com/pkg/errors"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
)

type ListAdapterPrinter struct {
	delegate printers.ResourcePrinter
}

func NewListAdapterPrinter(printer printers.ResourcePrinter) ListAdapterPrinter {
	return ListAdapterPrinter{printer}
}

func (n *ListAdapterPrinter) PrintObj(r runtime.Object, w io.Writer) error {
	if meta.IsListType(r) {
		subs, err := meta.ExtractList(r)
		if err != nil {
			return errors.Wrap(err, "extract resource list")
		}
		for _, o := range subs {
			if err := n.PrintObj(o, w); err != nil {
				return err
			}
		}
		return nil
	}

	return n.delegate.PrintObj(r, w)
}
