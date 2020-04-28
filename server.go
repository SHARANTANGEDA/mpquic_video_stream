package main

import (
	"fmt"
	quic "github.com/lucas-clemente/quic-go"
	"os"
)

const SERVER_IP = "0.0.0.0"
const SERVER_PORT = "1060"

func main() {
	cfgServer := &quic.Config{
		CreatePaths: true,
	}
	tlsConfig := generateTLSConfig()
	listener, err := quic.ListenAddr(SERVER_IP+":"+SERVER_PORT, tlsConfig, cfgServer)
	if err != nil {
		fmt.Println(err)
		fmt.Println("There was an error in socket")
		return
	}
	fmt.Println("RTSP QUIC Connection has been established")
	//rtspSocket, err := net.Listen("tcp", SERVER_IP+":"+SERVER_PORT)
	//if err != nil {
	//	fmt.Println("Error listening:", err.Error())
	//	os.Exit(1)
	//}
	//defer rtspSocket.Close()
	fmt.Println("Listening on " + SERVER_IP + ":" + SERVER_PORT)

	for {
		rtspSocket, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		var ser ServerWorker
		ser.serverWorker(rtspSocket)
		ser.run()
	}

}
