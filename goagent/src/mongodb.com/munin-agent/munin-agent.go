package main

import (
	"bufio"
	"fmt"
	"mongodb.com/munin-agent/components"
	"net"
	"strings"
)

func handleConnection(conn *net.TCPConn, cpu *components.CPUMonitor, iostat *components.IOStat) {
	defer conn.Close()
	conn.Write([]byte("# munin node\r\n"))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		switch tokens[0] {
		case "fetch":
			switch tokens[1] {
				case "cpu":
					cpuTimes := cpu.GetCPUTimes()
					conn.Write([]byte(fmt.Sprintf("user.value %d\r\n", cpuTimes.UserTime)))
					conn.Write([]byte(fmt.Sprintf("nice.value %d\r\n", cpuTimes.NiceTime)))
					conn.Write([]byte(fmt.Sprintf("system.value %d\r\n", cpuTimes.SystemTime)))
					conn.Write([]byte(fmt.Sprintf("idle.value %d\r\n", cpuTimes.IdleTime)))
					conn.Write([]byte(fmt.Sprintf("iowait.value %d\r\n", cpuTimes.IOWaitTime)))
					conn.Write([]byte(fmt.Sprintf("irq.value %d\r\n", cpuTimes.IRQTime)))
					conn.Write([]byte(fmt.Sprintf("softirq.value %d\r\n", cpuTimes.SoftIRQTime)))
					conn.Write([]byte(fmt.Sprintf("steal.value %d\r\n", cpuTimes.StealTime)))
					conn.Write([]byte(fmt.Sprintf("guest.value %d\r\n", cpuTimes.GuestTime)))
					conn.Write([]byte(".\r\n"))
				case "iostat":
					countData := iostat.GetCountData()
					for i, count := range countData {
						conn.Write([]byte(fmt.Sprintf("dev0_%d_read.value %d\r\n", i, count.ReadCount)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_write.value %d\r\n", i, count.WriteCount)))
					}
					conn.Write([]byte(".\r\n"))
				case "iostat_ios":
					averageData := iostat.GetAverageTimeData()
					for i, average := range averageData {
						if average.AverageReadTime < 0 {
							conn.Write([]byte(fmt.Sprintf("dev0_%d_rtime.value U\r\n", i)))
						} else {
							conn.Write([]byte(fmt.Sprintf("dev0_%d_rtime.value %f\r\n", i, average.AverageReadTime)))
						}
						if average.AverageWriteTime < 0 {
							conn.Write([]byte(fmt.Sprintf("dev0_%d_wtime.value U\r\n", i)))
						} else {
							conn.Write([]byte(fmt.Sprintf("dev0_%d_wtime.value %f\r\n", i, average.AverageWriteTime)))
						}
					}
					conn.Write([]byte(".\r\n"))
			}
		case "config":
			switch tokens[1] {
				case "iostat":
					labels := iostat.GetLabels()
					conn.Write([]byte("graph_title IOstat\r\n"))
					conn.Write([]byte("graph_args --base 1024 -l 0\r\n"))
					conn.Write([]byte("graph_vlabel blocks per ${graph_period} read (-) / written (+)\r\n"))
					conn.Write([]byte("graph_category disk\r\n"))
					conn.Write([]byte("graph_info This graph shows the I/O to and from block devices.\r\n"))
					orderLine := "graph_order"
					for i := range labels {
						orderLine += fmt.Sprintf(" dev0_%d_read dev0_%d_write", i, i)
					}
					conn.Write([]byte(orderLine + "\r\n"))
					for i, label := range labels {
						conn.Write([]byte(fmt.Sprintf("dev0_%d_read.label %s\r\n", i, label)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_read.type DERIVE\r\n", i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_read.min 0\r\n", i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_read.graph no\r\n", i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_write.label %s\r\n", i, label)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_write.info I/O on device %s\r\n", i, label)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_write.type DERIVE\r\n", i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_write.min 0\r\n", i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_write.negative dev0_%d_read\r\n", i, i)))
					}
					conn.Write([]byte(".\r\n"))
				case "iostat_ios":
					labels := iostat.GetLabels()
					conn.Write([]byte("graph_title IO Service time\r\n"))
					conn.Write([]byte("graph_args --base 1000 --logarithmic\r\n"))
					conn.Write([]byte("graph_category disk\r\n"))
					conn.Write([]byte("graph_vlabel seconds\r\n"))
					conn.Write([]byte("graph_info For each applicable device this plugin shows the latency (or delay) for I/O operations on that disk device. The delay is in part made up of waiting for the disk to flush the data, and if the data arrives at the disk faster than it can read or write it then the delay time will include the time needed for waiting in the queue.\r\n"))
					for i, label := range labels {
						conn.Write([]byte(fmt.Sprintf("dev0_%d_rtime.label %s read\r\n", i, label)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_rtime.type GAUGE\r\n", i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_rtime.cdef dev0_%d_rtime,1000,/\r\n", i, i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_wtime.label %s write\r\n", i, label)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_wtime.type GAUGE\r\n", i)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_wtime.cdef dev_0_%d_wtime,1000,/\r\n", i, i)))
					}
					conn.Write([]byte(".\r\n"))
				default:
					conn.Write([]byte("# Unknown service\r\n"))
					conn.Write([]byte(".\r\n"))
			}
		case "quit":
			return
		}
	}
}

func main() {
	cpu := components.NewCPUMonitor()
	iostat := components.NewIOStat()
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 4949})
	if err != nil {
		fmt.Printf("Failed to start server!")
		return
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept connection")
			continue
		}
		go handleConnection(conn, cpu, iostat)
	}
}