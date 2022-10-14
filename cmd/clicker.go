package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/MagicPiperSec/software-license/softlic"
	"phishingAutoClicker/common"
	"phishingAutoClicker/config"
	"phishingAutoClicker/utils"
)

const DEBUG_FLAG = false

var (
	confName    = flag.String("c", "./config.json", "config file after encryption")
	workMode    = flag.Int("m", 1, "work mode, 1=email, 2=link") // 1=email, 2=link
	showVersion = flag.Bool("v", false, "show version")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	flag.Parse()
}

func main() {
	// debug only
	if DEBUG_FLAG {
		log.Println("DEBUG_FLAG is true")
		log.Println("Work Mode:", *workMode)
		log.Println("Config Name:", *confName)
		log.Println("End User License Type: ", common.EndUserLicenseType)
		log.Println("End User Public Key: ", common.LicensePublicKey)
	}

	// version flag
	if *showVersion {
		fmt.Println("phishingAutoClicker version:", common.VERSION)
		return
	}

	// check license
	go func() {
		for {
			if fstat, ftype := utils.CheckExists("user.lic"); !fstat || ftype != 0 {
				log.Fatalln("user.lic not exists")
			}
			licData, err := os.ReadFile("user.lic")
			if err != nil {
				panic(err)
			}
			rand.Seed(time.Now().UnixNano())
			if DEBUG_FLAG {
				fmt.Println("licPub: ", common.LicensePublicKey)
			}
			err = softlic.ValidateLicense(common.EndUserID, common.EndUserNonce, common.EndUserLicenseType, common.LicensePublicKey, licData)
			if DEBUG_FLAG {
				fmt.Println(err)
			}
			if err != nil {
				time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
				selfExe, err := os.Executable()
				if err == nil {
					os.Remove("user.lic")
					os.Remove(selfExe)
				}
				panic("license invalid, fuck pirate!")
			}
			time.Sleep(time.Duration(rand.Intn(60))*time.Second + 10*time.Second)
		}
	}()
	// print version
	log.Println("phishingAutoClicker version: ", common.VERSION)
	// no multiple instance
	if !pidLock() {
		panic(errors.New("multiple instance is not allowed"))
	}
	defer cleanPIDLock()
	// load config from file

	log.Println("Start read config file.")
	conf := config.MailConfigFile{}
	err := conf.Load(*confName)
	if err != nil {
		if DEBUG_FLAG {
			fmt.Println(err)
		}
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
		err = os.WriteFile(pidLockPath, []byte(strconv.Itoa(myPid)), 0644)
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
