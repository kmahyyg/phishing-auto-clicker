package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"phishingAutoClicker/common"
	"phishingAutoClicker/config"
	"phishingAutoClicker/utils"
)

var (
	confName = flag.String("c", "./config.json", "config file")
	workMode = flag.Int("m", 1, "work mode") // 1=email, 2=link
)

func init() {
	flag.Parse()
}

func main() {
	// load config from file
	log.Println("Start read config file.")
	conf := config.MailConfigFile{}
	err := conf.Load(*confName)
	if err != nil {
		panic(err)
	}
	// check last char of conf.protocol if ending in s
	if conf.Protocol[len(conf.Protocol)-1:] == "s" {
		conf.IsTLS = true
		conf.Protocol = conf.Protocol[:len(conf.Protocol)-1]
	}
	// get temp folder
	if conf.SaveTo == "" {
		// build temp folder
		log.Println("Create storage for temporary file.")
		conf.SaveTo, err = os.MkdirTemp("", "foolguy-")
		if err != nil {
			panic(err)
		}
	} else {
		a, b := utils.CheckExists(conf.SaveTo)
		if !a || b != 1 {
			// todo: create folder
			panic(errors.New("storage folder not exists"))
		}
	}
	defer os.RemoveAll(conf.SaveTo)
	common.GlobalTemporaryStorage = conf.SaveTo
	// start clicker
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	log.Println("Start clicker worker. Press ^C to terminate.")
	go conf.StartWorker(*workMode)
	<-sig
	os.Exit(0)
}
