package printer

import (
	"github.com/pkg/errors"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
)

type ListAdapterPrinter struct {
	delegate printers.ResourcePrinter
}

func NewListAdapterPrinter(printer printers.ResourcePrinter) ListAdapterPrinter {
	return ListAdapterPrinter{printer}
}

func (n *ListAdapterPrinter) PrintObj(r runtime.Object, w io.Writer) error {
	if meta.IsListType(r) {
		subs, err := meta.ExtractList(r)
		if err != nil {
			return errors.Wrap(err, "extract resource list")
		}
		for _, o := range subs {
			if err := n.PrintObj(o, w); err != nil {
				return err
			}
		}
		return nil
	}

	return n.delegate.PrintObj(r, w)
}
