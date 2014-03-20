package templar

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var templars = map[string]*template.Template{}

func children(dir string) (result []string, err error) {
	dirFile, err := os.Open(dir)
	if err != nil {
		return
	}
	defer dirFile.Close()
	files, err := dirFile.Readdir(-1)
	if err != nil {
		return
	}
	for _, file := range files {
		if strings.Index(file.Name(), ".") != 0 {
			if file.IsDir() {
				var subChildren []string
				subChildren, err = children(filepath.Join(dir, file.Name()))
				if err != nil {
					return
				}
				result = append(result, subChildren...)
			} else {
				result = append(result, filepath.Join(dir, file.Name()))
			}
		}
	}
	return
}

func Add(name, text string) (err error) {
	tmpl := template.New(name)
	if tmpl, err = tmpl.Parse(text); err != nil {
		return
	}
	templars[name] = tmpl
	return
}

func Get(name string) (result *template.Template, err error) {
	result, found := templars[name]
	if found {
		return
	}
	return template.ParseFiles(name)
}

func Generate(dir, dst string) (err error) {
	dst, err = filepath.Abs(dst)
	if err != nil {
		return
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}
	templates, err := children(dir)
	if err != nil {
		return
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}
	defer dstFile.Close()
	if _, err = fmt.Fprintf(dstFile, `package %v
import "text/templates"
import "github.com/zond/templar"
func init() {
`, filepath.Base(filepath.Dir(dst))); err != nil {
		return
	}
	buf := &bytes.Buffer{}
	var templateFile *os.File
	for _, template := range templates {
		buf.Reset()
		templateFile, err = os.Open(template)
		if err != nil {
			return
		}
		_, err = io.Copy(buf, templateFile)
		templateFile.Close()
		if err != nil {
			return
		}
		if _, err = fmt.Fprintf(dstFile, `  if err := templar.Add(%#v, %#v); err != nil {
    panic(err)
  }
`, template[len(filepath.Dir(dir))+1:], buf.String()); err != nil {
			return
		}
	}
	if _, err = fmt.Fprintln(dstFile, "}"); err != nil {
		return
	}
	return
}
