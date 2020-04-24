cd client
sudo go mod tidy
sudo go build client.go client_worker.go rtp_packet.go
./client 10.0.0.1 1060 3001 video3.mjpeg