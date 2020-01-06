package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"text/template"
)

const IMAGE_URL_COL = 21
const PRODUCT_NAME_COL = 6
const PRICE_COL = 14

var purge bool

type TemplateVariable struct {
	Key string
}

type NercConf struct {
	Input     string                 `yaml:"input"`
	Templates string                 `yaml:"templates"`
	Output    string                 `yaml:"output"`
	Variables map[string]interface{} `yaml:"variables"`
}

func main() {
	flag.BoolVar(&purge, "purge", false, "Purge all existing files from output directory.")
	flag.Parse()

	nercConf := NercConf{}
	confFile, readErr := ioutil.ReadFile("nerc.yml")
	if readErr != nil {
		panic(readErr)
	}
	parseErr := yaml.Unmarshal(confFile, &nercConf)
	if parseErr != nil {
		panic(parseErr)
	}

	if nercConf.Output == "" {
		nercConf.Output = "output/"
	}

	if _, err := os.Stat(nercConf.Input); os.IsNotExist(err) {
		fmt.Println("Could not find '" + nercConf.Input + "'. Specify input file with -i=<filepath>.")
	} else {
		fmt.Println("Reading input file: " + nercConf.Input)
		csvFile, _ := os.Open(nercConf.Input)
		r := csv.NewReader(bufio.NewReader(csvFile))

		if purge {
			fmt.Println("Purging output directory...")
			err := os.RemoveAll(nercConf.Output)
			if err != nil {
				fmt.Println(err)
			}
		}
		os.Mkdir(nercConf.Output, os.ModePerm)
		csvToConfigs(r, nercConf)
	}

	fmt.Println("Done")
}

// process applies the data structure 'vars' onto an already
// parsed template 't', and returns the resulting string.
func process(t *template.Template, vars interface{}) string {
	var tmplBytes bytes.Buffer

	err := t.Execute(&tmplBytes, vars)
	if err != nil {
		panic(err)
	}
	return tmplBytes.String()
}

// ProcessFile parses the supplied filename and compiles its template
// using the given variables.
func ProcessFile(fileName string, vars interface{}) string {
	tmpl, err := template.ParseFiles(fileName)

	if err != nil {
		panic(err)
	}
	return process(tmpl, vars)
}

func visitPath(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() {
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}

// Read given csv file and build NexRender configurations
// out of the csv and hard coded variation parameters.
func csvToConfigs(r *csv.Reader, nercConf NercConf) {
	var files []string

	if _, err := os.Stat(nercConf.Templates); !os.IsNotExist(err) {
		err := filepath.Walk(nercConf.Templates, visitPath(&files))
		if err != nil {
			panic(err)
		}
		fmt.Println(strconv.Itoa(len(files)) + " templates found from " + nercConf.Templates)
	}

	configCount := 0
	for {
		row, err := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for i, templateFile := range files {
			writeConf(row, templateFile, i, nercConf)
			configCount += 1
		}
	}
	fmt.Println("Wrote " + strconv.Itoa(configCount) + " config files to " + nercConf.Output)
}

func writeConf(row []string, template string, i int, nercConf NercConf) {
	templateVars := make(map[string]interface{})
	for k, v := range nercConf.Variables {
		if reflect.TypeOf(v).Kind() == reflect.Int {
			templateVars[k] = row[v.(int)]
		} else {
			templateVars[k] = v
		}
	}
	conf := ProcessFile(template, templateVars)
	outputFile := "sku_" + row[0] + "_version_" + strconv.Itoa(i) + ".json"
	err := ioutil.WriteFile(path.Join(nercConf.Output, outputFile), []byte(conf), 0644)
	if err != nil {
		fmt.Println(err)
	}
}
