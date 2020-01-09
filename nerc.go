package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"text/template"
)

const IMAGE_URL_COL = 21
const PRODUCT_NAME_COL = 6
const PRICE_COL = 14

var purgeOnly bool
var verbose bool

type TemplateVariable struct {
	Key          string      `yaml:"key"`
	CSVSourceCol int         `yaml:"csvSourceCol"`
	Value        interface{} `yaml:"value"`
	Type         string      `yaml:"type"`
}

type NercConf struct {
	Input           string                 `yaml:"input"`
	Templates       string                 `yaml:"templates"`
	Output          string                 `yaml:"output"`
	Variables       []TemplateVariable     `yaml:"variables"`
	StaticVariables map[string]interface{} `yaml:"staticVariables"`
	CSVMapping      map[string]int         `yaml:"csvMapping"`
}

func main() {
	flag.BoolVar(&purgeOnly, "purge", false, "Purge all existing files from output directory and stop.")
	flag.BoolVar(&verbose, "v", false, "Use verbose output.")
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
		purgeOutput(nercConf, !verbose)

		if purgeOnly {
			fmt.Println("Purged output directory")
		} else {
			fmt.Println("Reading input file: " + nercConf.Input)
			csvFile, _ := os.Open(nercConf.Input)
			r := csv.NewReader(bufio.NewReader(csvFile))
			os.Mkdir(nercConf.Output, os.ModePerm)
			csvToConfigs(r, nercConf)
		}
	}

	fmt.Println("Done")
}

func purgeOutput(nercConf NercConf, silent bool) {
	err := os.RemoveAll(nercConf.Output)
	if err != nil && !silent {
		fmt.Println(err)
	}
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
	firstLine := true
	for {
		row, err := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if firstLine {
			// Treat first line as header line and skip it
			firstLine = false
		} else {
			for i, templateFile := range files {
				writeConf(row, templateFile, i, nercConf)
				configCount += 1
			}
		}
	}
	fmt.Println("Wrote " + strconv.Itoa(configCount) + " config files to " + nercConf.Output)
}

func toPriceString(price interface{}) (string, error) {
	if s, err := strconv.ParseFloat(fmt.Sprintf("%v", price), 64); err == nil {
		return fmt.Sprintf("%.2f", s), nil
	} else {
		return "", errors.New(fmt.Sprintf("Could not convert '%v' to a price string", price))
	}
}

func writeConf(row []string, template string, i int, nercConf NercConf) {
	templateVars := make(map[string]interface{})
	for _, variable := range nercConf.Variables {
		value := string(row[variable.CSVSourceCol])
		if variable.Type == "price" && value != "" {
			price, err := toPriceString(value)
			if err != nil {
				fmt.Println("Error in sku " + row[0] + ": " + err.Error())
			} else {
				value = price
			}
		}
		templateVars[variable.Key] = value
	}
	conf := ProcessFile(template, templateVars)
	outputFile := "sku_" + row[0] + "_version_" + strconv.Itoa(i) + ".json"
	err := ioutil.WriteFile(path.Join(nercConf.Output, outputFile), []byte(conf), 0644)
	if err != nil {
		fmt.Println(err)
	}
}
