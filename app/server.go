package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	Port     = 4221
	FilePath = ""
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

func NewRequest() Request {
	return Request{
		Headers: make(map[string]string),
	}
}

type Response struct {
	Status  int
	Body    []byte
	Headers map[string]string
}

func NewResponse(status int) Response {
	return Response{
		Status:  status,
		Headers: make(map[string]string, 0),
	}
}

func main() {
	fmt.Printf("Server running on port %d\r\n", Port)

	dir := flag.String("directory", "", "directory for http server files")
	flag.Parse()

	if dir != nil {
		FilePath = *dir
	}

	// FilePath = "../"

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

	resource, params := parsePath(request.Path)

	switch {
	case resource == "/" || request.Path == "":
		writeResponse(conn, NewResponse(200))
	case resource == "/echo":
		response := NewResponse(200)
		response.Headers["Content-Type"] = "text/plain"
		response.Headers["Content-Length"] = strconv.Itoa(len(params))
		response.Body = []byte(params)
		writeResponse(conn, response)
	case resource == "/user-agent":
		response := NewResponse(200)

		//get user-agent from request and add to response body
		agent, ok := request.Headers["User-Agent"]
		if ok {
			response.Body = []byte(agent)
		}

		//add response headers
		response.Headers["Content-Type"] = "text/plain"
		response.Headers["Content-Length"] = strconv.Itoa(len(response.Body))

		writeResponse(conn, response)
	case resource == "/files" && request.Method == "POST":
		if params == "" || FilePath == "" {
			writeResponse(conn, NewResponse(404))
			break
		}

		err := os.WriteFile(fmt.Sprintf("%s%s", FilePath, params), []byte(request.Body), 0644)
		if err != nil {
			writeResponse(conn, NewResponse(404))
			break
		}

		writeResponse(conn, NewResponse(201))
	case resource == "/files":
		if params == "" || FilePath == "" {
			writeResponse(conn, NewResponse(404))
			break
		}

		content, err := os.ReadFile(fmt.Sprintf("%s%s", FilePath, params))
		if err != nil {
			writeResponse(conn, NewResponse(404))
			break
		}

		//create response
		response := NewResponse(200)
		response.Body = content
		response.Headers["Content-Type"] = "application/octet-stream"
		response.Headers["Content-Length"] = strconv.Itoa(len(response.Body))

		writeResponse(conn, response)
	default:
		writeResponse(conn, NewResponse(404))
	}

	conn.Close()
}

func parsePath(path string) (resource string, params string) {
	resource = path
	parts := strings.Split(path, "/")

	//base path
	if len(parts) <= 2 {
		return resource, params
	}

	//extract params from path
	resource = fmt.Sprintf("/%s", parts[1])
	for _, part := range parts[2:] {
		if len(params) == 0 {
			params = part
			continue
		}
		params = fmt.Sprintf("%s/%s", params, part)
	}

	return resource, params
}

func parseRequest(conn net.Conn) Request {
	var headerEnd int
	result := NewRequest()

	//read data from connection
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

	//Read Headers
	for idx, line := range lines[1:] {
		headerEnd = idx
		if !strings.Contains(line, ": ") {
			break
		}
		parts := strings.Split(line, ": ")
		if _, ok := result.Headers[parts[0]]; !ok {
			result.Headers[parts[0]] = parts[1]
		}
	}

	//Read Body
	bodyStart := headerEnd + 2
	for _, line := range lines[bodyStart:] {
		result.Body = fmt.Sprintf("%s%s", result.Body, line)
	}

	return result
}

func writeResponse(conn net.Conn, response Response) {
	//StatusLine
	conn.Write([]byte(getStringMessage(response.Status)))

	//Headers
	if len(response.Headers) != 0 {
		for key, value := range response.Headers {
			conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		}
	}
	conn.Write([]byte("\r\n"))

	//Body
	if len(response.Body) != 0 {
		conn.Write([]byte(response.Body))
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

func getStringMessage(status int) string {
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n", status, statusToText(status))
}
