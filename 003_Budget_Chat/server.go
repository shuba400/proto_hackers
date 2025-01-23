package main

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

type userSession struct {
	conn net.Conn
	user  string
	lock  sync.Mutex
}

type activeUserSessions struct {
	userSessions []*userSession
	lock         sync.Mutex
}

func addMessageToSession(mssg string, session *userSession) {
	defer session.lock.Unlock()
	session.lock.Lock()
	mssg = fmt.Sprintf("%s\n", mssg)
	session.conn.Write([]byte(mssg))
	return
}

func addUserToActiveSessions(sessions *activeUserSessions, currentUser *userSession) {
	defer sessions.lock.Unlock()
	sessions.lock.Lock()
	sessions.userSessions = append(sessions.userSessions, currentUser)
	mssg_for_user := "* The room contains: "
	mssg_for_other_user := fmt.Sprintf("* %s has joined the room", currentUser.user)
	for _, session := range sessions.userSessions {
		if currentUser != session {
			addMessageToSession(mssg_for_other_user, session)
			mssg_for_user = fmt.Sprintf("%s , %s", mssg_for_user, session.user)
			fmt.Print("Vlaue from here: ", mssg_for_user, "\n")
		}
	}
	addMessageToSession(mssg_for_user, currentUser)
	return
}

func removeUserFromActiveSession(sessions *activeUserSessions, currentUser *userSession) {
	defer sessions.lock.Unlock()
	sessions.lock.Lock()
	fmt.Printf("Removing user %s ", currentUser.user)
	idx_to_rem := -1
	for idx, session := range sessions.userSessions {
		if session == currentUser {
			idx_to_rem = idx
		}
	}
	if idx_to_rem == -1 {
		return
	}
	mssg := fmt.Sprintf("* %s has left the room", currentUser.user)
	sessions.userSessions = append(sessions.userSessions[:idx_to_rem], sessions.userSessions[idx_to_rem+1:]...)
	for _, session := range sessions.userSessions {
		addMessageToSession(mssg, session)
	}
	return
}

func remove_last_char(s string) string {
	return s[:len(s)-1]
}

func readConn(conn net.Conn) (string,bool) {
	buffer := bytes.Buffer{}
	tmp := make([]byte, 1024)
	for {
		n, err := conn.Read(tmp)
		if err != nil{
			return buffer.String(), true
		}
		buffer.Write(tmp[:n])
		if tmp[n-1] == '\n' {
			break
		}
	}
	return remove_last_char(buffer.String()), false
}

func broadcastMessage(mssg string, sessions *activeUserSessions, currUserSession *userSession) {
	defer sessions.lock.Unlock()
	sessions.lock.Lock()
	for _, session := range sessions.userSessions {
		if session != currUserSession {
			addMessageToSession(mssg, session)
		}
	}
	return
}

func verifyUserName(s string) bool {
	if len(s) < 1 || len(s) > 16 {
		return false
	}
	for _, c := range s {
		if !((c <= '9' && c >= '0') || (c <= 'z' && c >= 'a') || (c <= 'Z' && c >= 'A')) {
			return false
		}
	}
	return true

}

func handleUser(conn net.Conn, sessions *activeUserSessions) {
	defer conn.Close()
	fmt.Printf("Started connection from %s\n", conn.RemoteAddr())
	conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
	user_name, isdisconnect := readConn(conn)
	if isdisconnect || !verifyUserName(user_name) {
		return
	}
	currUserSession := userSession{user: user_name,conn: conn}
	addUserToActiveSessions(sessions, &currUserSession)
	defer removeUserFromActiveSession(sessions, &currUserSession)
	for {
		buf, isdisconnect := readConn(conn)
		if isdisconnect {
			return
		}
		if len(buf) != 0 && len(buf) <= 1001 {
			buf = fmt.Sprintf("[%s] %s", user_name, buf)
			broadcastMessage(buf, sessions, &currUserSession)
		}
	}

}

func main() {
	address := "0.0.0.0:8080"
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Print("Not able to start a server: \n")
		panic(err)
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		fmt.Print("Error getting port")
		panic(err)
	}
	fmt.Printf("Started server at %s,%s\n", host, port)
	sessions := activeUserSessions{}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("Not able to connect")
			panic(err)
		}
		go handleUser(conn, &sessions)
	}

}
