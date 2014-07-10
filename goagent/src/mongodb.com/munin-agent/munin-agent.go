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
	fmt.Printf("Accepted connection\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Printf("%s\n", scanner.Text())
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
					for i := 0; i < len(countData); i++ {
						conn.Write([]byte(fmt.Sprintf("dev0_%d_read.value %d\r\n", i, countData[i].ReadCount)))
						conn.Write([]byte(fmt.Sprintf("dev0_%d_write.value %d\r\n", i, countData[i].WriteCount)))
					}
					conn.Write([]byte(".\r\n"))
				case "iostat_ios":
					conn.Write([]byte("dev202_1_rtime.value 0\r\n"))
					conn.Write([]byte("dev202_1_wtime.value 0\r\n"))
					conn.Write([]byte(".\r\n"))
			}
		case "config":
			conn.Write([]byte("# Unknown service\r\n"))
			conn.Write([]byte(".\r\n"))
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