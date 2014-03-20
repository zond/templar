package templar

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var templates = map[string]*template.Template{}
var blobs = map[string][]byte{}

func GetMatchingBlobNames(diskSearch bool, reg string) (result []string, err error) {
	pat, err := regexp.Compile(reg)
	if err != nil {
		return
	}
	for k, _ := range blobs {
		if pat.MatchString(k) {
			result = append(result, k)
		}
	}
	if len(result) > 0 {
		return
	}
	if diskSearch {
		var allChildren []string
		allChildren, err = children(".")
		if err != nil {
			return
		}
		for _, child := range allChildren {
			if pat.MatchString(child) {
				result = append(result, child)
			}
		}
	}
	return
}

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

func AddBlob(name string, text string) {
	blobs[name] = []byte(text)
}

func AddTemplate(name, text string) (err error) {
	tmpl := template.New(name)
	if tmpl, err = tmpl.Parse(text); err != nil {
		return
	}
	templates[name] = tmpl
	return
}

func GetMatchingTemplates(diskSearch bool, baseName string, reg string) (result *template.Template, err error) {
	pat, err := regexp.Compile(reg)
	if err != nil {
		return
	}
	result = template.New(baseName)
	if _, err = result.Parse(""); err != nil {
		return
	}
	for name, templar := range templates {
		if pat.MatchString(name) {
			if _, err = result.AddParseTree(filepath.Base(name), templar.Tree); err != nil {
				err = fmt.Errorf("Error adding %#v to result: %v", filepath.Base(name), err)
				return
			}
		}
	}
	if len(result.Templates()) > 1 {
		return
	}
	if diskSearch {
		var allChildren []string
		allChildren, err = children(".")
		if err != nil {
			return
		}
		filenames := []string{}
		for _, child := range allChildren {
			if pat.MatchString(child) {
				filenames = append(filenames, child)
			}
		}
		if result, err = result.ParseFiles(filenames...); err != nil {
			return
		}
	}
	return
}

func GetTemplate(diskSearch bool, name string) (result *template.Template, err error) {
	result, found := templates[name]
	if found {
		return
	}
	if diskSearch {
		result, err = template.ParseFiles(name)
	}
	return
}

func GetBlob(diskSearch bool, name string) (result io.ReadCloser, err error) {
	b, found := blobs[name]
	if found {
		result = ioutil.NopCloser(bytes.NewBuffer(b))
		return
	}
	result, err = os.Open(name)
	return
}

func GenerateTemplates(dir, dst string) (err error) {
	return generate("if err := templar.AddTemplate(%#v, %#v); err != nil { panic(err) }", dir, dst)
}

func GenerateBlobs(dir, dst string) (err error) {
	return generate("templar.AddBlob(%#v, %#v)", dir, dst)
}

func generate(addTemplate, dir, dst string) (err error) {
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
		if _, err = fmt.Fprintf(dstFile, "  "+addTemplate+"\n", template[len(filepath.Dir(dir))+1:], buf.String()); err != nil {
			return
		}
	}
	if _, err = fmt.Fprintln(dstFile, "}"); err != nil {
		return
	}
	return
}
