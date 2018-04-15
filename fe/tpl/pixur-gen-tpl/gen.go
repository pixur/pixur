package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const tplContent = `
package tpl
const (
{{range .}}
	{{.VarName}} = {{.VarContent}}
{{end}}
)
`

type filelike struct {
	Name, Content string
}

// TODO: handle directories
func (fl *filelike) VarName() string {
	nosuffix := strings.TrimSuffix(fl.Name, filepath.Ext(fl.Name))
	parts := strings.Split(nosuffix, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}

	return strings.Join(parts, "")
}

func (fl *filelike) VarContent() string {
	return strconv.Quote(fl.Content)
}

func get() ([]*filelike, error) {
	var files []*filelike
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.Contains(path, ".html") {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			data, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			files = append(files, &filelike{
				Name:    path,
				Content: string(data),
			})
		}
		return nil
	})
	return files, err
}

func main() {
	flag.Parse()
	fls, err := get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't generate templates: %v\n", err)
		os.Exit(1)
		return
	}
	tpl, err := template.New("tpl.go").Parse(tplContent)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	if err := tpl.Execute(&b, fls); err != nil {
		fmt.Fprintf(os.Stderr, "Can't execute templates: %v\n", err)
		os.Exit(1)
		return
	}
	dst, err := format.Source(b.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't format Go code: %v\n", err)
		os.Exit(1)
		return
	}
	w, err := os.Create(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't open destination file: %v\n", err)
		os.Exit(1)
		return
	}
	defer w.Close()
	if _, err := w.Write(dst); err != nil {
		fmt.Fprintf(os.Stderr, "Can't write destination file: %v\n", err)
		os.Exit(1)
		return
	}

}
