package components

import (
	"github.com/winlabs/gowin32"

	"sync"
	"time"
)

type IOStatCountData struct {
	ReadCount uint
	WriteCount uint
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
	ticker := time.Tick(time.Second)
	for {
		<-ticker
		iostat.mutex.Lock()
		for i, volume := range iostat.volumes {
			info, _ := gowin32.GetDiskPerformanceInfo("\\\\.\\" + volume + ":")
			iostat.countDiffs[i].ReadCount = info.ReadCount - iostat.lastCounts[i].ReadCount
			iostat.countDiffs[i].WriteCount = info.WriteCount - iostat.lastCounts[i].WriteCount
			iostat.lastCounts[i].ReadCount = info.ReadCount
			iostat.lastCounts[i].WriteCount = info.WriteCount
			iostat.timeDiffs[i].ReadTime = info.ReadTime - iostat.lastTimes[i].ReadTime
			iostat.timeDiffs[i].WriteTime = info.WriteTime - iostat.lastTimes[i].WriteTime
			iostat.lastTimes[i].ReadTime = info.ReadTime
			iostat.lastTimes[i].WriteTime = info.WriteTime
		}
		iostat.mutex.Unlock()
	}
}

func NewIOStat() *IOStat {
	possibleVolumes := []string{ "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z" }
	volumes := make([]string, 0, 26)
	labels := make([]string, 0, 26)
	for _, candidate := range possibleVolumes {
		root := candidate + ":\\"
		driveType := gowin32.GetVolumeDriveType(root)
		if driveType == gowin32.DriveFixed {
			volumes = append(volumes, candidate)
			info, _ := gowin32.GetVolumeInfo(root)
			if len(info.VolumeName) > 0 {
				labels = append(labels, info.VolumeName)
			} else {
				labels = append(labels, candidate + "Drive")
			}
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
