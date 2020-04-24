package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/gotk3/gotk3/gtk"
	quic "github.com/lucas-clemente/quic-go"
	"io"
	"log"
	"math/big"
	"net"
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
const CACHE_FILE_NAME = "cache-"
const CACHE_FILE_EXT = ".jpg"

type ClientWorker struct {
	state         int
	counter       int
	serverAddr    string
	serverPort    string
	rtpPort       int
	filename      string
	rtspSeq       int
	sessionId     int
	requestSent   int
	teardownAcked int
	frameNbr      int
	rtpSocket     quic.Session
	setup         gtk.Button
	win           *gtk.Window
	mainBox       *gtk.Box
	event         sync.WaitGroup
	rtspSocket    net.Conn
	rtpPacket     RtpPacket
	cAddr         net.Addr
	prevImage     *gtk.Image
	image         *gtk.Image
}

func (cw *ClientWorker) clientWorker(win *gtk.Window, serverAddr string, serverPort string, rtpPort string, filename string) {
	cw.state = INIT
	cw.counter = 0
	cw.serverAddr = serverAddr
	cw.serverPort = serverPort
	cw.rtpPort, _ = strconv.Atoi(rtpPort)
	cw.filename = filename
	cw.rtspSeq = 0
	cw.sessionId = 0
	cw.requestSent = -1
	cw.teardownAcked = 0
	cw.win = win
	cw.createWidgets()
	cw.connectToServer()
	cw.frameNbr = 0
	cw.prevImage = nil
	cw.image = nil
	cw.setupMovie()
	time.Sleep(2 * time.Second)
	cw.playMovie()
}

func (cw *ClientWorker) createWidgets() {
	cw.mainBox, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	setup, err := gtk.ButtonNew()
	if err != nil {
		log.Fatal("Unable to create SetupBtn:", err)
	}
	setup.SetLabel("Setup")
	setup.Connect("clicked", cw.setupMovie)

	start, err := gtk.ButtonNew()
	if err != nil {
		log.Fatal("Unable to create PlayBtn:", err)
	}
	start.SetLabel("Play")
	start.Connect("clicked", cw.playMovie)

	pause, err := gtk.ButtonNew()
	if err != nil {
		log.Fatal("Unable to create PauseBtn:", err)
	}
	pause.SetLabel("Pause")
	pause.Connect("clicked", cw.pauseMovie)

	teardown, err := gtk.ButtonNew()
	if err != nil {
		log.Fatal("Unable to create TeardownBtn:", err)
	}
	teardown.SetLabel("Teardown")
	teardown.Connect("clicked", cw.exitClient)
	cw.win.Add(cw.mainBox)
	cw.mainBox.PackStart(setup, false, false, 0)
	cw.mainBox.PackStart(start, false, false, 0)
	cw.mainBox.PackStart(pause, false, false, 0)
	cw.mainBox.PackStart(teardown, false, false, 0)

}

func (cw *ClientWorker) setupMovie() {
	if cw.state == INIT {
		fmt.Println("IN INIT Sending Rtsp Request")
		go cw.setupQuicServer()
		cw.sendRtspRequest(SETUP)

	}
}

func (cw *ClientWorker) setupQuicServer() {
	cfgServer := &quic.Config{}
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	fmt.Println("CHECK: ", cw.rtpPort)
	listener, err := quic.ListenAddr("0.0.0.0:3001", tlsConfig, cfgServer)
	cw.rtpSocket, err = listener.Accept()
	if err != nil {
		fmt.Println(err)
		fmt.Println("There was an error in socket")
		return
	}
	fmt.Println("Client has been connected")
}

func (cw *ClientWorker) exitClient() {
	cw.sendRtspRequest(TEARDOWN)
	gtk.MainQuit()
	_ = os.Remove(CACHE_FILE_NAME + strconv.Itoa(cw.sessionId) + CACHE_FILE_EXT)
	rate := float64(cw.counter / cw.frameNbr)
	_ = fmt.Sprintf("-", "\nRTP Packet Loss Rate : %f", rate, "\n", "-")
}

func (cw *ClientWorker) pauseMovie() {
	if cw.state == PLAYING {
		cw.sendRtspRequest(PAUSE)
	}
}

