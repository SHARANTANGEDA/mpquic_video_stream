package main

import (
	crrand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	quic "github.com/lucas-clemente/quic-go"
	"io"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const SETUP = 11
const PLAY = 12
const PAUSE = 13
const TEARDOWN = 14
const INIT = 0
const READY = 1
const PLAYING = 2
const OK_200 = 0
const FILE_NOT_FOUND_404 = 1
const CON_ERR_500 = 2

type ServerWorker struct {
	clientInfo struct {
		rtspSocket  quic.Session
		videoStream VideoStream
		session     int
		rtpPort     string
		rtpSocket   quic.Session
		eventMutex  sync.Mutex
		event       sync.WaitGroup
	}
	state int
}

func (ser *ServerWorker) serverWorker(rtspSoc quic.Session) {
	ser.state = INIT
	ser.clientInfo.rtspSocket = rtspSoc

	//raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:3001")
	//caddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:5000")
	//if err != nil {
	//	fmt.Println(err)
	//	fmt.Println("There was an error in socket")
	//	return
	//}
	//ser.clientInfo.rtpSocket, err = net.DialUDP("udp",caddr, raddr)
	//socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
}

func receiveData(conn quic.Session) ([]byte, int) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 256)
	// Read the incoming connection into the buffer.
	stream, err := conn.AcceptStream()
	readLen, err := io.ReadAtLeast(stream, buf, 5)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		os.Exit(1)
	}
	// Send a response back to person contacting us.
	//_, _ = conn.Write([]byte("Message received."))

	return buf, readLen
}

func (ser *ServerWorker) run() {
	go ser.recvRtspRequest()
	//cfgServer := &quic.Config{}
	//tlsConfig := generateTLSConfig()
	//listener, err := quic.ListenAddr("0.0.0.0:3001", tlsConfig, cfgServer)
	//sess, err := listener.Accept()
	//ser.clientInfo.rtpSocket, err = sess.AcceptStream()
	//if err != nil {
	//	fmt.Println(err)
	//	fmt.Println("There was an error in socket")
	//	return
	//}
}

func (ser *ServerWorker) recvRtspRequest() {
	connSocket := ser.clientInfo.rtspSocket
	for {
		data, readLen := receiveData(connSocket)
		if readLen > 0 {
			fmt.Println(readLen, "\nData received:\n", string(data))
			ser.processRtspRequest(string(data[:readLen]))
		}

	}
}

func (ser *ServerWorker) processRtspRequest(data string) {
	request := strings.Split(data, "\n")
	line1 := strings.Split(request[0], " ")
	fmt.Println("Process Request", request[0], "SEP", request[1:], "SEP", request[1])
	var requestType int

	if line1[0] == "SETUP" {
		fmt.Println("RECD", line1[0], line1)
		requestType = 11
	} else if line1[0] == "PLAY" {
		requestType = 12
	} else if line1[0] == "PAUSE" {
		requestType = 13
	} else if line1[0] == "TEARDOWN" {
		requestType = 14
	}
	filename := line1[1]
	seq := strings.Split(request[1], " ")
	if requestType == SETUP {
		if ser.state == INIT {
			fmt.Println("SETUP Request received")
			ser.clientInfo.videoStream.videoStream(filename)
			ser.state = READY
			//except IOError:
			//	self.replyRtsp(self.FILE_NOT_FOUND_404, seq[1])
			go ser.createQUICClient()
			ser.clientInfo.session = rand.Intn(999999-100000) + 100000
			fmt.Println("Int For Session", ser.clientInfo.session)
			ser.replyRtsp(OK_200, seq[0])
			fmt.Println("sequenceNum is ", seq[0])
			ser.clientInfo.rtpPort = strings.Split(request[2], " ")[3]
			fmt.Println("--", "\nrtpPort is :"+ser.clientInfo.rtpPort, "\n", "--")
			fmt.Println("filename is ", filename)
		}
	} else if requestType == PLAY {
		if ser.state == READY {
			fmt.Println('-'*60, "\nPLAY Request Received\n", '-'*60)
			ser.state = PLAYING

			ser.replyRtsp(OK_200, seq[0])
			fmt.Println("-", "\nSequence Number (", seq[0], ")\nReplied to client\n", "-")
			//ser.clientInfo.event.Add(1)
			//defer ser.clientInfo.event.Done()
			go ser.sendRtp()
		} else if ser.state == PAUSE {
			fmt.Println('-'*60, "\nRESUME Request Received\n", '-'*60)
			ser.state = PLAYING
		}
	} else if requestType == PAUSE {
		if ser.state == PLAYING {
			fmt.Println('-'*60, "\nPAUSE Request Received\n", '-'*60)
			ser.state = READY

			ser.clientInfo.event.Add(1)
			ser.replyRtsp(OK_200, seq[0])
		}
	} else if requestType == TEARDOWN {
		fmt.Println('-'*60, "\nTEARDOWN Request Received\n", '-'*60)
		ser.clientInfo.event.Add(1)
		ser.replyRtsp(OK_200, seq[0])
		ser.clientInfo.rtpSocket.Close()
	}
}

