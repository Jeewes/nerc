package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
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

var inputFile string
var purge bool
var templatesDir string
var outputDir string

func main() {
	flag.StringVar(&inputFile, "i", "input.csv", "Input file")
	flag.StringVar(&templatesDir, "t", "templates/", "Path to nexrender templates dir")
	flag.StringVar(&outputDir, "o", "output/", "Output directory path")
	flag.BoolVar(&purge, "purge", false, "Purge all existing files from output directory.")
	flag.Parse()

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Println("Could not find '" + inputFile + "'. Specify input file with -i=<filepath>.")
	} else {
		fmt.Println("Reading input file: " + inputFile)
		csvFile, _ := os.Open(inputFile)
		r := csv.NewReader(bufio.NewReader(csvFile))

		if purge {
			fmt.Println("Purging output directory...")
			err := os.RemoveAll(outputDir)
			if err != nil {
				fmt.Println(err)
			}
		}
		os.Mkdir(outputDir, os.ModePerm)
		csvToConfigs(r, templatesDir, outputDir)
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
func csvToConfigs(r *csv.Reader, templatesDir string, outputDir string) {
	var files []string

	if _, err := os.Stat(templatesDir); !os.IsNotExist(err) {
		err := filepath.Walk(templatesDir, visitPath(&files))
		if err != nil {
			panic(err)
		}
		fmt.Println(strconv.Itoa(len(files)) + " templates found from " + templatesDir)
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
			writeConf(row, templateFile, i, outputDir)
			configCount += 1
		}
	}
	fmt.Println("Wrote " + strconv.Itoa(configCount) + " config files to " + outputDir)
}

func writeConf(row []string, template string, i int, outputDir string) {
	vars := make(map[string]interface{})
	vars["Template"] = template
	vars["VideoOutputFile"] = "todo/path/sample.avi"
	vars["ProductImagePath"] = row[IMAGE_URL_COL]
	vars["ProductPrice"] = row[PRICE_COL]
	vars["ProductName"] = row[PRODUCT_NAME_COL]
	conf := ProcessFile("nerc_conf.json", vars)
	outputFile := "sku_" + row[0] + "_version_" + strconv.Itoa(i) + ".json"
	err := ioutil.WriteFile(path.Join(outputDir, outputFile), []byte(conf), 0644)
	if err != nil {
		fmt.Println(err)
	}
}
