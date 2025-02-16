package main

/*

721011081081114432119
*/

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type ThreadSafeWriter struct {
	l      sync.Mutex
	writer *bufio.Writer
}

func (threadSafeWriter *ThreadSafeWriter) Write(buff []byte) bool {
	threadSafeWriter.l.Lock()
	defer threadSafeWriter.l.Unlock()
	_, err := threadSafeWriter.writer.Write(buff)
	if err != nil {
		return true
	}
	err2 := threadSafeWriter.writer.Flush()
	if err2 != nil {
		return true
	}
	return false

}

type CarPunish struct {
	l  sync.Mutex
	db map[string][]int
}

func (carPunish *CarPunish) hasTicketOnDay(plate string, day1 int, day2 int) bool {
	carPunish.l.Lock()
	defer carPunish.l.Unlock()

	if carPunishDb.db == nil {
		carPunishDb.db = make(map[string][]int)
	}
	val, _ := carPunish.db[plate]
	for _, d := range val {
		if d == day1 || d == day2 {
			return true
		}
	}
	val = append(val, day1)
	val = append(val, day2)
	carPunish.db[plate] = val
	return false

}

type Ticket struct {
	plate      string
	id         int
	mile1      int
	timestamp1 int
	mile2      int
	timestamp2 int
	speed      int
}

type TicketList struct {
	l       sync.Mutex
	tickets []Ticket
}

func (ticketList *TicketList) sentTicket(writer *ThreadSafeWriter) bool {
	ticketList.l.Lock()
	defer ticketList.l.Unlock()

	for len(ticketList.tickets) != 0 {
		ticket := ticketList.tickets[0]
		//logic to sent this only if car has not yet recieved a ticket so far
		day1 := ticket.timestamp2 / 86400
		day2 := ticket.timestamp1 / 86400 //this is hacky / might be incorrect but whatever
		if carPunishDb.hasTicketOnDay(ticket.plate, day1, day2) {
			ticketList.tickets = ticketList.tickets[1:]
			continue
		}

		buffer := converNumToByte(33, 8)
		buffer = append(buffer, convertStrToByte(ticket.plate)...)
		buffer = append(buffer, converNumToByte(ticket.id, 16)...)
		buffer = append(buffer, converNumToByte(ticket.mile1, 16)...)
		buffer = append(buffer, converNumToByte(ticket.timestamp1, 32)...)
		buffer = append(buffer, converNumToByte(ticket.mile2, 16)...)
		buffer = append(buffer, converNumToByte(ticket.timestamp2, 32)...)
		buffer = append(buffer, converNumToByte(ticket.speed, 16)...)
		err := writeByteToConn(writer, buffer)
		if err {
			return err
		}
		ticketList.tickets = ticketList.tickets[1:]
	}
	return false
}

func (ticketList *TicketList) addTicket(ticket Ticket) {
	ticketList.l.Lock()
	defer ticketList.l.Unlock()
	ticketList.tickets = append(ticketList.tickets, ticket)
	return
}

type Road struct {
	l         sync.Mutex
	limit     int
	plates    []string
	miles     []int
	timestamp []int
}

func (r *Road) setLimit(limit int) {
	r.l.Lock()
	defer r.l.Unlock()
	r.limit = limit
	return
}

func (r *Road) addNewEntry(plate string, mile int, timestamp int, roadId int) {
	r.l.Lock()
	defer r.l.Unlock()
	r.plates = append(r.plates, plate)
	r.miles = append(r.miles, mile)
	r.timestamp = append(r.timestamp, timestamp)
	for i := range r.plates {
		if i == len(r.plates)-1 {
			break
		}
		if r.plates[i] != plate {
			continue
		}
		time_delta := math.Abs(float64(r.timestamp[i]) - float64(timestamp))
		dist_delta := (math.Abs(float64(r.miles[i]) - float64(mile)))
		speed := (dist_delta / time_delta) * 3600
		right_idx := len(r.plates) - 1
		left_idx := i
		if r.timestamp[left_idx] > r.timestamp[right_idx] {
			left_idx, right_idx = right_idx, left_idx
		}
		if speed > float64(r.limit) {
			ticket := Ticket{
				plate:      r.plates[i],
				id:         roadId,
				mile1:      r.miles[left_idx],
				timestamp1: r.timestamp[left_idx],
				mile2:      r.miles[right_idx],
				timestamp2: r.timestamp[right_idx],
				speed:      int(speed * 100),
			}
			roadTicketList[roadId].addTicket(ticket)
		}
	}
}

/*
message_id --> code,(u8)
str ---> len,(u8) len*u8
*/

func converNum(arr []byte) int {
	n := len(arr)

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
	return ans
}

func readNum(reader *bufio.Reader, len int) (int, bool) {
	var buf []byte
	len /= 8
	for len > 0 {
		len--
		b, err := reader.ReadByte()
		if err != nil {
			return 0, true
		}
		buf = append(buf, b)
	}
	return converNum(buf), false

}

func readStr(reader *bufio.Reader) (string, bool) {
	len, ret := readNum(reader, 8)
	if ret {
		return "", ret
	}
	var ans strings.Builder
	for len > 0 {
		len--
		c, ret := readNum(reader, 8)
		if ret {
			return "", ret
		}
		ans.WriteRune(rune(c))
	}
	return ans.String(), ret
}