func (cw *ClientWorker) playMovie() {
	if cw.state == READY {
		fmt.Println("Playing Movie")
		go cw.listenRtp()
		cw.sendRtspRequest(PLAY)
		//cw.listenRtp()
	}
}

func (cw *ClientWorker) listenRtp() {
	for {
		fmt.Println("Trying to read")
		buf := make([]byte, 20480)
		// Read the incoming connection into the buffer.
		var reqLen int
		var err error
		//rAddr, err := net.ResolveUDPAddr("udp", cw.serverAddr+":"+cw.serverPort)
		//cw.rtpSocket, err = net.ListenUDP("udp", rAddr)
		if err != nil {
			fmt.Println("Error in creating UDP Socket", err)
			os.Exit(1)
		}
		//for {
		//	if cw.rtpSocket != nil {
		//		fmt.Println("0")
		//		break
		//	}
		//}
		stream, err := cw.rtpSocket.AcceptStream()
		reqLen, err = io.ReadAtLeast(stream, buf, 15)
		if err != nil {
			fmt.Println("Didn`t receive data!")
			//cw.event.Done()
			if cw.teardownAcked == 1 {
				cw.rtpSocket.Close()
				break
			}
			fmt.Println("Error reading:", err.Error())
		}
		fmt.Println("Data recieved check", reqLen)
		if reqLen > 0 {
			cw.rtpPacket.decode(buf)
			fmt.Println("||Received Rtp Packet #", strconv.Itoa(cw.rtpPacket.seqNum()), "|| ")
			var currFrameNbr int
			if cw.frameNbr+1 != cw.rtpPacket.seqNum() {
				cw.counter += 1
				fmt.Println("==", "\nPACKET LOSS\n", "==")
				currFrameNbr = cw.rtpPacket.seqNum()
			}
			if currFrameNbr > cw.frameNbr {
				cw.frameNbr = currFrameNbr
				fmt.Println("GOing to update frame", reqLen)
				cw.updateMovie(cw.writeFrame(cw.rtpPacket.getPayload()))
			}
		}
	}
}

func (cw *ClientWorker) writeFrame(data []byte) string {
	cacheName := CACHE_FILE_NAME + strconv.Itoa(cw.sessionId) + CACHE_FILE_EXT
	file, err := os.Create(cacheName)
	if err != nil {
		log.Fatal(err)
	}
	file.Write(data)
	file.Close()
	return cacheName
}

func (cw *ClientWorker) updateMovie(imageFile string) {
	//imageBox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	var err error
	if cw.image == nil {
		cw.image, err = gtk.ImageNewFromFile(imageFile)
		cw.mainBox.PackStart(cw.image, false, false, 0)
	} else {
		cw.image.SetFromFile(imageFile)
	}
	cw.image.Show()
	//image, err := gtk.ImageNewFromFile(imageFile)
	if err != nil {
		fmt.Println("Photo Error")
	}
	//if cw.prevImage != nil {
	//	fmt.Println("Remove Package")
	//	cw.mainBox.PackEnd(cw.prevImage, false, false, 0)
	//	cw.prevImage.Clear()
	//}
	//
	time.Sleep(100 * time.Millisecond)
	//cw.prevImage = image

	//cw.win.Add(cw.mainBox)
	//imagePixBuffer := image.GetPixbuf()
	//horizontalSize := imagePixBuffer.GetWidth()
	//verticalSize := imagePixBuffer.GetHeight()
	//
	//cw.win.SetSizeRequest(horizontalSize, verticalSize)
}

func (cw *ClientWorker) connectToServer() {
	l, err := net.ResolveTCPAddr("tcp", cw.serverAddr+":"+cw.serverPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	cw.rtspSocket, err = net.DialTCP("tcp", nil, l)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}
	fmt.Println("TCP Server connected")
	//cfgClient := &quic.Config{}
	//tlsConfig := &tls.Config{InsecureSkipVerify: true}
	//session, err := quic.DialAddr("localhost:3001", tlsConfig, cfgClient)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Connected")
	//cw.rtpSocket, err = session.OpenStreamSync()
}

