package main

import (
	"bufio"
	"fmt"
	"net"
)

func handleConnection(conn *net.TCPConn) {
	defer conn.Close()
	conn.Write([]byte("# munin node\r\n"))
	fmt.Printf("Accepted connection\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Printf("%s\n", scanner.Text())
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