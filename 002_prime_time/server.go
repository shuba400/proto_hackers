package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
)

type returnObj struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func check_prime(num float64) bool {
	if num != math.Trunc(num) {
		return false
	}

	val := int64(num) 
	//Kind of hack for insanly large number that one of the test required, which due to overflow becomes -1 so adding a check below
	if(val < 2){ 
		return false
	}
	for i := int64(2); i < val; i++ {
		if i*i > val {
			break
		}
		if (val % i) == 0 {
			return false
		}
	}
	return true
}

func handle_request(conn net.Conn, buffer []byte) bool {
	var requestData interface{}
	malformed_response := []byte("Oopsie\n")
	err := json.Unmarshal(buffer, &requestData)
	if err != nil {
		fmt.Printf("Not able to unmarshall request body %v: \n", err)
		conn.Write(malformed_response)
		return false
	}
	requestMap, json_ok := requestData.(map[string]interface{})
	if json_ok == false {
		fmt.Printf("Not a valid json\n")
		conn.Write(malformed_response)
		return false
	}
	method_val, method_ok := requestMap["method"]
	number_val, number_ok := requestMap["number"]
	if method_ok == false || number_ok == false {
		fmt.Printf("Not getting required fields\n")
		conn.Write(malformed_response)
		return false
	}
	method, ok1 := method_val.(string)
	number, ok2 := number_val.(float64)
	if ok1 == false || ok2 == false {
		fmt.Printf("Fields are not of current type\n")
		conn.Write(malformed_response)
		return false
	}
	if method != "isPrime" {
		fmt.Printf("Request string is not isPrime\n")
		conn.Write(malformed_response)
		return false
	}
	ret := returnObj{Method: "isPrime", Prime: check_prime(number)}
	fmt.Printf("Number is : %f", number)
	jsonRet, err := json.Marshal(ret)
	if err != nil {
		fmt.Print(err)
	}
	jsonRet = []byte(string(jsonRet) + "\n")
	fmt.Printf(string(jsonRet))
	conn.Write(jsonRet)
	return true
}

func handle_connection(conn net.Conn, logger *log.Logger) {
	defer conn.Close()
	logger.Printf("Getting connection from at %s", conn.RemoteAddr())
	reader := bufio.NewReader(conn)
	for {
		bytes, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if handle_request(conn, bytes) == false {
			break
		}
	}
	logger.Printf("Closing connection from at %s", conn.RemoteAddr())
}

func main() {
	addr := "0.0.0.0:8080"
	logger := log.Default()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	host, port, err := net.SplitHostPort(addr)
	logger.Printf("Starrted server at %s:%s", host, port)
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handle_connection(conn, logger)
	}

}
