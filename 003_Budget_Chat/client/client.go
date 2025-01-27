package main

/*
Moving away from chat gpt and finally understanding pipes a bit
*/

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func get_conn_mssg(conn net.Conn,connmssg chan string){
	reader := bufio.NewReader(conn)
	for{
		s,err := reader.ReadString('\n') 
		if(err != nil){
			fmt.Printf("Cannot read the message %v",err)
		}
		connmssg <- s
	}
}

func get_stdin_mssg(stdinmssg chan string){
	reader := bufio.NewReader(os.Stdin)
	for{
		s,err := reader.ReadString('\n') 
		if(err != nil){
			fmt.Printf("Cannot read the message %v",err)
		}
		stdinmssg <- s
	}
}

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Printf("Something went wrong %v", err)
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	stdinmssg := make(chan string)
	connmssg := make(chan string)
	go get_conn_mssg(conn, connmssg)
	go get_stdin_mssg(stdinmssg)
	for {
		select {
		case mssg := <-stdinmssg:
			conn.Write([]byte(mssg))
		case mssg := <-connmssg:
			fmt.Print(mssg)
		}
	}

}
