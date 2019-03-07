// Copyright © 2019 Christian Rolly <mail@chromium-solutions.de>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/flect"

	"github.com/spf13/cobra"
)

// resourceCmd represents the resource command
var (
	resourceCmd = &cobra.Command{
		Use:   "resource name [flags]",
		Short: "Adds CRUDL functions for the defined resource",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// instantiate new resource model and parse given attributes
			modelName := args[0]
			m := newModel(modelName, false, attributes, !noID, addDates)

			// add resource to mug.config.json
			addResourceConfig(m)
			// create resource directory
			createResourceDirectory(modelName)
			// render templates with data
			renderTemplates(m)
			// update Makefile
			renderMakefile(modelName)
			// update serverless.yml
			renderSLS()

			// write definition to resource folder
			writeResourceDefinition(m, modelName)
		},
	}

	attributes string
	noID       bool
	addDates   bool
)

func init() {
	addCmd.AddCommand(resourceCmd)
	resourceCmd.Flags().StringVarP(&attributes, "attributes", "a", "", "attributes of the resource")
	resourceCmd.Flags().BoolVarP(&noID, "noID", "n", false, "automatically generate id attribute as hash key for resource")
	resourceCmd.Flags().BoolVarP(&addDates, "addDates", "d", false, "automatically add createdAt and updatedAt attributes")
}

func addResourceConfig(m Model) {
	config := readConfig()

	singular := m.Ident.Singularize().String()
	plural := m.Ident.Pluralize().String()

	resource := Resource{
		// Name:  m.Name,
		Ident: flect.New(m.Name),
		Functions: []Function{
			Function{Name: "create" + "_" + singular, Handler: "create", Path: plural, Method: "post"},
			Function{Name: "read" + "_" + singular, Handler: "read", Path: fmt.Sprintf("%s/{id}", plural), Method: "get"},
			Function{Name: "update" + "_" + singular, Handler: "update", Path: fmt.Sprintf("%s/{id}", plural), Method: "put"},
			Function{Name: "delete" + "_" + singular, Handler: "delete", Path: fmt.Sprintf("%s/{id}", plural), Method: "delete"},
			Function{Name: "list" + "_" + plural, Handler: "list", Path: plural, Method: "get"},
		},
	}
	config.Resources[m.Name] = resource

	writeConfig(config)

}

func createResourceDirectory(name string) {
	wd := getWorkingDir()

	os.MkdirAll(filepath.Join(wd, "functions", name), 0755)
}

func renderTemplates(m Model) {
	// // load templates
	// ts, err := template.ParseGlob("./templates/functions/*.tmpl")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// iterate over templates and execute
	for _, tmpl := range functionsBox.List() {

		// create file
		fileName := filepath.Join(getWorkingDir(), "functions", m.Ident.Camelize().String(), strings.Replace(tmpl, ".tmpl", "", 1))
		f, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		// load template
		t := LoadTemplateFromBox(functionsBox, tmpl)

		// execute template and save to file
		err = t.Execute(f, m)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func renderMakefile(name string) {
	config := readConfig()

	// load Makefile template
	t := LoadTemplateFromBox(projectBox, "Makefile.tmpl")

	// open file and execute template
	f, err := os.OpenFile(filepath.Join(getWorkingDir(), "Makefile"), os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// execote template and save to file
	err = t.Execute(f, config)
	if err != nil {
		log.Fatal(err)
	}
}

func writeResourceDefinition(m Model, name string) {
	wd := getWorkingDir()

	json, _ := json.MarshalIndent(m, "", "  ")
	_ = ioutil.WriteFile(filepath.Join(wd, "functions", name, fmt.Sprintf("%s.json", name)), json, 0644)
}

func renderSLS() {
	config := readConfig()

	// load Makefile template
	t := LoadTemplateFromBox(projectBox, "serverless.yml.tmpl")

	// open file and execute template
	f, err := os.OpenFile(filepath.Join(getWorkingDir(), strings.Replace(t.Name(), ".tmpl", "", 1)), os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// execote template and save to file
	err = t.Execute(f, config)
	if err != nil {
		log.Fatal(err)
	}
}
