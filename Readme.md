# MPQUIC Video Stream using RTSP and RTP Protocols from scratch in GO

## Follow the instructions in CoNEXT 17 mp-quic site
## Make sure to change quic-go repo version to 0.11.2 before add remote of MP-QUIC
## Use Mininet for MPQUIC support
### This application is for Go 1.14.2
##### For Server build
`sudo GO111MODULE=off go build server.go server_worker.go rtp_packet.go video_stream.go`

##### For Client Build
`sudo GO111MODULE=off go build client.go client_worker.go rtp_packet.go`


##### Server Run
`./server`

##### Client Run
`./client 10.0.0.1 1060 3001 video3.mjpeg`