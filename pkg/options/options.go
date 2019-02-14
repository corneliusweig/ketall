package options

import (
	"github.com/corneliusweig/ketall/pkg/printer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
)

type KetallOptions struct {
	CfgFile         string
	GenericCliFlags *genericclioptions.ConfigFlags
	PrintFlags      KAPrintFlags
}

func NewCmdOptions() *KetallOptions {
	return &KetallOptions{
		GenericCliFlags: genericclioptions.NewConfigFlags(),
		PrintFlags:      KAPrintFlags{genericclioptions.NewPrintFlags("")},
	}
}

type KAPrintFlags struct {
	*genericclioptions.PrintFlags
}

func (f *KAPrintFlags) ToPrinter() (printers.ResourcePrinter, error) {
	if f.OutputFormat == nil || *f.OutputFormat == "" {
		return printer.BasicTablePrinter{}, nil
	}
	return f.PrintFlags.ToPrinter()
}
