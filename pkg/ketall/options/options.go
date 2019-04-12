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
	"bytes"
	"os"

	"github.com/corneliusweig/ketall/pkg/ketall/printer"
	"github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
)

type KetallOptions struct {
	CfgFile         string
	GenericCliFlags *genericclioptions.ConfigFlags
	PrintFlags      KAPrintFlags
	UseCache        bool
	Scope           string
	Since           string
	Selector        string
	Exclusions      []string
	Streams         *genericclioptions.IOStreams
}

func NewCmdOptions() *KetallOptions {
	return &KetallOptions{
		GenericCliFlags: genericclioptions.NewConfigFlags(true),
		PrintFlags:      KAPrintFlags{genericclioptions.NewPrintFlags("")},
		Streams: &genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

// Sets up options with in-memory buffers as in- and output-streams
func NewTestTestCmdOptions() (*KetallOptions, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	iostreams, in, out, errout := genericclioptions.NewTestIOStreams()
	logrus.SetOutput(errout)
	return &KetallOptions{
		GenericCliFlags: genericclioptions.NewConfigFlags(true),
		PrintFlags:      KAPrintFlags{genericclioptions.NewPrintFlags("")},
		Streams:         &iostreams,
	}, in, out, errout
}

type KAPrintFlags struct {
	*genericclioptions.PrintFlags
}

func (f *KAPrintFlags) ToPrinter() (printers.ResourcePrinter, error) {
	if f.OutputFormat == nil || *f.OutputFormat == "" {
		return &printer.TablePrinter{}, nil
	}
	return f.PrintFlags.ToPrinter()
}
