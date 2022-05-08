package utils

import (
	"archive/zip"
	"context"
	"errors"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"phishingAutoClicker/common"
	"runtime"
	"strings"
	"time"
)

func KillStartedProcess() {
	// windows only
	if runtime.GOOS != "windows" {
		return
	}
	for {
		// kill all started processes
		procList := []string{
			"iexplore.exe",
			"winword.exe",
			"excel.exe",
		}
		for _, proc := range procList {
			log.Println("Trigger Windows Kill Process: ", proc)
			_ = exec.Command("taskkill", "/F", "/IM", proc).Run()
		}
		// sleep for a while
		time.Sleep(time.Second * 120)
	}
}

func tryUnzipFile(fdpath string) (exeFile string, err error) {
	tmpPath, err := os.MkdirTemp(common.GlobalTemporaryStorage, "-extract")
	if err != nil {
		return "", err
	}
	r, err := zip.OpenReader(fdpath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	log.Println("Extract files from ", fdpath)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			// do not process any directory
			continue
		}
		// all early return will result in only first file to be extracted
		fileExt := filepath.Ext(f.Name)
		log.Println("Currently extract: " + f.Name)
		if len(fileExt) == 0 {
		}
		exeFile = tmpPath + "/" + uuid.NewString() + fileExt
		exeFile, err = filepath.Abs(exeFile)
		log.Println("Extracted to: " + exeFile)
		if err != nil {
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		fileData, err := ioutil.ReadAll(rc)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(exeFile, fileData, 0755)
		if err != nil {
			return "", err
		}
		rc.Close()
		log.Println("Extract operation successfully finished.")
		return exeFile, nil
	}
	return "", errors.New("unknown reason error for unzipping")
}

func submitCredentialsToHacker(destUrl string) {
	postForm := url.Values{}
	postForm.Add("username", common.GlobalCred_Username)
	postForm.Add("password", common.GlobalCred_Password)
	_, _ = http.PostForm(destUrl, postForm)
}

func clickBaitOffline(filename string, data []byte) {
	var finalPath string
	var err error
	if !filepath.IsAbs(filename) {
		// resolve file path
		fileExt := filepath.Ext(filename)
		finalName := uuid.New().String() + fileExt
		finalPath = common.GlobalTemporaryStorage + "/" + finalName
		finalPath, err = filepath.Abs(finalPath)
		if err != nil {
			log.Println(err)
			return
		}
		// create and write file
		err = ioutil.WriteFile(finalPath, data, 0755)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		finalPath = filename
	}
	// if ended with .zip, try extract with password: infected, return extracted first file
	if strings.HasSuffix(finalPath, ".zip") {
		finalPath, err = tryUnzipFile(finalPath)
		if err != nil {
			log.Println(err)
			return
		}
	}
	// execute file
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelFunc()
	switch runtime.GOOS {
	case "windows":
		cmd := exec.CommandContext(ctx, "cmd", "/c", "start", finalPath)
		_ = cmd.Start()
		_ = cmd.Wait()
	case "linux":
		var cmd *exec.Cmd
		if !strings.HasSuffix(finalPath, ".elf") {
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", "xdg-open", finalPath)
		} else {
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", finalPath)
		}
		_ = cmd.Start()
		_ = cmd.Wait()
	default:
		log.Println("Unsupported OS")
	}
	return
}

func clickBaitOnline(url string) {
	// if url is ending with .exe / .doc / .docm / .xlsm / .elf / .zip , download and execute it
	if strings.HasSuffix(url, ".exe") || strings.HasSuffix(url, ".doc") ||
		strings.HasSuffix(url, ".docm") || strings.HasSuffix(url, ".xlsm") ||
		strings.HasSuffix(url, ".elf") || strings.HasSuffix(url, ".zip") {
		// download	files
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			return
		}
		// get extensions
		lastIdx := strings.LastIndex(url, ".")
		fileExt := url[lastIdx:]
		respData, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			log.Println(err)
			return
		}
		// save file
		savedFile := common.GlobalTemporaryStorage + "/" + uuid.New().String() + fileExt
		savedFile, err = filepath.Abs(savedFile)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Saving file: ", savedFile)
		err = ioutil.WriteFile(savedFile, respData, 0755)
		// execute file
		// use clickBaitOffline
		clickBaitOffline(savedFile, nil)
	} else if strings.HasSuffix(url, "submit") {
		// if url is ending with submit, submit our credentials
		submitCredentialsToHacker(url)
	} else {
		// click URL use IE only
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*60)
		defer cancelFunc()
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "iexplore.exe", url)
		case "linux":
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", "xdg-open", url)
		}
		_ = cmd.Start()
		_ = cmd.Wait()
	}
}
