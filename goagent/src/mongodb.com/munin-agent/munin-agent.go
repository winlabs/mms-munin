package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func handleConnection(conn *net.TCPConn) {
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
					conn.Write([]byte("user.value 0\r\n"))
					conn.Write([]byte("nice.value 0\r\n"))
					conn.Write([]byte("system.value 0\r\n"))
					conn.Write([]byte("idle.value 0\r\n"))
					conn.Write([]byte("iowait.value 0\r\n"))
					conn.Write([]byte("irq.value 0\r\n"))
					conn.Write([]byte("softirq.value 0\r\n"))
					conn.Write([]byte("steal.value 0\r\n"))
					conn.Write([]byte("guest.value 0\r\n"))
					conn.Write([]byte(".\r\n"))
				case "iostat":
					conn.Write([]byte("dev202_0_read.value 0\r\n"))
					conn.Write([]byte("dev202_0_write.value 0\r\n"))
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
		go handleConnection(conn)
	}
}