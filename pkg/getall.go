package pkg

import (
	"github.com/corneliusweig/ketall/pkg/client"
	"github.com/corneliusweig/ketall/pkg/options"
	"github.com/corneliusweig/ketall/pkg/printer"
	"github.com/sirupsen/logrus"
	"os"
)

func KetAll(ketallOptions *options.KetallOptions) {
	all, err := client.GetAllServerResources(ketallOptions)
	if err != nil {
		logrus.Fatal(err)
	}

	p, err := ketallOptions.PrintFlags.ToPrinter()
	if err != nil {
		logrus.Fatal(err)
	}

	wp := printer.NewWrappingPrinter(p)
	if err = wp.PrintObject(os.Stdout, all); err != nil {
		logrus.Fatal(err)
	}
}
