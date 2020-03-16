package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

type User struct {
	Name   string
	Output chan Message
}

type Message struct {
	Username string
	Text     string
}

type ChatServer struct {
	Users map[string]User
	Join  chan User
	Leave chan User
	Input chan Message
}

func CheckMsg(s string) bool {
	for _, e := range s {
		if e != ' ' && e != '\t' && e != '\n' {
			return true
		}
	}
	return false
}

func (cs *ChatServer) Run() {
	for {
		select {
		case user := <-cs.Join:
			cs.Users[user.Name] = user
			go func() {
				cs.Input <- Message{
					Username: user.Name,
					Text:     fmt.Sprint("has joined our chat..."),
				}
			}()
		case user := <-cs.Leave:
			delete(cs.Users, user.Name)
			go func() {
				cs.Input <- Message{
					Username: user.Name,
					Text:     fmt.Sprintf("has left our chat..."),
				}
			}()
		case msg := <-cs.Input:
			for _, user := range cs.Users {
				select {
				case user.Output <- msg:
				default:
				}
			}
		}
	}
}

func handleConn(chatServer *ChatServer, conn net.Conn) {
	defer conn.Close()
	io.WriteString(conn, "Welcome to TCP-Chat!\n")

	b, Perr := ioutil.ReadFile("pingvi.txt")
	// can file be opened?
	if Perr != nil {
		fmt.Print(Perr)
		os.Exit(0)
	}

	// convert bytes to string
	myStr := string(b)
	io.WriteString(conn, myStr+"\n")
	io.WriteString(conn, "[ENTER YOUR NAME]: ")

	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	user := User{
		Name:   scanner.Text(),
		Output: make(chan Message, 10),
	}
	chatServer.Join <- user
	defer func() {
		chatServer.Leave <- user
	}()

	// Read from conn
	go func() {
		for scanner.Scan() {

			ln := scanner.Text()
			chatServer.Input <- Message{user.Name, ln}
		}
	}()

	// write to conn
	for msg := range user.Output {
		// if msg.Username != user.Name {
		if CheckMsg(msg.Text) {
			currentTime := time.Now()
			if msg.Text == "has joined our chat..." {
				_, err := io.WriteString(conn, msg.Username+" "+msg.Text+"\n")
				if err != nil {
					break
				}
			} else if msg.Text == "has left our chat..." {
				_, err := io.WriteString(conn, msg.Username+" "+msg.Text+"\n")
				if err != nil {
					break
				}
			} else {
				_, err := io.WriteString(conn, "["+currentTime.Format("2006-01-02 15:04:05")+"]"+"["+msg.Username+"]"+": "+msg.Text+"\n")
				if err != nil {
					break
				}
			}
		}
		// }
	}
}

func main() {
	server, err := net.Listen("tcp", ":8989")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	chatServer := &ChatServer{
		Users: make(map[string]User),
		Join:  make(chan User),
		Leave: make(chan User),
		Input: make(chan Message),
	}
	go chatServer.Run()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		go handleConn(chatServer, conn)
	}
}
