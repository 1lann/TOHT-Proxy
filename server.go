package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"runtime/debug"
)

const port = "9002"
const target = ""

type client struct {
	listenChannel        chan bool // Channel that the client is listening on
	transmitChannel      chan bool // Channel that the client is writing to
	listener             io.Writer // The thing to write to
	listenerConnected    bool
	transmitter          io.Reader // The thing to listen from
	transmitterConnected bool
}

var connectedClients map[string]client = make(map[string]client)

func bindServer(clientId string) {
	if connectedClients[clientId].listenerConnected && connectedClients[clientId].transmitterConnected {
		log.Println("Two-way connection to client established!")
		log.Println("Client <=|F|=> Proxy <-...-> VPN")

		defer func() {
			connectedClients[clientId].listenChannel <- true
			connectedClients[clientId].transmitChannel <- true
		}()

		serverConn, err := net.Dial("tcp", target)
		if err != nil {
			log.Println("Failed to connect to remote server :/", err)
		}

		defer serverConn.Close()

		wait := make(chan bool)

		go func() {
			_, err := io.Copy(connectedClients[clientId].listener, serverConn)
			if err != nil {
				log.Println("Disconnect:", err)
			}

			wait <- true
		}()

		go func() {
			_, err := io.Copy(serverConn, connectedClients[clientId].transmitter)
			if err != nil {
				log.Println("Disconnect:", err)
			}

			wait <- true
		}()

		log.Println("Full connection established!")
		log.Println("Client <=|F|=> Proxy <---> VPN")

		<-wait
		log.Println("Connection closed")
	}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Connection panic:", r)
			debug.PrintStack()
		}
	}()

	reader := bufio.NewReader(clientConn)

	line, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Failed to read first line")
		return
	}
	if line == "GET /listen HTTP/1.0\r\n" {
		// This is for LISTENING
		resolvedId := ""
		for line, err = reader.ReadString('\n'); true; line, err = reader.ReadString('\n') {
			if err != nil {
				log.Println("Failed to read following lines")
				return
			}

			if len(line) > 10 && line[:10] == "ClientId: " {
				resolvedId = line[10:30]
			}

			if line == "\r\n" {
				break
			}
		}

		if len(resolvedId) > 1 {
			fmt.Fprintf(clientConn, "HTTP/1.0 200 OK\r\n")
			fmt.Fprintf(clientConn, "Content-Type: application/octet-stream\r\n")
			fmt.Fprintf(clientConn, "Connection: keep-alive\r\n")
			fmt.Fprintf(clientConn, "Content-Length: 12345789000\r\n\r\n")

			wait := make(chan bool)

			if _, ok := connectedClients[resolvedId]; !ok {
				connectedClients[resolvedId] = client{}
			}

			currentClient := connectedClients[resolvedId]

			currentClient.listener = clientConn
			currentClient.listenChannel = wait
			currentClient.listenerConnected = true

			connectedClients[resolvedId] = currentClient

			log.Println("Attempting to bind listener")

			go bindServer(resolvedId)

			<-wait
		} else {
			log.Println("Failed to find client id!")
		}

	} else if line == "POST /transmit HTTP/1.0\r\n" {
		// This is for TRANSMITTING
		resolvedId := ""
		for line, err = reader.ReadString('\n'); true; line, err = reader.ReadString('\n') {
			if err != nil {
				log.Println("Failed to read following lines")
				return
			}

			if len(line) > 10 && line[:10] == "ClientId: " {
				resolvedId = line[10:30]
			}

			if line == "----------SWAG------BOUNDARY----\r\n" {
				break
			}
		}

		if len(resolvedId) > 1 {
			wait := make(chan bool)

			if _, ok := connectedClients[resolvedId]; !ok {
				connectedClients[resolvedId] = client{}
			}

			currentClient := connectedClients[resolvedId]

			currentClient.transmitter = reader
			currentClient.transmitChannel = wait
			currentClient.transmitterConnected = true

			connectedClients[resolvedId] = currentClient

			log.Println("Attempting to bind transmission")

			go bindServer(resolvedId)

			<-wait
		} else {
			log.Println("Failed to find client id!")
		}

	} else {
		fmt.Fprintf(clientConn, "HTTP/1.0 404 Not found\r\n")
		fmt.Fprintf(clientConn, "Content-Type: text/plain\r\n")
		fmt.Fprintf(clientConn, "Content-Length: 8\r\n\r\n")
		fmt.Fprintf(clientConn, "u wot m9")
	}
}

func main() {
	log.Println("Listening...")
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println("Error listening!", err)
		return
	}

	for true {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
			continue
		}

		go handleConnection(conn)
	}
}
