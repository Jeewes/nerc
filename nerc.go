package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const IMAGE_URL_COL = 21
const PRODUCT_NAME_COL = 6
const PRICE_COL = 14

type NexRenderConf struct {
	Template TemplateConf `json:"template"`
	Assets   []AssetConf  `json:"assets"`
	Actions  ActionsConf  `json:"actions"`
}
type ActionsConf struct {
	PostRender []PostRenderConf `json:"postrender"`
}
type AssetConf struct {
	Src         string `json:"src,omitempty"`
	Type        string `json:"type"`
	LayerIndex  int    `json:"layerIndex,omitempty"`
	Composition string `json:"composition"`
	Property    string `json:"property,omitempty"`
	Expression  string `json:"expression,omitempty"`
	LayerName   string `json:"layerName,omitempty"`
}
type PostRenderConf struct {
	Module string `json:"module"`
	Input  string `json:"input"`
	Output string `json:"output"`
}
type TemplateConf struct {
	Src              string `json:"src"`
	Composition      string `json:"composition"`
	SettingsTemplate string `json:"settingsTemplate"`
	OutputModule     string `json:"outputModule"`
	OutputExt        string `json:"outputExt"`
}

//type NexRenderConf struct {
//	ProductName   string `json:"ProductName"`
//	ProductPrice  string `json:"ProductPrice"`
//	VideoTemplate string `json:"VideoTemplate"`
//}

var inputFile string
var purge bool
var templatesDir string

func main() {
	flag.StringVar(&inputFile, "i", "input.csv", "Input file")
	flag.StringVar(&templatesDir, "t", "templates/", "Path to nexrender templates dir")
	flag.BoolVar(&purge, "purge", false, "Purge all existing files from output directory.")
	flag.Parse()

	fmt.Println("Reading input file: " + inputFile)

	csvFile, _ := os.Open(inputFile)
	r := csv.NewReader(bufio.NewReader(csvFile))

	if purge {
		fmt.Println("Purging output directory...")
		err := os.RemoveAll("output")
		if err != nil {
			fmt.Println(err)
		}
	}
	os.Mkdir("output", os.ModePerm)

	configs := csvToConfigs(r, templatesDir)
	for idx, conf := range configs {
		file, _ := json.MarshalIndent(conf, "", "  ")
		filepath := "output/" + strconv.Itoa(idx) + "_tuote.json"
		err := ioutil.WriteFile(filepath, file, 0644)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("Done")
}

func visitPath(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		*files = append(*files, path)
		return nil
	}
}

// Read given csv file and build NexRender configurations
// out of the csv and hard coded variation parameters.
func csvToConfigs(r *csv.Reader, templatesDir string) []NexRenderConf {
	var configs []NexRenderConf
	var files []string

	err := filepath.Walk(templatesDir, visitPath(&files))
	if err != nil {
		panic(err)
	}

	for {
		row, err := r.Read()

		if err == io.EOF {
			fmt.Println("End of input file")
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for _, template := range files {
			conf := buildConf(row, template)
			configs = append(configs, conf)
		}
	}
	return configs
}

// Build NexRender configuration out or the csv row and template variation
func buildConf(row []string, template string) NexRenderConf {
	var conf = NexRenderConf{
		Template: TemplateConf{
			Src:              template,
			Composition:      "main",
			SettingsTemplate: "Best Settings",
			OutputModule:     "Lossless",
			OutputExt:        "avi",
		},
		Assets: []AssetConf{
			{
				Src:         row[IMAGE_URL_COL],
				Type:        "image",
				LayerIndex:  1,
				Composition: "Kuva",
			},
			{
				Type:        "data",
				LayerIndex:  1,
				Composition: "Kuva",
				Property:    "Scale",
				Expression:  "if(thisLayer.width > thisLayer.height) { s=100*thisComp.width/thisLayer.width; [s,s]; } else { s=100*thisComp.height/thisLayer.height; [s,s]; }",
			},
			{
				Type:        "data",
				LayerName:   "tuotenimi",
				Composition: "Tuoteplanssi",
				Property:    "Source Text",
				Expression:  "text.sourceText = '" + row[PRODUCT_NAME_COL] + "'",
			},
			{
				Type:        "data",
				LayerName:   "hinta",
				Composition: "Tuoteplanssi",
				Property:    "Source Text",
				Expression:  "text.sourceText ='" + row[PRICE_COL] + "'",
			},
		},
		Actions: ActionsConf{
			PostRender: []PostRenderConf{
				{
					Module: "@nexrender/action-copy",
					Input:  "result.avi",
					Output: "U:/Digimarkkinointi/Youtube-videot/Dynaaminen Youtube-video/Dynaaminen/Exports/262227016.avi",
				},
			},
		},
	}
	//confJson, _ := json.Marshal(conf)
	//fmt.Println(string(confJson))
	return conf
}
