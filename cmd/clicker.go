package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"phishingAutoClicker/common"
	"phishingAutoClicker/config"
	"phishingAutoClicker/utils"
	"strconv"
)

var (
	confName = flag.String("c", "./config.json", "config file after encryption")
	workMode = flag.Int("m", 1, "work mode, 1=email, 2=link") // 1=email, 2=link
)

func init() {
	flag.Parse()
}

func main() {
	// print version
	log.Println("phishingAutoClicker version:", common.VERSION)
	// no multiple instance
	if !pidLock() {
		panic(errors.New("multiple instance is not allowed"))
	}
	// load config from file
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
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
	go utils.KillStartedProcess()
	go conf.StartWorker(*workMode)
	<-sig
	cleanPIDLock()
	os.Exit(0)
}

func pidLock() bool {
	myPid := os.Getpid()
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pidLockPath := cwd + "/" + ".phishing-auto-clicker.lock"
	pidLockPath, _ = filepath.Abs(pidLockPath)
	if _, err = os.Stat(pidLockPath); os.IsNotExist(err) {
		// only single instance
		err = ioutil.WriteFile(pidLockPath, []byte(strconv.Itoa(myPid)), 0644)
		if err != nil {
			panic(err)
		}
		return true
	} else {
		// multiple instance, exit with EPERM
		log.Fatalln("DO NOT RUN MULTIPLE INSTANCE SIMULTANEOUSLY!")
	}
	return false
}

func cleanPIDLock() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	pidLockPath := cwd + "/" + ".phishing-auto-clicker.lock"
	pidLockPath, _ = filepath.Abs(pidLockPath)
	os.Remove(pidLockPath)
}
