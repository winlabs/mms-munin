package components

import (
	"syscall"
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

type IOStat struct {
	volumes []string
	labels []string
}

type IOStatCountData struct {
	ReadCount uint32
	WriteCount uint32
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

	return &IOStat{ volumes: volumes, labels: labels }
}

func (iostat *IOStat) GetCountData() []IOStatCountData {
	kernelDLL := syscall.MustLoadDLL("KERNEL32.DLL")
	deviceIoControl := kernelDLL.MustFindProc("DeviceIoControl")

	data := make([]IOStatCountData, len(iostat.volumes))
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
		data[i].ReadCount = diskPerformance.ReadCount
		data[i].WriteCount = diskPerformance.WriteCount
	}
	return data
}

func (iostat *IOStat) GetLabels() []string {
	return iostat.labels
}
