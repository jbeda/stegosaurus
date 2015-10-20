package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/kubernetes/pkg/conversion"
)

const contextFilename = "context.yml"
const templateDir = "templates"
const outputDir = "output"

func main() {
	context := loadContext()

	t := template.New("<root>")
	err := loadBaseTemplates(t)
	if err != nil {
		log.Fatalf("Error loading templates: %v", err)
	}

	err = evalTemplates(t, context)
	if err != nil {
		log.Fatalf("Error evaluating templates: %v", err)
	}
}

// loadTemplates walks templateDir and loads all base templates into t.  These
// are files that have a '.tmpl' extensions and a leading underscore.  The
// template name is the filename with those stripped.
func loadBaseTemplates(t *template.Template) error {
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		basename := filepath.Base(path)
		// Only handle files that start with _
		if !strings.HasPrefix(basename, "_") {
			return nil
		}

		// Only handle files with ".tmpl" extension
		ext := filepath.Ext(basename)
		if ext != ".tmpl" {
			return nil
		}

		fmt.Printf("Loading template file: %v\n", path)

		// Strip off "_" and ".tmpl"
		name := strings.TrimPrefix(strings.TrimSuffix(basename, filepath.Ext(basename)), "_")
		data, err := ioutil.ReadFile(path)
		_, err = t.New(name).Parse(string(data))
		if err != nil {
			return err
		}
		return nil
	})
}

// evalTemplates walks templateDir and either copies files or evaluates
// templates.  The results are put into outputDir Anything that has a '.tmpl'
// extension but no leading underscore is evaluated.  Any non-template file is
// just copied.
//
// Template files are preprocessed first to extract an optional leading YAML
// document.  This is merged into the template evaluation context.
func evalTemplates(t *template.Template, context interface{}) error {
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		basename := filepath.Base(path)

		// Skip if the file begins with "_"
		if strings.HasPrefix(basename, "_") {
			return nil
		}

		// Find where we are going to put this output
		rel, err := filepath.Rel(templateDir, path)
		if err != nil {
			return err
		}
		output := filepath.Join(outputDir, rel)

		ext := filepath.Ext(basename)
		if ext == ".tmpl" {
			output = strings.TrimSuffix(output, ".tmpl")
			fmt.Printf("Rendering %v to %v\n", path, output)
			return evalTemplate(path, output, t, context)
		} else {
			fmt.Printf("Copying %v to %v\n", path, output)
			return copyFile(path, output)
		}
		return nil
	})
}

// evalTemplate evaluates a single template (path) and saves it to output.
func evalTemplate(path string, output string, t *template.Template, context interface{}) error {
	newT, err := t.Clone()
	if err != nil {
		return err
	}
	tData, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Pull out front matter
	tData, context, err = processFrontmatter(tData, context)
	if err != nil {
		return err
	}

	_, err = newT.Parse(string(tData))
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(output), os.ModePerm)
	if err != nil {
		return err
	}

	of, err := os.Create(output)
	if err != nil {
		return err
	}
	defer of.Close()

	// Ideally we wouldn't have to name the template here, but
	// https://github.com/golang/go/issues/12996
	return newT.ExecuteTemplate(of, "<root>", context)
}

// processFrontmatter strips a leading YAML document from data and updates
// context with its contents.  The remaining doc, and cloned and updated
// context are returned.
func processFrontmatter(data []byte, context interface{}) ([]byte, interface{}, error) {
	r := regexp.MustCompile(`(?m)^---\s*$`)
	fmStart := r.FindIndex(data)
	if fmStart == nil || fmStart[0] != 0 {
		return data, context, nil
	}
	retData := data[fmStart[1]:]
	fmEnd := r.FindIndex(retData)
	if fmEnd == nil {
		return nil, nil, fmt.Errorf("Cannot find end of front matter.")
	}

	fmData := retData[0:fmEnd[0]]
	retData = retData[fmEnd[1]:]

	// Clone the context and update it with the front matter.
	cloner := conversion.NewCloner()
	retContext, err := cloner.DeepCopy(context)
	if err != nil {
		return nil, nil, err
	}
	err = yaml.Unmarshal(fmData, &retContext)
	if err != nil {
		return nil, nil, err
	}

	return retData, retContext, nil
}

// copyFile copies a file from path to output.  It makes no attempt to handle
// permissions, links or non-normal files.
func copyFile(path, output string) error {
	s, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer s.Close()

	err = os.MkdirAll(filepath.Dir(output), os.ModePerm)
	if err != nil {
		return err
	}

	d, err := os.Create(output)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}
	return nil
}

// loadContext parses contextFilename into a new map and returns it.
func loadContext() map[interface{}]interface{} {
	ret := make(map[interface{}]interface{})
	f, err := os.Open(contextFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return ret
		}
		log.Fatalf("Could not open %s: %v", contextFilename, err)
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("Could not read %s: %v", contextFilename, err)
	}

	err = yaml.Unmarshal(data, &ret)
	if err != nil {
		log.Fatalf("Could not unmarshal %s: %v", contextFilename, err)
	}

	return ret
}
