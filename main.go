package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func panicf(format string, v ...any) {
	panic(fmt.Sprintf(format, v...))
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal("at least one port is required")
	}

	var wg sync.WaitGroup
	tcpListeners := make([]net.Listener, len(os.Args)-1)
	udpListeners := make([]*net.UDPConn, len(os.Args)-1)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	for i := range tcpListeners {
		port := os.Args[i+1]
		lisAddr := net.JoinHostPort("0.0.0.0", port)

		tcpLis, err := net.Listen("tcp4", lisAddr)
		if err != nil {
			panicf("error listening on tcp port %s: %v", port, err)
		}
		defer tcpLis.Close()

		tcpListeners[i] = tcpLis
		log.Printf("listening on tcp port %s", port)

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				conn, err := tcpLis.Accept()
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					} else {
						log.Printf("error accepting tcp connection on port %d: %v", i, err)
					}

					continue
				}

				log.Printf("got tcp connection on port %s from %s", port, conn.RemoteAddr())
				conn.Close()
			}
		}()

		addr, err := net.ResolveUDPAddr("udp4", lisAddr)
		if err != nil {
			panicf("error resolving udp port %s: %v", port, err)
		}
		udpLis, err := net.ListenUDP("udp4", addr)
		if err != nil {
			panicf("error listening on udp port %s: %v", port, err)
		}
		defer udpLis.Close()

		udpListeners[i] = udpLis
		log.Printf("listening on udp port %s", port)

		wg.Add(1)
		go func() {
			defer wg.Done()

			buf := make([]byte, 1)
			for {
				_, conn, err := udpLis.ReadFromUDP(buf)
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						return
					} else {
						log.Printf("error accepting udp connection on port %d: %v", i, err)
					}
				}

				log.Printf("got udp connection on port %s from %s", port, conn)
			}
		}()
	}

	log.Println("waiting for connections")

	<-ctx.Done()
	log.Println("shutting down")
	for _, l := range tcpListeners {
		l.Close()
	}
	for _, l := range udpListeners {
		l.Close()
	}
	wg.Wait()
}
