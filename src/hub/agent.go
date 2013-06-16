package main

import (
	"log"
	"net"
	"sync/atomic"
)

import (
	"cfg"
	. "helper"
	"hub/protos"
	"misc/packet"
)

const (
	MAXCHAN = 100000
)

//---------------------------------------------------------- 连入主机的ID标示
var _host_genid int32

func init() {
	log.SetPrefix("[HUB]")
}

//---------------------------------------------------------- Hub processing
func HubAgent(incoming chan []byte, conn net.Conn) {
	config := cfg.Get()
	if config["profile"] == "true" {
		SetMemProfileRate(1)
		defer func() {
			GC()
			DumpHeap()
			PrintGCSummary()
		}()
	}

	hostid := atomic.AddInt32(&_host_genid, 1)
	// forward buffer
	forward := make(chan []byte, MAXCHAN)
	// output buffer
	output := make(chan []byte, MAXCHAN)

	protos.AddServer(hostid, forward)
	log.Printf("server id:%v connected\n", hostid)

	go _write_routine(output, conn)

	defer func() {
		protos.RemoveServer(hostid)
		close(forward)
		close(output)

		log.Printf("server id:%v disconnected\n", hostid)
	}()

	for {
		select {
		case msg, ok := <-incoming: // from hub
			if !ok {
				return
			}

			reader := packet.Reader(msg)
			go protos.HandleRequest(hostid, reader, output)
		case msg := <-forward: // send forward packet
			SendChan(0, msg, output)
		}
	}

}

//---------------------------------------------------------- writer routine
func _write_routine(output chan []byte, conn net.Conn) {
	for {
		msg, ok := <-output
		if !ok {
			break
		}

		_, err := conn.Write(msg) // write operation is assumed to be atomic
		if err != nil {
			log.Println("Error send reply to GS:", err)
		}
	}
}
