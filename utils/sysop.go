package utils

import (
	"log"
	"os/exec"
	"runtime"
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
