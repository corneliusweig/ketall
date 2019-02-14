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

package pkg

import (
	"github.com/corneliusweig/ketall/pkg/client"
	"github.com/corneliusweig/ketall/pkg/options"
	"github.com/corneliusweig/ketall/pkg/printer"
	"github.com/sirupsen/logrus"
	"io"
	"text/tabwriter"
)

func KetAll(w io.Writer, ketallOptions *options.KetallOptions) {
	all, err := client.GetAllServerResources(ketallOptions)
	if err != nil {
		logrus.Fatal(err)
	}

	resourcePrinter, err := ketallOptions.PrintFlags.ToPrinter()
	if err != nil {
		logrus.Fatal(err)
	}

	out := w
	if p, ok := resourcePrinter.(*printer.TablePrinter); ok {
		logrus.Debug("Using tabwriter")
		tw := tabwriter.NewWriter(w, 4, 4, 2, ' ', 0)
		defer tw.Flush()
		out = tw
		if err := p.PrintHeader(out); err != nil {
			logrus.Fatal("print header", err)
		}
	} else {
		logrus.Debug("Using default writer")
	}

	printer := printer.NewListAdapterPrinter(resourcePrinter)
	if err = printer.PrintObj(all, out); err != nil {
		logrus.Fatal(err)
	}
}
