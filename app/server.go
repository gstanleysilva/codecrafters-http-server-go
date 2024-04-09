package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var Port = 4221

func main() {
	fmt.Printf("Server running on port %d\r\n", Port)

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

	request := parseRequest(conn)

	fmt.Println(request.Method, request.Path)

	if request.Path == "/" {
		writeResponse(conn, 200)
	} else {
		writeResponse(conn, 404)
	}

	conn.Close()
}

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
}

func parseRequest(conn net.Conn) Request {
	var result Request

	buff := make([]byte, 1024)

	size, err := conn.Read(buff)
	if err != nil {
		fmt.Println(err.Error())
	}

	strRequest := string(buff[:size])

	lines := strings.Split(strRequest, "\r\n")

	//Start Line
	parts := strings.Split(lines[0], " ")
	result.Method = parts[0]
	result.Path = parts[1]

	return result
}

func writeResponse(conn net.Conn, statusCode int) {
	_, err := conn.Write([]byte(getStringMessage(statusCode)))
	if err != nil {
		fmt.Println("failed to write bytes: ", err.Error())
		os.Exit(1)
	}
}

func statusToText(statusCode int) string {
	switch statusCode {
	case 200:
		return "OK"
	case 404:
		return "Not Found"
	default:
		return "Internal Server Error"
	}
}

func getStringMessage(statusCode int) string {
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", statusCode, statusToText(statusCode))
}
