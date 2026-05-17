package main

import (
	"log"
	"net"
	"time"
)

func main() {
	// create server port
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		// wait for client to connect and establish socket
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(conn)
		// create new go routine => new thread to handle
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	log.Println(conn.RemoteAddr())
	// read data from client
	var buffer []byte = make([]byte, 1000)

	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	// simulating process
	time.Sleep(time.Second * 10)

	// reply
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nHello, world\r\n"))

	// close connection
	conn.Close()
}

func handleMultipleRequest(conn net.Conn) {
	defer conn.Close()
	log.Println(conn.RemoteAddr())

	// read data from client
	var buffer []byte = make([]byte, 1000)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Fatal(err)
		}
		request := string(buffer[:n])
		log.Println("request:", request)
		// simulating process
		time.Sleep(time.Second * 10)

		// reply
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nHello, world\r\n"))
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}
