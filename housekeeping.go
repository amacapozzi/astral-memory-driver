package main

import (
	"fmt"
	"os"
	"time"
)

func SetUp(handleMode string, serviceName string, driverFullPath string) bool {
	if err := EnableSeDebugPrivilege(); err != nil {
		LogStatus("Failed to enable SeDebugPrivilege", err, false)
		LogStatus("Make sure you are running as privileged user", nil, false)
		return false
	}
	if handleMode == HANDLEMODE_PROCEXP {
		return SetUpDriverMode(serviceName, driverFullPath)
	}
	return true
}

func SetUpDriverMode(serviceName string, driverFullPath string) bool {
	if err := WriteDriverOnDisk(driverFullPath); err != nil {
		LogStatus("Failed to set up service", err, false)
		return false
	}
	if err := SetUpService(serviceName, driverFullPath); err != nil {
		LogStatus("Failed to set up service", err, false)
		return false
	}

	for i := 0; i < 3; i++ {
		err := VerifyServiceRunning(serviceName)
		if err == nil {
			break
		}
		if err.Error() == ErrServiceStartPending {
			time.Sleep(2 * time.Second)
			continue
		}
		if i == 2 {
			LogStatus("Failed to start service", err, false)
			return false
		}
	}

	if err := EnableSeDebugPrivilege(); err != nil {
		LogStatus("Failed to enable SeDebugPrivilege", err, false)
		LogStatus("Make sure you are running as privileged user", nil, false)
		return false
	}

	return true
}

func CleanUp(serviceName string, driverFullPath string, handleMode string) {
	if handleMode != HANDLEMODE_PROCEXP {
		return
	}
	if err := RemoveService(serviceName, driverFullPath); err != nil {
		LogStatus(fmt.Sprintf("Failed to remove service with servicename: %s", serviceName), err, false)
	}
	if err := os.Remove(driverFullPath); err != nil {
		LogStatus(fmt.Sprintf("Failed to delete driver file at: %s", driverFullPath), err, false)
	}
}
