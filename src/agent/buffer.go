package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
)

import (
	"cfg"
	"misc/packet"
)

type Buffer struct {
	ctrl    chan bool   // receive exit signal
	pending chan []byte // pending Packet
	max     int         // max queue size
	conn    net.Conn    // connection
}

const (
	DEFAULT_QUEUE_SIZE = 5
)

//------------------------------------------------ send packet
func (buf *Buffer) Send(data []byte) (err error) {
	// len of Channel: the number of elements queued (un-sent) in the channel buffer
	if len(buf.pending) < buf.max {
		buf.pending <- data
		return nil
	} else {
		return errors.New(fmt.Sprintf("Send Buffer Overflow, possible DoS attack. Remote: %v", buf.conn.RemoteAddr()))
	}
}

//------------------------------------------------ packet sender goroutine
func (buf *Buffer) Start() {
	defer func() {
		recover()
	}()

	for {
		select {
		case data, ok := <-buf.pending:
			if !ok {
				buf.conn.Close()
				return
			} else {
				buf.raw_send(data)
			}
		case _ = <-buf.ctrl:
			close(buf.pending)
		}
	}
}

//------------------------------------------------ send packet online
func (buf *Buffer) raw_send(data []byte) {
	writer := packet.Writer()
	writer.WriteU16(uint16(len(data)))
	writer.WriteRawBytes(data)

	_, err := buf.conn.Write(writer.Data())
	if err != nil {
		log.Println("Error send reply :", err)
		return
	}
}

//------------------------------------------------ create a new write buffer
func NewBuffer(conn net.Conn, ctrl chan bool) *Buffer {
	config := cfg.Get()
	max, err := strconv.Atoi(config["packet_queue"])
	if err != nil {
		max = DEFAULT_QUEUE_SIZE
		log.Println("cannot parse packet_queue from config", err)
	}

	buf := Buffer{conn: conn}
	buf.pending = make(chan []byte, max)
	buf.ctrl = ctrl
	buf.max = max
	return &buf
}
