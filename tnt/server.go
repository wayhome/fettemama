package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	port = 1337
)

type TelnetServer struct {
	formatter BlogFormatter

	control_chan chan string
	status_chan  chan string

	sessionsMutex   sync.RWMutex
	sessions        map[int]*BlogSession
	last_session_id int
}

//// server
func NewTelnetServer(formatter BlogFormatter) *TelnetServer {
	ts := &TelnetServer{
		formatter: formatter,
	}

	ts.control_chan = make(chan string)
	ts.status_chan = make(chan string)
	ts.sessions = make(map[int]*BlogSession)

	return ts
}

func (srv *TelnetServer) Run() {
	go srv.serverFunc()

	for {
		select {
		case command := <-srv.control_chan:
			if command == "diediedie" {
				return
			}
		case status := <-srv.status_chan:
			log.Println(status)
		}
	}
}

func (srv *TelnetServer) Shutdown() {
	srv.control_chan <- "diediedie"
}

func (srv *TelnetServer) PostStatus(status string) {
	srv.status_chan <- status
}

func bcast(ses *BlogSession, message string) {
	ses.Send("\n*** Broadcast: " + message)
	ses.SendPrompt()
}

//broadcasts message to all connected clients
func (srv *TelnetServer) Broadcast(message string) {
	for _, s := range srv.sessions {
		go bcast(s, message)
	}
}

func (srv *TelnetServer) GetUserCount() int {
	srv.sessionsMutex.RLock()
	defer srv.sessionsMutex.RUnlock()

	return len(srv.sessions)
}

func (srv *TelnetServer) registerSession(session *BlogSession) {
	srv.sessionsMutex.Lock()
	srv.last_session_id++
	session.SetId(srv.last_session_id)
	srv.sessions[srv.last_session_id] = session
	srv.sessionsMutex.Unlock()
}

func (srv *TelnetServer) unregisterSession(session *BlogSession) {
	srv.sessionsMutex.Lock()
	id := session.Id()
	delete(srv.sessions, id)
	srv.sessionsMutex.Unlock()
}

//spawn new blog session for client
func (srv *TelnetServer) handleClient(conn net.Conn) {
	defer conn.Close()

	session := NewBlogSession(srv, conn)
	go session.connReader()
	go session.connWriter()
	go session.inputProcessor()

	srv.PostStatus("* [" + (session.conn).RemoteAddr().String() + "] new connection")
	srv.registerSession(session)

	session.SendVersion()
	s := fmt.Sprintf("There are %d users active.\n", srv.GetUserCount())
	session.Send(s)
	session.SendPrompt()

	session.Run()
	srv.PostStatus("* [" + (session.conn).RemoteAddr().String() + "] disconnected")
	srv.unregisterSession(session)
}

func (srv *TelnetServer) serverFunc() {
	service := fmt.Sprintf(":%d", port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", service)
	listener, _ := net.ListenTCP("tcp4", tcpAddr)

	srv.PostStatus("listening on: " + tcpAddr.IP.String() + service)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go srv.handleClient(conn)
	}
}
