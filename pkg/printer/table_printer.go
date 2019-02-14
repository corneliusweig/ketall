package printer

import (
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
	"reflect"
	"strings"
	"time"
)

type TablePrinter struct{}

const (
	tableRow = "%s\t%s\t%s\t\n"
)

func (*TablePrinter) PrintObj(r runtime.Object, w io.Writer) error {
	if printers.InternalObjectPreventer.IsForbidden(reflect.Indirect(reflect.ValueOf(r)).Type().PkgPath()) {
		return fmt.Errorf(printers.InternalObjectPrinterErr)
	}

	if r.GetObjectKind().GroupVersionKind().Empty() {
		return fmt.Errorf("missing apiVersion or kind; try GetObjectKind().SetGroupVersionKind() if you know the type")
	}

	if err := printObj(r, w); err != nil {
		return err
	}
	return nil
}

func (*TablePrinter) PrintHeader(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s\t%s\t%s\n", "NAME", "NAMESPACE", "AGE")
	return err
}

func printObj(o runtime.Object, w io.Writer) error {
	groupKind := getObjectGroupKind(o)

	acc, err := meta.Accessor(o)
	if err != nil {
		return err
	}

	name := fullName(acc.GetName(), groupKind)
	timestamp := acc.GetCreationTimestamp()
	namespace := acc.GetNamespace()
	if namespace == "" {
		namespace = "<n/a>"
	}
	if _, err := fmt.Fprintf(w, tableRow, name, namespace, translateTimestampSince(timestamp)); err != nil {
		return err
	}
	return nil
}

func getObjectGroupKind(obj runtime.Object) schema.GroupKind {
	if obj == nil {
		return schema.GroupKind{Kind: "<unknown>"}
	}
	groupVersionKind := obj.GetObjectKind().GroupVersionKind()
	if len(groupVersionKind.Kind) > 0 {
		return groupVersionKind.GroupKind()
	}

	if uns, ok := obj.(*unstructured.Unstructured); ok {
		if len(uns.GroupVersionKind().Kind) > 0 {
			return uns.GroupVersionKind().GroupKind()
		}
	}

	return schema.GroupKind{Kind: "<unknown>"}
}

func fullName(name string, groupKind schema.GroupKind) string {
	if len(groupKind.Kind) == 0 {
		return name
	}

	if len(groupKind.Group) == 0 {
		return fmt.Sprintf("%s/%s", strings.ToLower(groupKind.Kind), name)
	}

	return fmt.Sprintf("%s.%s/%s", strings.ToLower(groupKind.Kind), groupKind.Group, name)
}

func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
