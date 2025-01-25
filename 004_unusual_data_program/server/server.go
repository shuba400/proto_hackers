package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func process(tmp []byte, db map[string]string) string {
	s := string(tmp)
	fmt.Print(s)
	fmt.Print('\n')
	var key strings.Builder
	var value strings.Builder
	done := 0
	for _, c := range s {
		if done == 1 {
			value.WriteRune(c)
		} else {
			if c == '=' {
				done = 1
			} else {
				key.WriteRune(c)
			}
		}
	}
	if done == 1 {
		if key.String() != "version" {
			db[key.String()] = value.String()
		}
	} else {
		value, _ := db[key.String()]
		return fmt.Sprintf("%s=%s", key.String(), value)
	}
	return os.DevNull //Def something that I is a hacky workaround for this challenge
}

func main() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"), // Explicitly use 0.0.0.0 to bind to all interfaces
		Port: 8080,
	})
	if err != nil {
		fmt.Printf("Not able to start a udp server %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Listening at add %s:%d\n", "0.0.0.0", 8080)
	db := make(map[string]string)
	db["version"] = "Key Val 1.0"
	for {
		tmp := make([]byte, 1024)
		n, udp_arr, err := conn.ReadFromUDP(tmp)
		if err != nil {
			fmt.Printf("Encountered an error : %v\n", err)
		}
		fmt.Printf("Got a connection from Address: %s:%d\n", udp_arr.IP, udp_arr.Port)
		s := process(tmp[:n], db)
		if s != os.DevNull {
			conn.WriteToUDP([]byte(s), udp_arr)
		}

	}
}