func (ser *ServerWorker) createQUICClient() {
	cfgClient := &quic.Config{}
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	var err error
	ser.clientInfo.rtpSocket, err = quic.DialAddr("localhost:3001", tlsConfig, cfgClient)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Connected")

}

func (ser *ServerWorker) sendRtp() {
	counter := 0
	//threshold := 10
	for {
		fmt.Println("In Send Data")
		//ser.clientInfo.event.Wait()
		//for {
		//	//jit := math.Floor(rand.uniform(-13,5.99))
		//	//jit = jit / 1000
		//
		//	//jit = jit + 0.020
		//	//
		//	//if ser.clientInfo['event'].isSet():
		//	//break
		//}
		//time.Sleep(time.Second)

		data := ser.clientInfo.videoStream.nextFrame()
		frameNumber := ser.clientInfo.videoStream.frameNbr()
		//prb := math.Floor(-13 + rand.Float64()*(5.99 + 13))

		//prb := 6.0
		if data != nil {
			fmt.Println("In Send Data", len(data), frameNumber)
			//for {
			//	if ser.clientInfo.rtpSocket != nil {
			//		fmt.Println("1")
			//		break
			//	}
			//}
			stream, err := ser.clientInfo.rtpSocket.OpenStreamSync()
			replyLen, err := stream.Write(ser.makeRtp(data, frameNumber))
			if err != nil {
				fmt.Println("Error")
			}
			//sendto(self.makeRtp(data, frameNumber),(self.clientInfo['rtspSocket'][1][0],port))
			counter += 1
			time.Sleep(10 * time.Millisecond)
			fmt.Println(replyLen, ": REPL LEN")
		}
	}

}

func (ser *ServerWorker) makeRtp(payload []byte, frameNbr int) []byte {
	version := 2
	padding := 0
	extension := 0
	cc := 0
	marker := 0
	pt := 26
	seqnum := frameNbr
	ssrc := 0
	var rtpP RtpPacket
	rtpP.rtpPacket()
	rtpP.encode(version, padding, extension, cc, seqnum, marker, pt, ssrc, payload)
	return rtpP.getPacket()
}

func (ser *ServerWorker) replyRtsp(code int, seq string) {
	if code == OK_200 {
		reply := "RTSP/1.0 200 OK\nCSeq: " + seq + "\nSession: " + strconv.Itoa(ser.clientInfo.session)
		b := []byte(reply)
		fmt.Println("REPLY", reply, "BYTE LEN:", len(b))
		stream, err := ser.clientInfo.rtspSocket.OpenStreamSync()
		fmt.Println("Here")
		replyLen, err := stream.Write(b)
		fmt.Println("Here")
		if err != nil {
			fmt.Println("Error", replyLen)
			os.Exit(1)
		}
	} else if code == FILE_NOT_FOUND_404 {
		fmt.Println("404 NOT FOUND")

	} else if code == CON_ERR_500 {
		fmt.Println("500 CONNECTION ERROR")
	}
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(crrand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(crrand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
