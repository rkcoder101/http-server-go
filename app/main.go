package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func url_parser(conn *net.Conn) string {
	buffer := make([]byte, 1024)
	_, err := (*conn).Read(buffer)
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	req := string(buffer)
	req_splitted := strings.Split(req, "\r\n")
	req_line := strings.Split(req_splitted[0], " ")
	url := req_line[1]
	return url
}
func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	// we will get hit by a GET req
	// we need to write to the client, with headers and response body
	url := url_parser(&conn)
	str := url[6:]
	res:=fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",len(str),str)
	conn.Write([]byte(res))
}
