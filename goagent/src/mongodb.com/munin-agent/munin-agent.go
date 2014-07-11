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
				case "cpu":
					lines := []string{
						"graph_title CPU usage",
						"graph_order system user nice idle iowait irq softirq",
						"graph_args --base 1000 -r --lower-limit 0 --upper-limit 100",
						"graph_vlabel %",
						"graph_scale no",
						"graph_info This graph shows how CPU time is spent.",
						"graph_category system",
						"graph_period second",
						"system.label system",
						"system.draw AREA",
						"system.min 0",
						"system.type DERIVE",
						"system.info CPU time spent by the kernel in system activities",
						"user.label user",
						"user.draw STACK",
						"user.min 0",
						"user.type DERIVE",
						"user.info CPU time spent by normal programs and daemons",
						"nice.label nice",
						"nice.draw STACK",
						"nice.min 0",
						"nice.type DERIVE",
						"nice.info CPU time spent by nice(1)d programs",
						"idle.label idle",
						"idle.draw STACK",
						"idle.min 0",
						"idle.type DERIVE",
						"idle.info Idle CPU time",
						"iowait.label iowait",
						"iowait.draw STACK",
						"iowait.min 0",
						"iowait.type DERIVE",
						"iowait.info CPU time spent waiting for I/O operations to finish when there is nothing else to do.",
						"irq.label irq",
						"irq.draw STACK",
						"irq.min 0",
						"irq.type DERIVE",
						"irq.info CPU time spent handling interrupts",
						"softirq.label softirq",
						"softirq.draw STACK",
						"softirq.min 0",
						"softirq.type DERIVE",
						"softirq.info CPU time spent handling \"batched\" interrupts",
						"steal.label steal",
						"steal.draw STACK",
						"steal.min 0",
						"steal.type DERIVE",
						"steal.info The time that a virtual CPU had runnable tasks, but the virtual CPU itself was not running",
						"guest.label guest",
						"guest.draw STACK",
						"guest.min 0",
						"guest.type DERIVE",
						"guest.info The time spent running a virtual CPU for guest operating systems under the control of the Linux kernel.",
						".",
					}
					for _, line := range lines {
						conn.Write([]byte(line + "\r\n"))
					}
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