package main

import (
	"fmt"
	"net"
	"os"
)

const SERVER_IP = "0.0.0.0"
const SERVER_PORT = "1060"

func main() {
	rtspSocket, err := net.Listen("tcp", SERVER_IP+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer rtspSocket.Close()
	fmt.Println("Listening on " + SERVER_IP + ":" + SERVER_PORT)

	for {
		rtspSoc, err := rtspSocket.Accept()
		if err!= nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		var ser ServerWorker
		ser.serverWorker(rtspSoc)
		ser.run()
	}

}