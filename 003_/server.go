package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net"
)

type transaction struct {
	time []int
	cost []int
}

func get_int(arr []byte) int {
	ans := 0
	tmp := 0
	ptr := 0
	for i := 3; i >= 0; i--{
		for j := 0; j < 8; j++{
			tmp = ((int(arr[i])>>j)&1)
			if ptr == 31 {
				ans -= (tmp * (1<<ptr))
			} else { 
				ans += (tmp * (1<<ptr))
			}
			ptr++
		}
	}
	return ans

}

func get_byte(num int) []byte {
	res := make([]byte, 4)
	ptr := 31
	tmp := 0
	for i := 0; i < 4; i++{
		tmp = 0
		for j := 7; j >= 0; j--{
			if(ptr == 31){
				if(num < 0){
					num += (1<<ptr)
					tmp += 1<<j
				} 
			} else {
				if (num>>ptr)&1 == 1 { 
					tmp += (1<<j)
				}
			}
			ptr--
		}
		res[i] = byte(tmp)
	}
	return res
}

func process_query(res []byte, conn net.Conn, trans *transaction) {
	a := get_int(res[1:5])
	b := get_int(res[5:9])
	q := int(res[0])
	if q == 73 {
		trans.time = append(trans.time, a)
		trans.cost = append(trans.cost, b)
		return
	}
	tot := 0.0
	val := 0.0
	for i := 0; i < len(trans.time); i++ {
		if trans.time[i] >= a && trans.time[i] <= b {
			val++
		}
	}
	for i := 0; i < len(trans.time); i++ {
		if trans.time[i] >= a && trans.time[i] <= b {
			tot += float64(trans.cost[i])/val
		}
	}
	mean := int(math.Trunc(tot))
	fmt.Printf("Mean is %d",mean)
	conn.Write(get_byte(mean))
	return
}

func handle_connection(conn net.Conn) {
	defer conn.Close()
	return_buffer := *&bytes.Buffer{}
	p := make([]byte, 9)
	fmt.Printf("Got connection from %s\n", conn.RemoteAddr())
	tmp_buffer := make([]byte, 1024)
	trans := new(transaction)
	for {
		n, err := conn.Read(tmp_buffer)
		return_buffer.Write(tmp_buffer[:n])
		for return_buffer.Len() >= 9 {
			return_buffer.Read(p)
			process_query(p, conn, trans)
		}
		// We should not check for network packet size since they are quite inconsistent
		// and it is not a gurantee that if packet size is less than buffer size then client
		// has stopped senting request
		if err == io.EOF {
			break
		}
	}
	fmt.Printf("Closing connection %d", conn.RemoteAddr())
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
