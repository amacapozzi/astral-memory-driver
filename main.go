package main

import (
	"flag"
)

const (
	MODE_DECRYPT          = "decrypt"
	MODE_CLEANUP          = "cleanup"
	MODE_DUMP             = "dump"
	MODE_DOTHATLSASSTHING = "dothatlsassthing"
)

const (
	HANDLEMODE_DIRECT  = "direct"
	HANDLEMODE_PROCEXP = "procexp"
)

const (
	DUMPMODE_LOCAL   = "local"
	DUMPMODE_NETWORK = "network"
)

const (
	NETWORKMODE_RAW = "raw"
	NETWORKMODE_SMB = "smb"
)

var SERVICENAME = flag.String("service", "astral", "Name of the service")
var STRINGSPATH = flag.String("path", "", "Path of strings")
var DRIVERPATH = flag.String("driver", DRIVER_FULL_PATH, "Path where the driver file will be dropped")

var TARGETPID = flag.Int("pid", 0, "PID of target process (prioritized over process name)")
var TARGETPROCNAME = flag.String("name", "", "Process name of target process")

var HANDLEMODE = flag.String("handle", HANDLEMODE_DIRECT, "Method to obtain target process handle [direct|procexp]")

func main() {
	flag.Parse()
	FillArguments()
	defer CleanUp(*SERVICENAME, *DRIVERPATH, *HANDLEMODE)

	if err := ValidateArguments(); err != nil {
		LogStatus("Failed to validate arguments", err, false)
		return
	}

	if status := SetUp(*HANDLEMODE, *SERVICENAME, *DRIVERPATH); !status {
		return
	}

	*TARGETPID = GetProcessId(*TARGETPID, *TARGETPROCNAME)
	if *TARGETPID == 0 {
		LogStatus("Could not open process with PID: 0", nil, false)
		return
	}

	USER_STRINGS, err := getStrings(*STRINGSPATH)

	if len(USER_STRINGS) < 1 {
		LogStatus("No strings to find", nil, false)
	}

	if err != nil {
		LogStatus("Failed to get users strings", err, false)
	}

	SearchStringsInMemory(*TARGETPID, USER_STRINGS)

}
