package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func get_string(conn net.Conn, c chan string) {
	reader := bufio.NewReader(conn)
	for {
		mssg, err := reader.ReadString('\n')
		if err != nil {
			c <- "EOF"
			return
		}
		c <- mssg
	}
}

func check_if_bogo(t string) bool {
	return len(t) >= 26 && len(t) <= 35 && t[0] == '7'
}

func overwrite(mssg *string) {
	tonyAddress := "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
	ans := *mssg
	if ans[len(ans)-1] == '\n' {
		ans = ans[:len(ans)-1]
	}
	s := strings.Split(ans, " ")
	ans = ""
	for _, t := range s {
		if check_if_bogo(t) {
			t = tonyAddress
		}
		if len(ans) == 0 {
			ans = t
		} else {
			ans += " " + t
		}
	}
	ans = ans + "\n"
	fmt.Print(ans)
	*mssg = ans
	return

}

func handle_connection(user_conn net.Conn) {
	defer user_conn.Close()
	addr, err := net.ResolveTCPAddr("tcp", "chat.protohackers.com:16963")
	fmt.Printf("Got an connection from %v\n", user_conn.RemoteAddr())
	if err != nil {
		fmt.Printf("Could not resolve Address %v\n", err)
		return
	}
	server_conn, err := net.DialTCP("tcp", nil, addr)
	user_message := make(chan string)
	server_message := make(chan string)
	go get_string(user_conn, user_message)
	go get_string(server_conn, server_message)
	defer server_conn.Close()
	for {
		select {
		case mssg := <-user_message:
			if mssg == "EOF" {
				return
			}
			overwrite(&mssg)
			server_conn.Write([]byte(mssg))
		case mssg := <-server_message:
			if mssg == "EOF" {
				return
			}
			overwrite(&mssg)
			user_conn.Write([]byte(mssg))
		}
	}

}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		panic(fmt.Sprintf("Got some error %v\n", err))
	}
	fmt.Printf("Started server at %v\n", l.Addr().String())
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Not able to connect to this %v\n", err)
			os.Exit(1)
		}
		go handle_connection(conn)
	}
}
