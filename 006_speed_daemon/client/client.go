package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func converNum(arr []byte) int {
	n := len(arr)
	print("Arr: ")
	for idx := range arr {
		print(arr[idx], " ")
	}
	print("\n")
	ans := 0
	t := 0
	for i := n - 1; i >= 0; i-- {
		x := int(arr[i])
		b := 0
		for b < 8 {
			ans += ((x >> b) & 1) * (1 << t)
			t++
			b++
		}
	}
	print("\n")
	return ans
}

func readNum(reader *bufio.Reader, len int) int {
	var buf []byte
	len /= 8
	for len > 0 {
		len--
		b, err := reader.ReadByte()
		if err != nil {
			fmt.Print(err)
			return 0
		}
		buf = append(buf, b)
	}
	return converNum(buf)

}

func readString(reader *bufio.Reader) string {
	len := readNum(reader, 8)
	var s strings.Builder
	for i := 0; i < len; i++ {
		s.WriteRune(rune(readNum(reader, 8)))
	}
	return s.String()
}

func get_conn_mssg(conn net.Conn, connmssg chan string) {
	reader := bufio.NewReader(conn)
	for {
		id := readNum(reader, 8)
		if id == -1 {
			connmssg <- "EOF"
		}
		var s strings.Builder
		s.WriteString(strconv.Itoa(id) + " ")
		switch id {
		case 65:
			s.WriteString("Got Hearbeat ")
		case 33:
			plate := readString(reader)
			road := readNum(reader, 16)
			mile1 := readNum(reader, 16)
			t1 := readNum(reader, 32)
			mile2 := readNum(reader, 16)
			t2 := readNum(reader, 32)
			speed := readNum(reader, 16)
			s.WriteString(plate + " ")
			s.WriteString(strconv.Itoa(road) + " ")
			s.WriteString(strconv.Itoa(mile1) + " ")
			s.WriteString(strconv.Itoa(t1) + " ")
			s.WriteString(strconv.Itoa(mile2) + " ")
			s.WriteString(strconv.Itoa(t2) + " ")
			s.WriteString(strconv.Itoa(speed) + " ")
		case 10:
			mssg := readString(reader)
			s.WriteString(mssg)
		}
		connmssg <- s.String()
	}
}

func convertStringToNum(arr string) int {
	ans := 0
	cnt := 1
	for i := len(arr) - 2; i >= 0; i-- {
		ans += (int(arr[i]) - '0') * cnt
		cnt *= 10
	}
	return ans
}

func converNumToByte(num int, len int) []byte {
	len /= 8
	ans := make([]byte, len)
	currptr := 0
	print("Ans : ")
	for i := len - 1; i >= 0; i-- {
		tmp := 0
		for j := 0; j < 8; j++ {
			tmp += ((num >> currptr) & 1) * (1 << j)
			currptr++
		}
		ans[i] = byte(tmp)
	}
	for idx := range ans {
		print(ans[idx], " ")
	}
	print("\n")
	return ans
}

func get_stdin_mssg(stdinmssg chan []byte) {
	reader := bufio.NewReader(os.Stdin)
	for {
		var buffer []byte
		s, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Cannot read the message %v", err)
			stdinmssg <- []byte{}
		}

		id := convertStringToNum(s)
		switch id {
		case 64:
			buffer = append(buffer, converNumToByte(64, 8)...)
			s, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Cannot read the message %v", err)
				stdinmssg <- []byte{}
			}
			interval := convertStringToNum(s)
			print(s)
			buffer = append(buffer, converNumToByte(interval, 32)...)
			stdinmssg <- buffer
		case 128:
			buffer = append(buffer, converNumToByte(id, 8)...)
			a, _ := reader.ReadString('\n')
			id := convertStringToNum(a)
			s, _ := reader.ReadString('\n')
			mile := convertStringToNum(s)
			t, _ := reader.ReadString('\n')
			limit := convertStringToNum(t)
			buffer = append(buffer, converNumToByte(id, 16)...)
			buffer = append(buffer, converNumToByte(mile, 16)...)
			buffer = append(buffer, converNumToByte(limit, 16)...)
			stdinmssg <- buffer
		case 32:
			buffer = append(buffer, converNumToByte(id, 8)...)
			plate, _ := reader.ReadString('\n')
			timestamp, _ := reader.ReadString('\n')
			t := convertStringToNum(timestamp)
			buffer = append(buffer, []byte(plate)...)
			buffer = append(buffer, converNumToByte(t, 32)...)
			stdinmssg <- buffer
		case 129:
			//array of size 1 --> road id 8
			buffer = append(buffer, converNumToByte(id, 8)...)
			buffer = append(buffer, converNumToByte(1, 8)...)
			buffer = append(buffer, converNumToByte(8, 16)...)
			stdinmssg <- buffer
			print("Heyy dispatcher\n")
		}

	}
}

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Printf("Something went wrong %v", err)
		os.Exit(1)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	stdinmssg := make(chan []byte)
	connmssg := make(chan string)
	go get_conn_mssg(conn, connmssg)
	go get_stdin_mssg(stdinmssg)
	defer conn.Close()
	for {
		select {
		case mssg := <-stdinmssg:
			if len(mssg) == 0 {
				return
			}
			conn.Write(mssg)
		case mssg := <-connmssg:
			if mssg == "EOF" {
				return
			}
			fmt.Print("Connection mssg : ", mssg, "\n")
		}
	}
}
