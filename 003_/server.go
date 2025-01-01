package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)


func get_int(arr []byte) int {
	ans := 0
	val := 1
	for i := 3; i > 0; i--{
		ans += val*int(arr[i])
		val *= 256
	}
	ans -= val*int(arr[0])
	return ans

}
func handle_connection(conn net.Conn) {
	defer conn.Close()
	return_buffer := *&bytes.Buffer{}
	p := make([]byte,9)
	fmt.Printf("Got connection from %s\n", conn.RemoteAddr())
	tmp_buffer := make([]byte, 1024)
	for {
		new_tansaction()
		n, err := conn.Read(tmp_buffer)
		fmt.Print(tmp_buffer[:n])
		return_buffer.Write(tmp_buffer[:n])
		for return_buffer.Len() >= 9 {
			return_buffer.Read(p)
			handle_transaction(p)
		}
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

/*
73 0 0 48 57 0 0 0 101 
73 0 0 48 58 0 0 0 102 
73 0 0 48 59 0 0 0 100 
73 0 0 160 0 0 0 0 5 
81 0 0 48 0 0 0 64 0
*/
