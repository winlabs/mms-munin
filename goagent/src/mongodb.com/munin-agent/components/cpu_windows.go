package components

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type CPUMonitor struct {
	mutex sync.Mutex
	lastIdleTime int64
	lastKernelTime int64
	lastUserTime int64
	lastIdleTimeDiff int64
	lastKernelTimeDiff int64
	lastUserTimeDiff int64
}

type CPUTimes struct {
	UserTime int64
	NiceTime int64
	SystemTime int64
	IdleTime int64
	IOWaitTime int64
	IRQTime int64
	SoftIRQTime int64
	StealTime int64
	GuestTime int64
}

func monitorCPUTime(cpu *CPUMonitor) {
	kernelDLL := syscall.MustLoadDLL("KERNEL32.DLL")
	getSystemTimes := kernelDLL.MustFindProc("GetSystemTimes")
	ticker := time.Tick(time.Second)
	for {
		<-ticker
		idleTime := int64(0)
		kernelTime := int64(0)
		userTime := int64(0)
		getSystemTimes.Call(
			uintptr(unsafe.Pointer(&idleTime)),
			uintptr(unsafe.Pointer(&kernelTime)),
			uintptr(unsafe.Pointer(&userTime)))
		cpu.mutex.Lock()
		cpu.lastIdleTimeDiff = idleTime - cpu.lastIdleTime
		cpu.lastKernelTimeDiff = kernelTime - cpu.lastKernelTime
		cpu.lastUserTimeDiff = userTime - cpu.lastUserTime
		cpu.lastIdleTime = idleTime
		cpu.lastKernelTime = kernelTime
		cpu.lastUserTime = userTime
		cpu.mutex.Unlock()
	}
}

func NewCPUMonitor() *CPUMonitor {
	cpu := &CPUMonitor{}
	go monitorCPUTime(cpu)
	return cpu
}

func (cpu *CPUMonitor) GetCPUTimes() CPUTimes {
	cpu.mutex.Lock()
	result := CPUTimes{
		UserTime: cpu.lastUserTimeDiff,
		SystemTime: cpu.lastKernelTimeDiff,
		IdleTime: cpu.lastIdleTimeDiff,
	}
	cpu.mutex.Unlock()
	totalTime := result.UserTime + result.SystemTime + result.IdleTime
	if totalTime != 0 {
		result.UserTime = result.UserTime * 10000 / totalTime
		result.SystemTime = result.SystemTime * 10000 / totalTime
		result.IdleTime = result.IdleTime * 10000 / totalTime
	}
	return result
}