func (cw *ClientWorker) sendRtspRequest(requestCode int) {
	if requestCode == SETUP && cw.state == INIT {
		go cw.recvRtspReply()
		fmt.Println("Recv Rtsp Reply called")
		cw.rtspSeq = 1
		request := "SETUP " + cw.filename + "\n" + strconv.Itoa(cw.rtspSeq) + "\n" + " RTSP/1.0 RTP/UDP " + strconv.Itoa(cw.rtpPort)
		writeLn, err := cw.rtspSocket.Write([]byte(request))
		if err != nil {
			fmt.Println("Error in Rtsp Socket write method", writeLn)
			fmt.Println(err)
			os.Exit(1)
		}
		cw.requestSent = SETUP
	} else if requestCode == PLAY && cw.state == READY {
		cw.rtspSeq = cw.rtspSeq + 1
		request := "PLAY " + "\n" + strconv.Itoa(cw.rtspSeq)
		cw.rtspSocket.Write([]byte(request))
		fmt.Println("-", "\nPLAY request sent to Server...\n", "-")
		cw.requestSent = PLAY
	} else if requestCode == PAUSE && cw.state == PLAYING {
		cw.rtspSeq = cw.rtspSeq + 1
		request := "PAUSE " + "\n" + strconv.Itoa(cw.rtspSeq)
		cw.rtspSocket.Write([]byte(request))
		fmt.Println("-", "\nPAUSE request sent to Server...\n", "-")
		cw.requestSent = PAUSE
	} else if requestCode == TEARDOWN && !(cw.state == INIT) {
		cw.rtspSeq = cw.rtspSeq + 1
		request := "TEARDOWN " + "\n" + strconv.Itoa(cw.rtspSeq)
		cw.rtspSocket.Write([]byte(request))
		fmt.Println("-", "\nTEARDOWN request sent to Server...\n", "-")
		cw.requestSent = TEARDOWN
	} else {
		return
	}
}

func (cw *ClientWorker) recvRtspReply() {
	for {
		buf := make([]byte, 1024)
		reply, err := cw.rtspSocket.Read(buf)
		if err != nil {
			fmt.Println("Error in recvReply", err, reply)
			os.Exit(1)
		}
		if reply > 0 {
			fmt.Println("Going into parse rtsp reply", string(buf[:reply]))
			cw.parseRtspReply(string(buf[:reply]))
		}
		fmt.Println("REPLY", reply)
		if cw.requestSent == TEARDOWN {
			cw.rtspSocket.Close()
			break
		}
	}
}

func (cw *ClientWorker) parseRtspReply(data string) {
	fmt.Println("Parsing Received Rtsp data...")
	lines := strings.Split(data, "\n")
	seqNumStr := strings.Split(lines[1], " ")
	fmt.Println(len(data))
	seqNum, err := strconv.Atoi(string(seqNumStr[1]))
	if err != nil {
		fmt.Println("On Parse", err.Error())
	}
	if seqNum == cw.rtspSeq {
		session, _ := strconv.Atoi(strings.Split(lines[2], " ")[1])
		if cw.sessionId == 0 {
			cw.sessionId = session
		}
		if cw.sessionId == session {
			temp, _ := strconv.Atoi(strings.Split(lines[0], " ")[1])
			if temp == 200 {
				if cw.requestSent == SETUP {
					fmt.Println("Updating RTSP state...")
					cw.state = READY
					fmt.Println("Setting Up RtpPort for Video Stream")
					cw.openRtpPort()
				} else if cw.requestSent == PLAY {
					cw.state = PLAYING
					fmt.Println("--", "\nClient is PLAYING...\n", "--", cw.state)
				} else if cw.requestSent == PAUSE {
					cw.state = READY
					cw.event.Add(1)
				} else if cw.requestSent == TEARDOWN {
					cw.teardownAcked = 1
				}
			}
		}
	}
}

func (cw *ClientWorker) openRtpPort() {
	//_ = cw.rtpSocket.SetDeadline(time.Now().Add(500 * time.Millisecond))
	fmt.Println("Bind RtpPort Success")
}

func (cw *ClientWorker) handler() {
	cw.pauseMovie()
	go cw.listenRtp()
	cw.sendRtspRequest(PLAY)
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
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
