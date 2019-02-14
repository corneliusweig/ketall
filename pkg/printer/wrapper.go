package printer

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
	"text/tabwriter"
)

type WrappingPrinter struct {
	printer printers.ResourcePrinter
}

func NewWrappingPrinter(printer printers.ResourcePrinter) WrappingPrinter {
	return WrappingPrinter{printer}
}

func (wp *WrappingPrinter) PrintObject(w io.Writer, r runtime.Object) error {
	var out io.Writer

	switch wp.printer.(type) {
	case BasicTablePrinter:
		logrus.Debug("Using tabwriter")
		tw := tabwriter.NewWriter(w, 4, 4, 2, ' ', 0)
		defer tw.Flush()
		if err := wp.printHeader(tw); err != nil {
			return errors.Wrap(err, "print header")
		}
		out = tw
	default:
		logrus.Debug("Using default writer")
		out = w
	}

	allObjects, err := meta.ExtractList(r)
	if err != nil {
		return errors.Wrap(err, "extract resource list")
	}
	for _, o := range allObjects {
		if err := wp.printer.PrintObj(o, out); err != nil {
			return errors.Wrapf(err, "cannot print %s", o)
		}
	}

	return nil
}

func (*WrappingPrinter) printHeader(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\n", "NAME", "NAMESPACE", "AGE")
	return err
}
