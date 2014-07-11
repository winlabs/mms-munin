package components

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const DRIVE_FIXED = 3
const IOCTL_DISK_PERFORMANCE = 0x00070020

type DISK_PERFORMANCE struct {
	BytesRead int64
	BytesWritten int64
	ReadTime int64
	WriteTime int64
	IdleTime int64
	ReadCount uint32
	WriteCount uint32
	QueueDepth uint32
	SplitCount uint32
	QueryTime int64
	StorageDeviceNumber uint32
	StorageManagerName [8]uint16
}

type IOStatCountData struct {
	ReadCount uint32
	WriteCount uint32
}

type IOStatTimeData struct {
	ReadTime int64
	WriteTime int64
}

type IOStatAverageTimeData struct {
	AverageReadTime float32
	AverageWriteTime float32
}

type IOStat struct {
	mutex sync.Mutex
	volumes []string
	labels []string
	lastCounts []IOStatCountData
	countDiffs []IOStatCountData
	lastTimes []IOStatTimeData
	timeDiffs []IOStatTimeData
}

func monitorIOStat(iostat* IOStat) {
	kernelDLL := syscall.MustLoadDLL("KERNEL32.DLL")
	deviceIoControl := kernelDLL.MustFindProc("DeviceIoControl")
	ticker := time.Tick(time.Second)
	for {
		<-ticker
		iostat.mutex.Lock()
		for i := 0; i < len(iostat.volumes); i++ {
			hFile, _ := syscall.CreateFile(
				syscall.StringToUTF16Ptr("\\\\.\\" + iostat.volumes[i] + ":"),
				0,
				syscall.FILE_SHARE_READ | syscall.FILE_SHARE_WRITE,
				nil,
				syscall.OPEN_EXISTING,
				0,
				0)
			diskPerformance := DISK_PERFORMANCE{}
			diskPerformanceSize := uint32(0)
			deviceIoControl.Call(
				uintptr(hFile),
				IOCTL_DISK_PERFORMANCE,
				0,
				0,
				uintptr(unsafe.Pointer(&diskPerformance)),
				unsafe.Sizeof(diskPerformance),
				uintptr(unsafe.Pointer(&diskPerformanceSize)),
				0)
			syscall.CloseHandle(hFile)
			iostat.countDiffs[i].ReadCount = diskPerformance.ReadCount - iostat.lastCounts[i].ReadCount
			iostat.countDiffs[i].WriteCount = diskPerformance.WriteCount - iostat.lastCounts[i].WriteCount
			iostat.lastCounts[i].ReadCount = diskPerformance.ReadCount
			iostat.lastCounts[i].WriteCount = diskPerformance.WriteCount
			iostat.timeDiffs[i].ReadTime = diskPerformance.ReadTime - iostat.lastTimes[i].ReadTime
			iostat.timeDiffs[i].WriteTime = diskPerformance.WriteTime - iostat.lastTimes[i].WriteTime
			iostat.lastTimes[i].ReadTime = diskPerformance.ReadTime
			iostat.lastTimes[i].WriteTime = diskPerformance.WriteTime
		}
		iostat.mutex.Unlock()
	}
}

func NewIOStat() *IOStat {
	kernelDLL := syscall.MustLoadDLL("KERNEL32.DLL")
	getDriveType := kernelDLL.MustFindProc("GetDriveTypeW")
	getVolumeInformation := kernelDLL.MustFindProc("GetVolumeInformationW")

	possibleVolumes := []string{ "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z" }
	volumes := make([]string, 0, 26)
	labels := make([]string, 0, 26)
	for i := 0; i < 26; i++ {
		root := possibleVolumes[i] + ":\\"
		rootPtr := syscall.StringToUTF16Ptr(root)
		driveType, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(rootPtr)))
		if driveType == DRIVE_FIXED {
			volumes = append(volumes, possibleVolumes[i])
			labelBuffer := [syscall.MAX_PATH]uint16{}
			getVolumeInformation.Call(
				uintptr(unsafe.Pointer(rootPtr)),
				uintptr(unsafe.Pointer(&labelBuffer)),
				syscall.MAX_PATH,
				0,
				0,
				0,
				0,
				0)
			labels = append(labels, syscall.UTF16ToString((&labelBuffer)[:]))
		}
	}

	iostat := &IOStat{
		volumes: volumes,
		labels: labels,
		lastCounts: make([]IOStatCountData, len(volumes)),
		countDiffs: make([]IOStatCountData, len(volumes)),
		lastTimes: make([]IOStatTimeData, len(volumes)),
		timeDiffs: make([]IOStatTimeData, len(volumes)),
	}
	go monitorIOStat(iostat)
	return iostat
}

func (iostat *IOStat) GetCountData() []IOStatCountData {
	iostat.mutex.Lock()
	result := iostat.countDiffs
	iostat.mutex.Unlock()
	return result
}

func (iostat* IOStat) GetAverageTimeData() []IOStatAverageTimeData {
	iostat.mutex.Lock()
	countData := iostat.countDiffs
	timeDiffs := iostat.timeDiffs
	iostat.mutex.Unlock()
	averages := make([]IOStatAverageTimeData, len(countData))
	for i, count := range countData {
		if count.ReadCount == 0 {
			averages[i].AverageReadTime = -1
		} else {
			averages[i].AverageReadTime = float32(timeDiffs[i].ReadTime / 10000) / float32(count.ReadCount)
		}
		if count.WriteCount == 0 {
			averages[i].AverageWriteTime = -1
		} else {
			averages[i].AverageWriteTime = float32(timeDiffs[i].WriteTime / 10000) / float32(count.WriteCount)
		}
	}
	return averages
}

func (iostat *IOStat) GetLabels() []string {
	return iostat.labels
}
