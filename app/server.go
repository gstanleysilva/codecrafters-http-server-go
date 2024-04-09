package main

import (
	"fmt"
	"net"
	"os"
)

var Port = 4221

func main() {
	fmt.Printf("Server running on port %d", Port)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", Port))
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		// Use for to keep server running
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleRequests(conn)
	}
}

func handleRequests(conn net.Conn) {
	defer conn.Close()

	var data []byte
	_, err := conn.Read(data)
	if err != nil {
		fmt.Println("failed to read bytes: ", err.Error())
		os.Exit(1)
	}

	response := []byte("HTTP/1.1 200 OK \r\n\r\n")
	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("failed to write bytes: ", err.Error())
		os.Exit(1)
	}
}
