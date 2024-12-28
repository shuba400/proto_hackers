package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

func handle_connection(conn net.Conn) {
	defer conn.Close()
	return_buffer := *&bytes.Buffer{}
	fmt.Printf("Got connection from %s\n", conn.RemoteAddr())
	tmp_buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(tmp_buffer)
		return_buffer.Write(tmp_buffer[:n])
		// We should not check for network packet size since they are quite inconsistent 
		// and it is not a gurantee that if packet size is less than buffer size then client 
		// has stopped senting request 
		if err == io.EOF {  
			break
		}
	}
	fmt.Printf("We are senting back data to %s - %d\n", conn.RemoteAddr(), return_buffer.Len())
	conn.Write(return_buffer.Bytes())
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
	host, port, err := net.SplitHostPort(lis.Addr().String())
	fmt.Printf("Server Started at: %s:%s\n", host, port)
	idx := 0
	for {
		conn, err := lis.Accept()
		if err != nil {
			panic(err)
		}
		go handle_connection(conn)
		idx++
	}
}
