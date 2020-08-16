// +build windows

package main

import (
	"syscall"
	"unsafe"
)

func GetProcesses() ([]Process, error) {
	list := make([]Process, 0)

	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(snapshot)

	var procEntry syscall.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))
	if err = syscall.Process32First(snapshot, &procEntry); err != nil {
		return nil, err
	}

	results := make([]syscall.ProcessEntry32, 0, 50)
	for {
		results = append(results, procEntry)
		if err := syscall.Process32Next(snapshot, &procEntry); err != nil {
			break
		}
	}
	//fmt.Println(results2)
	var p Process
	for _, v := range results {
		if v.ProcessID != 0 {
			p.PROCESSENTRY = v
			list = append(list, p)
		}
	}
	return list, nil
}
