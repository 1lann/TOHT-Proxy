package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
)

const proxyDomain = "127.0.0.1:9002"
const port = "8080"

var letters = []rune("abcdefghijklmnopqrstuvwyz1234567890")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	serverSend, err := net.Dial("tcp", proxyDomain)
	if err != nil {
		log.Println("Failed to connect to send proxy server!")
		return
	}

	serverListen, err := net.Dial("tcp", proxyDomain)
	if err != nil {
		log.Println("Failed to connect to listen proxy server!")
		return
	}

	defer serverListen.Close()
	defer serverSend.Close()

	clientId := randSeq(20)

	wait := make(chan bool)

	go func() {
		fmt.Fprintf(serverSend, "POST /transmit HTTP/1.1\r\n")
		fmt.Fprintf(serverSend, "Host: "+proxyDomain+"\r\n")
		fmt.Fprintf(serverSend, "Accept: */*\r\n")
		fmt.Fprintf(serverSend, "Connection: keep-alive\r\n")
		fmt.Fprintf(serverSend, "Clientid: "+clientId+"\r\n")
		fmt.Fprintf(serverSend,
			"Content-Type: multipart/form-data; boundary=----------SWAG------BOUNDARY----\r\n")
		// fmt.Fprintf(serverSend, "Transfer-Encoding: chunked\r\n")
		fmt.Fprintf(serverSend, "Content-Length: 12345789000\r\n\r\n")
		fmt.Fprintf(serverSend, "----------SWAG------BOUNDARY----\r\n")

		_, err = io.Copy(serverSend, clientConn)
		if err != nil {
			log.Println("Error copying client to server stream", err)
		}

		wait <- true
	}()

	go func() {
		fmt.Fprintf(serverListen, "GET /listen HTTP/1.1\r\n")
		fmt.Fprintf(serverListen, "Host: "+proxyDomain+"\r\n")
		fmt.Fprintf(serverListen, "Accept: */*\r\n")
		fmt.Fprintf(serverListen, "Clientid: "+clientId+"\r\n")
		fmt.Fprintf(serverListen, "Connection: keep-alive\r\n")
		fmt.Fprintf(serverListen, "\r\n")

		buf := bufio.NewReader(serverListen)

		success := false

		for line, err := buf.ReadString('\n'); true; line, err = buf.ReadString('\n') {
			if err != nil {
				log.Println("Failed to read following lines")
				return
			}

			if line == "HTTP/1.1 200 OK\r\n" {
				success = true
			}

			if success && line == "\r\n" {
				break
			}
		}

		if success {
			_, err = io.Copy(clientConn, buf)

			if err != nil {
				log.Println("Error copying server to client stream", err)
			}
		} else {
			log.Println("Failed to bind listen connection!")
		}

		wait <- true

	}()

	<-wait
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
