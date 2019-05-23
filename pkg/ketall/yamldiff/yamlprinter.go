package yamldiff

import (
	"errors"
	"fmt"
	"github.com/kr/text"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"reflect"
	"regexp"
)

type printerConfig struct {
	w io.Writer
	//indentWidth int
	color bool
}

var blank = regexp.MustCompile(" *")

type color int

const (
	red    = color(31)
	green  = color(32)
	purple = color(95)
	none   = color(0)
)

func (c color) Print(s string) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", c, s)
}

type yamlPrinter struct {
	indent     string
	path       string
	nodeType   int
	headerDone bool
	config     *printerConfig
}

func Diff(a, b interface{}) error {
	return FDiff(os.Stdout, a, b)
}

func FDiff(w io.Writer, a, b interface{}) error {
	return yamlPrinter{
		nodeType: nodeRoot,
		config:   &printerConfig{w: w},
	}.do(makeDiffTree(a, b))
}

func (y yamlPrinter) do(dt *diffTree) error {
	if dt == nil {
		return nil
	}

	switch dt.nodeType {
	case nodeObject, nodeRoot:
		var path string
		indent := y.indent
		nodeType := dt.nodeType
		if isList(dt.fields) {
			// never print lists inline
			y.header(dt.route)
		} else if len(dt.fields) > 1 {
			// start a new block if there are more than one children
			if !blank.MatchString(y.path) || dt.route != "" {
				indent += "  "
				y.header(dt.route)
			}
		} else {
			// if only one field, make the yaml route compact
			if y.path == "" {
				path = dt.route
			} else {
				path = fmt.Sprintf("%s.%s", y.path, dt.route)
			}
			// mark list of objects
			if y.nodeType == nodeList || y.nodeType == nodeMixed {
				nodeType = nodeMixed
			}
		}
		yc := yamlPrinter{path: path, indent: indent, nodeType: nodeType, config: y.config}

		for _, t := range dt.fields {
			if err := yc.do(t); err != nil {
				return err
			}
		}

	case nodeList:
		yc := yamlPrinter{path: y.path, indent: y.indent, nodeType: dt.nodeType, config: y.config}
		for _, t := range dt.fields {
			if err := yc.do(t); err != nil {
				return err
			}
		}

	case nodeLeaf:
		left, inlineLeft, errLeft := stringify(dt.left)
		right, inlineRight, errRight := stringify(dt.right)
		if errLeft != nil || errRight != nil {
			return fmt.Errorf("printing leaf at %s", y.path)
		}
		return y.print(left, right, inlineLeft && inlineRight)

	default:
		return errors.New("unknown node type")
	}
	return nil
}

func isList(dt []*diffTree) bool {
	for _, t := range dt {
		if t.nodeType == nodeList {
			return true
		}
	}
	return false
}

func (y yamlPrinter) print(l string, r string, inline bool) (err error) {
	label := ""
	if y.path != "." && y.path != "" {
		label = fmt.Sprintf(" %s:", y.path)
	}

	l, r = markup(inline, y.indent, l, r)

	if !inline {
		_, err = fmt.Fprintf(y.config.w, "%s%s\n%s%s", y.indent, label, l, r)
		return
	}

	switch y.nodeType {
	case nodeMixed:
		_, err = fmt.Fprintf(y.config.w, " %s-%s %s%s\n", y.indent, label, l, r)
	case nodeList:
		_, err = fmt.Fprintf(y.config.w, " %s- %s%s\n", y.indent, l, r)
	default:
		_, err = fmt.Fprintf(y.config.w, "%s%s %s%s\n", y.indent, label, l, r)
	}
	return
}

func markup(inline bool, indent, l, r string) (string, string) {
	if inline {
		return red.Print("[-" + l + "-]"), green.Print("{+" + r + "+}")
	}
	return red.Print(indentListItem(l, fmt.Sprintf("-%s  ", indent))),
		green.Print(indentListItem(l, fmt.Sprintf("-%s  ", indent)))
}

func (y yamlPrinter) header(name string) {
	if y.path == "" {
		fmt.Fprintf(y.config.w, " %s%s:\n", y.indent, name)
	} else {
		fmt.Fprintf(y.config.w, " %s%s.%s:\n", y.indent, y.path, name)
	}
}

func stringify(v reflect.Value) (s string, inline bool, err error) {
	if !v.IsValid() {
		return "", true, nil
	}

	switch kind := v.Type().Kind(); kind {
	case reflect.Bool:
		inline = true
		s = fmt.Sprintf("%v", v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		inline = true
		s = fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		inline = true
		s = fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		inline = true
		s = fmt.Sprintf("%v", v.Float())
	case reflect.Complex64, reflect.Complex128:
		inline = true
		s = fmt.Sprintf("%v", v.Complex())
	case reflect.String:
		// todo properly markup multiline yaml strings with `|-`
		s = fmt.Sprintf("'%s'", v.String())
		inline = true
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		s = fmt.Sprintf("%#x", v.Pointer())
		inline = true
	case reflect.Interface:
		return stringify(v.Elem())
	case reflect.Ptr:
		if v.IsNil() {
			s = "nil"
			inline = true
		} else {
			return stringify(v.Elem())
		}
	case reflect.Map, reflect.Struct:
		out, e := yaml.Marshal(v.Interface())
		if e != nil {
			err = e
			return
		}
		s = string(out)
	case reflect.Array, reflect.Slice:
		n := v.Len()
		if n == 0 {
			s = "[]\n"
			inline = true
		}
		for i := 0; i < n; i++ {
			// todo maybe not ignore this error
			itemText, _, _ := stringify(v.Index(i))
			indented := indentListItem(itemText, "  ")
			s += indented + "\n"
		}
	default:
		panic("unknown reflect Kind: " + kind.String())
	}

	return
}

func indentListItem(s, indent string) string {
	indented := text.IndentBytes([]byte(s), []byte(indent))
	indented[len(indent)-2] = '-'
	return string(indented)
}
