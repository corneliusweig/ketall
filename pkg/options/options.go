package options

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type KetallOptions struct {
	CfgFile         string
	GenericCliFlags *genericclioptions.ConfigFlags
	PrintFlags      *genericclioptions.PrintFlags
	Verbs           []v1.Verbs
}

func NewCmdOptions() *KetallOptions {
	return &KetallOptions{
		GenericCliFlags: genericclioptions.NewConfigFlags(),
		PrintFlags:      genericclioptions.NewPrintFlags(""),
	}
}
