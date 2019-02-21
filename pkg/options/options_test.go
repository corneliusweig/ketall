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

package options

import (
	"testing"

	"github.com/corneliusweig/ketall/pkg/printer"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
)

func TestKAPrintFlags_ToPrinter(t *testing.T) {
	flags := KAPrintFlags{genericclioptions.NewPrintFlags("")}

	flags.OutputFormat = nil
	p, err := flags.ToPrinter()
	assert.NoError(t, err)
	assert.IsType(t, &printer.TablePrinter{}, p)

	format := ""
	flags.OutputFormat = &format
	p, err = flags.ToPrinter()
	assert.NoError(t, err)
	assert.IsType(t, &printer.TablePrinter{}, p)

	format = "json"
	flags.OutputFormat = &format
	p, err = flags.ToPrinter()
	assert.NoError(t, err)
	assert.IsType(t, &printers.JSONPrinter{}, p)
}
