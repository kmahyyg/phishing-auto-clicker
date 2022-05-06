package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"phishingAutoClicker/utils"
)

var (
	inputFilepath  = flag.String("i", "./config.json.ori", "input file")
	outputFilepath = flag.String("o", "./config.json", "output file")
	XOR_KEY        = []byte{0x8A}
)

func init() {
	flag.Parse()
}

func main() {
	if boolStat, stat := utils.CheckExists(*inputFilepath); boolStat == false || stat < 0 {
		panic("input file not found")
	}
	log.Println("Input file found")
	log.Println("Reading input file")
	readData, err := ioutil.ReadFile(*inputFilepath)
	if err != nil {
		panic(err)
	}
	log.Println("XOR-ing input file")
	err = os.WriteFile(*outputFilepath, utils.XORStream(XOR_KEY, readData), 0644)
	if err != nil {
		panic(err)
	}
	log.Println("Done.")
}
