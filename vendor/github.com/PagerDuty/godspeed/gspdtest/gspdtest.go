// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

// Package gspdtest is a package used by Godspeed for testing. This package
// isn't really meant to be consumed by anyone.
package gspdtest

import (
	"bytes"
	"fmt"
	"net"
)

// Listener is a function which takes a *net.UDPConn and sends any data received
// on it back over the c channel. This function is meant to be ran within a
// goroutine. The ctrl channel is used to shut down the goroutine.
func Listener(l *net.UDPConn, ctrl chan int, c chan []byte) {
	for {
		select {
		case _, ok := <-ctrl:
			if !ok {
				close(c)
				return
			}
		default:
			buffer := make([]byte, 8193)

			_, err := l.Read(buffer)

			if err != nil {
				continue
			}

			c <- bytes.Trim(buffer, "\x00")
		}
	}
}

// BuildListener is a function which builds a *net.UDPConn listening on localhost
// on the port specified. It also returns a control channel and a return channel.
func BuildListener(port int) (*net.UDPConn, chan int, chan []byte) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))

	if err != nil {
		panic(fmt.Sprintf("getting address for test listener failed, bailing out. Here's everything I know: %v", err))
	}

	l, err := net.ListenUDP("udp", addr)

	if err != nil {
		panic(fmt.Sprintf("unable to listen for traffic: %v", err))
	}

	return l, make(chan int), make(chan []byte)
}
