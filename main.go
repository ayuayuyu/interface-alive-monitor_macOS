package main

import (
	"fmt"
	"os"
	"syscall"
)

func main() {
	// カーネルとのsocket作成
	fd, err := syscall.Socket(syscall.AF_ROUTE, syscall.SOCK_RAW, 0)
	if err != nil {
		fmt.Printf("failed to create socket: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("fd (type: %T): %d \n", fd, fd)

	defer syscall.Close(fd)

	fmt.Println("ネットワーク監視の開始...")

	buf := make([]byte, 2048)
	for {
		n, err := syscall.Read(fd, buf)
		if err != nil {
			fmt.Printf("failed to read from socket: %v", err)
			continue
		}
		if n == 0 {
			continue
		}
		ParseRouteMessage(buf[:n])
	}
}
