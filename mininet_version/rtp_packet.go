package main

import (
	"fmt"
	"time"
)

const HeaderSize = 12

type RtpPacket struct {
	header  []byte
	payload []byte
}

func (r *RtpPacket) rtpPacket() {
	r.header = make([]byte, HeaderSize)
}

func (r *RtpPacket) encode(version int, padding int, extension int, cc int, seqnum int, marker int, pt int, ssrc int, payload []byte) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	fmt.Println("TimeStamp:", timestamp)
	r.header = make([]byte, HeaderSize)
	r.header[0] = byte(version << 6)
	r.header[0] = byte(version << 6)
	r.header[0] = r.header[0] | byte(padding<<5)
	r.header[0] = r.header[0] | byte(extension<<4)
	r.header[0] = r.header[0] | byte(cc)
	r.header[1] = byte(marker << 7)
	r.header[1] = r.header[1] | byte(pt)

	r.header[2] = byte(seqnum >> 8)
	r.header[3] = byte(seqnum)

	r.header[4] = byte((timestamp >> 24) & 0xFF)
	r.header[5] = byte((timestamp >> 16) & 0xFF)
	r.header[6] = byte((timestamp >> 8) & 0xFF)
	r.header[7] = byte(timestamp & 0xFF)

	r.header[8] = byte(ssrc >> 24)
	r.header[9] = byte(ssrc >> 16)
	r.header[10] = byte(ssrc >> 8)
	r.header[11] = byte(ssrc)
	r.payload = payload
}

func (r *RtpPacket) decode(byteStream []byte) {
	r.header = byteStream[:HeaderSize]
	r.payload = byteStream[HeaderSize:]
}

func (r *RtpPacket) version() int {
	return int(r.header[0] >> 6)
}

func (r *RtpPacket) seqNum() int {
	seqNum := int(r.header[2])<<8 | int(r.header[3])
	return seqNum
}

func (r *RtpPacket) timestamp() int {
	timestamp := int(r.header[4])<<24 | int(r.header[5])<<16 | int(r.header[6])<<8 | int(r.header[7])
	return timestamp
}

func (r *RtpPacket) payloadType() int {
	pt := r.header[1] & 127
	return int(pt)
}

func (r *RtpPacket) getPayload() []byte {
	return r.payload
}

func (r *RtpPacket) getPacket() []byte {
	return append(r.header, r.payload...)
}
