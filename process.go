// +build windows

package main

import (
	"encoding/binary"
	"errors"
	"syscall"
	"unsafe"
)

// Windows API functions
var (
	modKernel32        = syscall.NewLazyDLL("kernel32.dll")
	Module32First      = modKernel32.NewProc("Module32FirstW")
	Module32Next       = modKernel32.NewProc("Module32NextW")
	readProcessMemory  = modKernel32.NewProc("ReadProcessMemory")
	writeProcessMemory = modKernel32.NewProc("WriteProcessMemory")
)

// Some constants from the Windows API
const (
	ERROR_NO_MORE_FILES = 0x12
	MAX_PATH            = 260
	MAX_MODULE_NAME32   = 255
	PROCESS_ALL_ACCESS  = 0x1F0FFF
)

type Process struct {
	PROCESSENTRY syscall.ProcessEntry32
	ModuleList   []ModuleEntry32
	Handle       syscall.Handle
}

type ProcessEntry32 syscall.ProcessEntry32

type ModuleEntry32 struct {
	Size         uint32
	ModuleID     uint32
	ProcessID    uint32
	GlobalUsage  uint32
	ProcessUsage uint32
	modBaseAddr  *uint8
	modBaseSize  uint32
	hModule      uintptr
	szModule     [MAX_MODULE_NAME32 + 1]uint16
	ExeFile      [MAX_PATH]uint16
}

func (p *Process) getPid() uint32 {
	return p.PROCESSENTRY.ProcessID
}

func (p *Process) getName() string {
	file := p.PROCESSENTRY.ExeFile
	var str string
	for _, value := range file {
		if int(value) == 0 {
			break
		}
		str += string(int(value))
	}
	return str
}

func NewFromPid(pid uint32) (*Process, error) {
	handle, err := syscall.OpenProcess(PROCESS_ALL_ACCESS, false, pid) // Opens with all rights
	if err != nil {
		return nil, err
	}
	return &Process{
		Handle: handle,
	}, nil
}

func NewFromName(name string) (*Process, error) {
	procs, err := GetProcesses()
	if err != nil {
		return nil, err
	}

	for _, proc := range procs {
		proc.getModules()
		for _, module := range proc.ModuleList {
			if module.ProcessID != 0 {
				if module.getName() == name {
					return NewFromPid(module.ProcessID)
				}
			}
		}
	}

	return nil, errors.New("process not found")
}

// func (p *Process) OpenProcess(pid uint32) error {
// 	kernel32 := syscall.MustLoadDLL("kernel32.dll")
// 	proc := kernel32.MustFindProc("OpenProcess")
// 	handle, _, _ := proc.Call(ptr(PROCESS_ALL_ACCESS), ptr(true), ptr(pid))

// 	// handle, err := syscall.OpenProcess(PROCESS_ALL_ACCESS, true, pid) // Opens with all rights
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	p.Handle = syscall.Handle(handle)
// 	return nil
// }

func (m *ModuleEntry32) getName() string {
	var str string
	for _, value := range m.szModule {
		if int(value) == 0 {
			break
		}
		str += string(int(value))
	}
	return str
}

func (m *ModuleEntry32) getFullPath() string {
	var str string
	for _, value := range m.ExeFile {
		if int(value) == 0 {
			break
		}
		str += string(int(value))
	}
	return str
}

func (p *Process) BaseAddress() uintptr {
	handle, _ := syscall.CreateToolhelp32Snapshot(
		syscall.TH32CS_SNAPMODULE,
		p.PROCESSENTRY.ProcessID)
	defer syscall.Close(handle)
	var entry ModuleEntry32

	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := Module32First.Call(uintptr(handle), uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		panic(syscall.GetLastError())
	}

	return uintptr(unsafe.Pointer(entry.modBaseAddr))
}

func (p *Process) MustWrite(address uintptr, buffer []byte) {
	err := p.Write(address, buffer)
	if err != nil {
		panic(err)
	}
}

func (p *Process) Write(address uintptr, buffer []byte) error {
	var bytesWritten uint32
	ret, _, err := writeProcessMemory.Call(uintptr(p.Handle), address, uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), uintptr(unsafe.Pointer(&bytesWritten)))
	if ret == 0 {
		return err
	}

	return nil
}

func (p *Process) Read(address uintptr, buffer []byte, size uint32) (uint32, error) {
	var bytesRead uint32

	ret, _, err := readProcessMemory.Call(uintptr(p.Handle), address, uintptr(unsafe.Pointer(&buffer[0])), uintptr(size), uintptr(unsafe.Pointer(&bytesRead)))
	if ret == 0 {
		return 0, err
	}

	return uint32(ret), nil
}

func (p *Process) ReadString(address uintptr, length uint32) (string, error) {
	buffer := make([]byte, length)
	_, err := p.Read(address, buffer, 24)
	if err != nil {
		panic(err)
	}
	return UTF16BytesToString(buffer, binary.LittleEndian), nil
}

func (p *Process) getModules() error {

	handle, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPMODULE, p.getPid())
	if err != nil {
		return err
	}
	defer syscall.Close(handle)

	var entry ModuleEntry32

	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := Module32First.Call(uintptr(handle), uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return syscall.GetLastError() //log.Panic("NO MODULES!?!¤(&/¤&3452))")
	}
	results := make([]ModuleEntry32, 128)
	for {
		results = append(results, entry)
		ret, _, _ := Module32Next.Call(uintptr(handle), uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}
	p.ModuleList = results
	return nil
}
