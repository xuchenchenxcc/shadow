package main

import (
"net"
"log"
"io"
"strings"
"fmt"
)

//流量牵引--xcc
func main() {
	// log.SetFlags(log.LstdFlags|log.Lshortfile)
	l, err := net.Listen("tcp", ":18081")

	var num *int
	cache!!!!

	if err != nil {
		log.Panic(err)
	}
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		//fmt.Printf("%s\n",client)
		go handle_request(client, *num, cache)
	}
}

func handle_request(client net.Conn, num *int, cache!!!!) {
	address := "192.168.1.1:80"
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("-----HHHHH-----:%s\n", err)
		return
	}
	shadow_address := "192.168.2.1:80"
	shadow_server, err := net.Dial("tcp", shadow_address)
	if err != nil {
		log.Printf("-----HHHHH-----:%s\n", err)
		return
	}
	var buffer [1024]byte
	for { 
		n, err := client.Read(buffer[:])
		if err != nil {
			log.Printf("-----HHHHH-----:%s\n", err)
			return
		}
		write_to_cache()
	}

	//go io.Copy(client, server)
	//io.Copy(server, client)
}


func deep_copy(conn1 net.Conn, conn2 net.Conn) {

} 

func shadow_copy(conn1 net.Conn, conn2 net.Conn) {

}