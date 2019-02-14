package pkg

import (
	"github.com/corneliusweig/ketall/pkg/client"
	"github.com/corneliusweig/ketall/pkg/options"
)

func Main(ketallOptions *options.KetallOptions) {
	_ = client.PrintAllServerResources(ketallOptions)
}
