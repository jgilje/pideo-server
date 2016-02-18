package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"time"
)

type client struct {
	conn    net.Conn
	ch      chan []byte
	pending int32
}

const maxBackBuffer = int32(64)

func main() {
	portPtr := flag.Uint("port", 12345, "port")
	if *portPtr > math.MaxUint16 || *portPtr == 0 {
		fmt.Println("Invalid port")
	}

	testPtr := flag.Bool("test", false, "use test stream")
	debugPtr := flag.Bool("debug", false, "debug go routines on exit")

	flag.Parse()

	port := uint16(*portPtr)
	log.Println("Using port", port)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	msgchan := make(chan []byte)
	addchan := make(chan *client)
	rmchan := make(chan *client)

	go handleMessages(msgchan, addchan, rmchan)

	if *testPtr {
		log.Println("Using test stream")
		go testStream(msgchan)
	} else {
		go raspiCameraReader(msgchan)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}

			go handleConnection(conn, msgchan, addchan, rmchan)
		}
	}()

	installZeroconfListener("Pideo", "_pideo._tcp", port)

	if *debugPtr {
		handler := make(chan os.Signal, 1)
		signal.Notify(handler, os.Interrupt)

		select {
		case <-handler:
			log.Println("Exiting")

			buf := make([]byte, 1<<20)
			runtime.Stack(buf, true)
			log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf)

			os.Exit(0)
		}
	} else {
		select {}
	}
}

func (c *client) writeLinesFrom() {
	for msg := range c.ch {
		_, err := c.conn.Write(msg)
		if err != nil {
			return
		}
		atomic.AddInt32(&c.pending, -1)
		c.conn.SetDeadline(time.Now().Add(5 * time.Second))
	}
}

func handleConnection(c net.Conn, msgchan chan<- []byte, addchan chan<- *client, rmchan chan<- *client) {
	defer c.Close()
	client := &client{
		conn: c,
		ch:   make(chan []byte, maxBackBuffer),
	}
	c.SetDeadline(time.Now().Add(5 * time.Second))

	addchan <- client
	defer func() {
		rmchan <- client
	}()

	client.writeLinesFrom()
}

func handleMessages(msgchan <-chan []byte, addchan <-chan *client, rmchan <-chan *client) {
	clients := make(map[net.Conn]*client)

	for {
		select {
		case msg := <-msgchan:
			/*
							var wg sync.WaitGroup
							wg.Add(len(clients))
							for _, cl := range clients {
								go func(c *client) {
									defer wg.Done()
				                    // defer recover

									c.ch <- msg
									pending := atomic.AddInt32(&c.pending, 1)
									if pending > 10 {
										log.Println(pending)
									}
									if pending >= maxBackBuffer {
										log.Println("Client is not keeping up, dropping", c.conn.RemoteAddr)
										delete(clients, c.conn)
										close(c.ch)
									}
								}(cl)
							}
							wg.Wait()
			*/
			for _, cl := range clients {
				cl.ch <- msg
				pending := atomic.AddInt32(&cl.pending, 1)
				if pending >= maxBackBuffer {
					log.Println("Client is not keeping up, dropping", cl.conn.RemoteAddr())
					delete(clients, cl.conn)
					close(cl.ch)
				}
			}

		case client := <-addchan:
			log.Println("New client:", client.conn.RemoteAddr())
			clients[client.conn] = client
		case client := <-rmchan:
			log.Println("Client disconnected:", client.conn.RemoteAddr())
			if _, ok := clients[client.conn]; ok {
				delete(clients, client.conn)
				close(client.ch)
			}
		}
	}
}

/*
	defer func() {
		x := recover()
		if x != nil {
			log.Println("Failed to send over channel")
		}
	}()
*/
