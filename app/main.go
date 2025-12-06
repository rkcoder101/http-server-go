package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var file_directory string

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
func method_parser(req string) string {
	req_splitted := strings.Split(req, "\r\n")
	req_line := strings.Split(req_splitted[0], " ")
	method := req_line[0]
	return method
}
func req_body_parser(req string) string{
	req_splitted := strings.Split(req, "\r\n")
	return req_splitted[len(req_splitted)-1]
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
	path:=file_directory+file
	// write file at given path with content of req_body
	fmt.Println(path)
	_,err:=os.Create(path)
	if err!=nil{
		fmt.Println("Error creating the file")
		return err
	}
	data:=[]byte(req_body)
	err=os.WriteFile(path,data,0644)
	if err!=nil{
		fmt.Println("Error writing to the file")
		return err
	}	
	return nil
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
	case strings.HasPrefix(url, "/files/"):
		method := method_parser(req)
		fmt.Println(method)
		switch {
		case method == "GET":
			file, err := serve_file(url[7:])
			if err != nil {
				(*conn).Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			} else {
				(*conn).Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(file), file)))
			}
		case method == "POST":
			req_body:=req_body_parser(req)
			fmt.Println(req_body)
			_=write_file(req_body,url[7:])
			(*conn).Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}

	default:
		(*conn).Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
func main() {

	flag.StringVar(&file_directory, "directory", "", "boingboing")
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
