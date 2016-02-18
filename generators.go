package main

import (
	"io"
	"log"
	"os/exec"
	"time"
)

func raspiCameraReader(ch chan<- []byte) {
	muxer := exec.Command("gst-launch-1.0", "-v", "mpegtsmux", "name=muxer", "!", "fdsink", "fdsrc", "do-timestamp=true", "!", "h264parse", "!", "muxer.")

	muxerIn, err := muxer.StdinPipe()
	if err != nil {
		log.Fatalln("Failed to get StdinPipe for muxer")
		return
	}
	muxerOut, err := muxer.StdoutPipe()
	if err != nil {
		log.Fatalln("Failed to get StdoutPipe for muxer")
		return
	}
	if err := muxer.Start(); err != nil {
		log.Fatal("Failed to start subcommand", err)
	}

	go func() {
		// recorder := exec.Command("raspivid", "-pf", "baseline", "-b", "1000000", "-ih", "-f", "-vf", "-hf", "-t", "0", "-h", "720", "-w", "1280", "-fps", "25", "-ex", "nightpreview", "-ISO", "3200", "-o", "-")
		recorder := exec.Command("raspivid", "-pf", "baseline", "-b", "1000000", "-ih", "-f", "-vf", "-hf", "-t", "0", "-h", "720", "-w", "1280", "-fps", "25", "-ex", "night", "-ISO", "800", "-o", "-")
		recorderOut, err := recorder.StdoutPipe()
		if err != nil {
			log.Fatalln("Failed to get StdoutPipe for recorder")
			return
		}
		if err := recorder.Start(); err != nil {
			log.Fatal("Failed to start subcommand", err)
		}

		var recorderBuffer = make([]byte, 16384)
		for {
			bytes, err := recorderOut.Read(recorderBuffer)
			if err != nil {
				log.Fatalln("Failed to read from recorder", err)
				return
			}

			_, err = muxerIn.Write(recorderBuffer[0:bytes])
			if err != nil {
				log.Fatalln("Failed to write to muxer", err)
			}
		}
	}()

	for {
		var videoBuffer = make([]byte, 16384)
		videoBufferBytes, err := muxerOut.Read(videoBuffer)
		if err != nil {
			if err == io.EOF {
				log.Println("No data from muxer")
				time.Sleep(500 * time.Millisecond)
				log.Println("Slept well")
			} else {
				log.Fatalln("Failed to read from muxer", err)
				return
			}
		}

		ch <- videoBuffer[0:videoBufferBytes]
		// log.Println("generator %d", videoBufferBytes)
	}
}

func testStream(ch chan<- []byte) {
	muxer := exec.Command("gst-launch-1.0", "-q", "videotestsrc", "is-live=true", "!", "x264enc", "byte-stream=true", "!", "video/x-h264,profile=baseline", "!", "mpegtsmux", "!", "fdsink")
	muxerOut, err := muxer.StdoutPipe()
	if err != nil {
		log.Fatalln("Failed to get StdoutPipe for recorder")
		return
	}
	if err := muxer.Start(); err != nil {
		log.Fatal("Failed to start subcommand", err)
	}

	for {
		var videoBuffer = make([]byte, 16384)
		videoBufferBytes, err := muxerOut.Read(videoBuffer)
		if err != nil {
			if err == io.EOF {
				log.Println("No data from muxer")
				time.Sleep(500 * time.Millisecond)
				log.Println("Slept well")
			} else {
				log.Fatalln("Failed to read from muxer", err)
				return
			}
		}

		ch <- videoBuffer[0:videoBufferBytes]
	}
}
