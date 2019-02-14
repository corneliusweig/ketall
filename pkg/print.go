package pkg

import (
	"github.com/corneliusweig/ketall/pkg/client"
	"github.com/corneliusweig/ketall/pkg/options"
)

func Main(gaOptions *options.CmdOptions) {
	//cmd.GAOptions.GenericCliFlags
	_ = client.PrintAllServerResources(gaOptions)
}
