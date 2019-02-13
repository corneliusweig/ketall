package pkg

import "k8s.io/cli-runtime/pkg/genericclioptions"

type CmdOptions struct {
	CfgFile string
	GenericCliFlags *genericclioptions.ConfigFlags
}

func NewCmdOptions() *CmdOptions{
	return &CmdOptions{
		GenericCliFlags: genericclioptions.NewConfigFlags(),
	}
}