func converNumToByte(num int, len int) []byte {
	len /= 8
	ans := make([]byte, len)
	currptr := 0
	for i := len - 1; i >= 0; i-- {
		tmp := 0
		for j := 0; j < 8; j++ {
			tmp += ((num >> currptr) & 1) * (1 << j)
			currptr++
		}
		ans[i] = byte(tmp)
	}
	return ans
}

func convertStrToByte(str string) []byte {
	var buffer []byte
	buffer = append(buffer, converNumToByte(len(str), 8)...)
	for i := 0; i < len(str); i++ {
		buffer = append(buffer, converNumToByte(int(str[i]), 8)...)
	}
	return buffer
}

func writeByteToConn(writer *ThreadSafeWriter, buff []byte) bool {
	err := writer.Write(buff)
	if err {
		return true
	}
	return false
}

/*
Error --> 10,str
Plate --> 20, timestamp(32)
Ticket --> 21, [plate,str --- road(u16) ---- mile1(u16) , timestamp1 (u32) , mile2(u16) , timestamp2(u32), speed(u16)[100 * miles_per_hour]]
WantHeartbeat 40 --> interval(u32)
HeartBeat 41 --> none
IAmCamera 80 --> [road(u16),mile(u16),limit(u16)]
IamDispatcher 81 [numroads(u8),road[][each u16]]
*/

func handleError(writer *ThreadSafeWriter, mssg string) bool {
	var buffer []byte
	buffer = append(buffer, converNumToByte(16, 8)...)
	buffer = append(buffer, convertStrToByte(mssg)...)
	return writeByteToConn(writer, buffer)
}

func handleHeartBeat(writer *ThreadSafeWriter, delay int) {
	startTime := time.Now().UnixMilli()
	delay *= 100
	buffer := converNumToByte(65, 8)
	for {
		currTime := time.Now().UnixMilli()
		if currTime-startTime >= int64(delay) {
			ret := writeByteToConn(writer, buffer)
			if ret {
				return
			}
			startTime = currTime
		}
	}
}

func dispatchTicket(writer *ThreadSafeWriter, roads []int) {
	for {
		for _, r := range roads {
			ret := roadTicketList[r].sentTicket(writer)
			if ret {
				return
			}
		}
	}
}

func handleDispatcher(reader *bufio.Reader, writer *ThreadSafeWriter) bool {
	val, err := readNum(reader, 8)
	if err {
		return err
	}
	var roads []int
	for val > 0 {
		val--
		road_id, err := readNum(reader, 16)
		if err {
			return err
		}
		roads = append(roads, road_id)
	}
	go dispatchTicket(writer, roads)
	return false
}

var carPunishDb CarPunish
var roadTicketList [65537]TicketList
var roadVehicleList [65537]Road

func handle_connection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := ThreadSafeWriter{writer: bufio.NewWriter(conn)}
	heartBeatStarted := false
	clientType := ""
	roadId := 0
	mile := 0
	limit := 0
	for {
		// Get message val here
		id, ret := readNum(reader, 8)
		if ret {
			//handle this by senting an error message
			handleError(&writer, "Can't get new connection")
			return
		}
		switch id {
		case 64:
			if heartBeatStarted {
				handleError(&writer, "Hearbeat already exists")
				return
			}
			heartBeatStarted = true
			interval, ret := readNum(reader, 32)
			if ret {
				handleError(&writer, "Connection got closed")
				return
			}
			if interval > 0 {
				go handleHeartBeat(&writer, interval)
			}
		case 128:
			if clientType != "" {
				handleError(&writer, "Connection already has a client type")
				return
			}
			//camera
			clientType = "camera"
			tmproad, err := readNum(reader, 16)
			if err {
				handleError(&writer, "Not able to read road id")
				return
			}
			tmpmile, err := readNum(reader, 16)
			if err {
				handleError(&writer, "Not able to read mile")
				return
			}
			tmplimit, err := readNum(reader, 16)
			if err {
				handleError(&writer, "Not able to read limit")
				return
			}
			roadId, mile, limit = tmproad, tmpmile, tmplimit // This is stupid, find a better way to do this
			roadVehicleList[roadId].limit = limit            //Are we good if this is not handled via mutex?

		case 32:
			//plate
			if clientType != "camera" {
				handleError(&writer, "Client is not a camera")
				return
			}
			plate, err := readStr(reader)
			if err {
				handleError(&writer, "Not able to read plate string")
				return
			}
			timestamp, err := readNum(reader, 32)
			if err {
				handleError(&writer, "Not able to read timestamp")
				return
			}
			roadVehicleList[roadId].addNewEntry(plate, mile, timestamp, roadId)

		case 129:
			if clientType != "" {
				handleError(&writer, "Connection already has a client type")
				return
			}
			// dispatcher
			clientType = "dispatcher"
			err := handleDispatcher(reader, &writer)
			if err {
				handleError(&writer, "Not able to handle dispatch")
				return
			}
		default:
			handleError(&writer, "Unknown code error\n")
			return
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
