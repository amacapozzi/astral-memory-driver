package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/process"
)

func ValidateArguments() error {
	if *STRINGSPATH == "" {
		return CreateError(errors.New("--path is required"))
	}

	if *HANDLEMODE != HANDLEMODE_PROCEXP && *HANDLEMODE != HANDLEMODE_DIRECT {
		return CreateError(errors.New("invalid handle mode [--handle]"))
	}

	if *TARGETPID == 0 && *TARGETPROCNAME == "" {
		return CreateError(errors.New("--pid or --name required"))
	}

	if !filepath.IsAbs(*DRIVERPATH) {
		if _, err := filepath.Abs(*DRIVERPATH); err != nil {
			return CreateError(errors.New("invalid dump path " + err.Error()))
		}
	}
	return nil
}

func FillArguments() {
	*HANDLEMODE = HANDLEMODE_PROCEXP

	if !filepath.IsAbs(*DRIVERPATH) {
		*DRIVERPATH, _ = filepath.Abs(*DRIVERPATH)
	}
}

func LogStatus(message string, err error, success bool) {

	if success {
		fmt.Println(fmt.Sprintf("[+] %s", message))
		return
	}
	if err != nil {
		fmt.Println(fmt.Sprintf("[-] %s. Error: %s", message, err.Error()))
		return
	}
	fmt.Println(fmt.Sprintf("[-] %s", message))
}

func CreateError(err error) error {
	if err == nil {
		return err
	}
	var callerName = "UnknownFunction"
	if info, _, _, ok := runtime.Caller(1); ok {
		details := runtime.FuncForPC(info)
		if details != nil {
			callerName = details.Name()
		}
	}
	callerNameSplit := strings.Split(callerName, ".")
	newErrorText := fmt.Sprintf("%s error: %s", callerNameSplit[len(callerNameSplit)-1], err.Error())
	return errors.New(newErrorText)
}

func GetProcessId(pid int, name string) int {
	if pid != 0 {
		return pid
	}
	processes, err := process.Processes()
	if err != nil {
		return 0
	}
	for _, each := range processes {
		if procName, err := each.Name(); err == nil && procName == name {
			return int(each.Pid)
		}
	}
	return 0
}

func getStrings(stringPath string) ([]string, error) {
	fileContent, err := os.Open(stringPath)

	if err != nil {
		return []string{}, CreateError(err)
	}

	fileScanner := bufio.NewScanner(fileContent)

	userStrings := []string{}

	for fileScanner.Scan() {
		userStrings = append(userStrings, fileScanner.Text())
	}

	return userStrings, nil
}
