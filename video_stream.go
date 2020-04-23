package main

import (
	"fmt"
	"log"
	"os"
)

type VideoStream struct {
	filename string
	frameNum int
	file *os.File
}

func(s *VideoStream) videoStream(filename string){
	s.filename = filename
	var err error
	s.file, err = os.Open(filename) // For read access.
	fmt.Println(filename, s.file)
	if err != nil {
		log.Fatal(err)
	}
	s.frameNum = 0
}


func(s *VideoStream) nextFrame() []byte {
	data := make([]byte, 5)
	count, err := s.file.Read(data)
	if err != nil {
		fmt.Println(count, "REPL LEN STREAM")
		log.Fatal(err)
	}
	fmt.Println("Here")
	fmt.Printf("read %d bytes: %q\n", count, data[:count])
	data_int := (int64(data[0]) - 48) * 10000 + (int64(data[1]) - 48) * 1000 + (int64(data[2]) - 48) * 100 + (int64(data[3]) - 48) * 10 + (int64(data[4]) - 48)
	final_data_int := data_int
	frame_length := final_data_int
	frame := make([]byte, frame_length)
	count, err = s.file.Read(frame)
	if err != nil {
		log.Fatal(err)
	}
	s.frameNum +=1
	fmt.Println('-'*10, "\nNext Frame (#", string(s.frameNum), ") length:", string(frame_length), "\n", '-'*10)
	return frame
}

func(s *VideoStream) frameNbr() int{
	return s.frameNum
}