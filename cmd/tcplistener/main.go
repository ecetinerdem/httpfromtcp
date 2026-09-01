package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"net"
	"os"
)

func main() {

	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error: %v", err)
			continue
		}
		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}

}

// func getLinesFromChannel(f io.ReadCloser) <-chan string {
// 	channel := make(chan string)
// 	go readingFile(channel, f)
//
// 	return channel
//
// }
//
// func readingFile(channel chan string, f io.ReadCloser) {
// 	defer close(channel)
// 	defer f.Close()
// 	b := make([]byte, 8)
//
// 	var cl string
// 	for {
// 		n, err := f.Read(b)
//
// 		if err != nil {
// 			if errors.Is(err, io.EOF) {
// 				if cl != "" {
// 					channel <- cl
// 				}
// 				return
// 			}
// 			fmt.Printf("Err: %v\n", err)
// 			return
// 		}
//
// 		parts := bytes.Split(b[:n], []byte("\n"))
//
// 		for i, part := range parts {
// 			cl += string(part)
// 			if i < len(parts)-1 {
// 				channel <- cl
// 				cl = ""
// 			}
//
// 		}
//
// 	}
// }
