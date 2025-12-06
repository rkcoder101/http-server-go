package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

func req_parser(conn *net.Conn) string {
	buffer := make([]byte, 1024)
	_, err := (*conn).Read(buffer)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	req := string(buffer)
	return req
}
func url_parser(req string) string {
	req_splitted := strings.Split(req, "\r\n")
	req_line := strings.Split(req_splitted[0], " ")
	url := req_line[1]
	return url
}
func userAgent_parser(req string) (string, error) {
	req_splitted := strings.SplitSeq(req, "\r\n")
	for v := range req_splitted {
		if strings.HasPrefix(v, "User-Agent") {
			return v[12:], nil
		}
	}
	return "", errors.New("User-Agent not found")
}
func handleConnection(conn *net.Conn) {
	req := req_parser(conn)
	url := url_parser(req)
	switch {
	case url == "/":
		(*conn).Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	case strings.HasPrefix(url, "/echo/"):
		(*conn).Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(url[6:]), url[6:])))
	case strings.HasPrefix(url, "/user-agent"):
		user_agent, _ := userAgent_parser(req)
		(*conn).Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(user_agent), user_agent)))
	default:
		(*conn).Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
func main() {

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
