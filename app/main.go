package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var file_directory string

type Request struct {
	url, method, body string
	headers           map[string][]string
}

func serve_file(file string) (string, error) {
	path := file_directory + file
	dat, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}
func write_file(req_body string, file string) error {
	path := filepath.Join(file_directory, file)
	fmt.Println(path)
	return os.WriteFile(path, []byte(req_body), 0644)
}
func parseRequest(conn *net.Conn) (Request, error) {
	buffer := make([]byte, 1024)
	n, err := (*conn).Read(buffer)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	raw := string(buffer[:n])
	raw_splitted := strings.Split(raw, "\r\n")
	req_line := strings.Split(raw_splitted[0], " ")
	// method
	method := req_line[0]
	// url
	url := req_line[1]
	// headers
	headers := make(map[string][]string)
	i := 1
	for ; i < len(raw_splitted); i++ {
		if raw_splitted[i] == "" {
			break
		}
		x := strings.SplitN(raw_splitted[i], ": ", 2) // Header: option1, option2, option3....
		if len(x) >= 2 {
			y := strings.SplitSeq(x[1], ", ")
			for v := range y {
				headers[x[0]] = append(headers[x[0]], v)
			}
		}
	}
	// body
	body := strings.Join(raw_splitted[i+1:], "\r\n")
	return Request{
		url:     url,
		method:  method,
		headers: headers,
		body:    body,
	}, nil

}
func handleConnection(conn *net.Conn) {
	defer (*conn).Close()
	req, err := parseRequest(conn)
	if err != nil {
		fmt.Println("Error in parsing request")
		return
	}
	switch {
	case req.url == "/":
		(*conn).Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case strings.HasPrefix(req.url, "/echo/"):
		req.body = req.url[6:]
		gzip_present:=false
		for _,v :=range req.headers["Accept-Encoding"]{
			if (v=="gzip"){
				gzip_present=true
				break
			}
		}
		if  gzip_present{
			(*conn).Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(req.body), req.body)))
		} else {
			(*conn).Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(req.body), req.body)))			
		}

	case strings.HasPrefix(req.url, "/user-agent"):
		user_agent := req.headers["User-Agent"]
		(*conn).Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(user_agent), user_agent)))
	case strings.HasPrefix(req.url, "/files/"):
		switch {
		case req.method == "GET":
			file, err := serve_file(req.url[7:])
			if err != nil {
				(*conn).Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			} else {
				(*conn).Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(file), file)))
			}
		case req.method == "POST":
			_ = write_file(req.body, req.url[7:])
			(*conn).Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}

	default:
		(*conn).Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
func main() {

	flag.StringVar(&file_directory, "directory", "", "directory for recieving/writing files")
	flag.Parse()
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Listening on port 4221")
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(&conn)
	}
}
