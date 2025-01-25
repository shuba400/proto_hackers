package main

import (
	"bufio"
	"net"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	addr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:8080")
	for {
		text, _ := reader.ReadString('\n')
		conn, _ := net.DialUDP("udp", nil, addr)
		// Send a message to the server
		text = text[:len(text)-1]
		_, _ = conn.Write([]byte(text))
		tmp := make([]byte, 1024)
		n, _ := conn.Read(tmp)
		tmp = tmp[:n]
		print(string(tmp))
		print("\n")
	}
}
