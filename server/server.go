package server

import (
	"fmt"
	"log"
	"net"
	"redis/config"
	"syscall"
)

func StartTCPServer() {
	max_clients := 20000
	con_clients := 0

	Host := config.Host
	Port := config.Port
	// create EPOLL Event objects to hold events
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	// create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)

	if err != nil {
		panic(err)
	}

	defer syscall.Close(serverFD)

	if err = syscall.SetNonblock(serverFD, true); err != nil {
		panic(err)
	}

	ip4 := net.ParseIP(Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		panic(err)
	}

	if err = syscall.Listen(serverFD, max_clients); err != nil {
		panic(err)
	}

	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err) // why log.Fatal(err) not return err
	}

	defer syscall.Close(epollFD)

	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		panic(err)
	}
	fmt.Println("Starting redis server.....")
	for {
		nevents, e := syscall.EpollWait(epollFD, events[:], 5)
		if e != nil {
			continue
		}
		for i := 0; i < nevents; i++ {
			if int(events[i].Fd) == serverFD {
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}
				con_clients++
				syscall.SetNonblock(serverFD, true)
				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}
				if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &socketClientEvent); err != nil {
					log.Fatal(err)
				}
				log.Printf("New client %d socket created", socketClientEvent.Fd)
				log.Printf("Number of connected clients : %d", con_clients)
			} else {
				buf := make([]byte, 1024) // Create a buffer to store incoming data

				n, err := syscall.Read(int(events[i].Fd), buf) // Read from the client socket
				if err != nil {
					if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
						// No data available or socket would block, skip this iteration
						continue
					}
					log.Println("Error reading from client:", err)
					syscall.Close(int(events[i].Fd)) // Close the connection on error
					continue
				}

				if n == 0 {
					// Client has closed the connection
					log.Printf("Client socket %d closed the connection", int(events[i].Fd))
					syscall.Close(int(events[i].Fd)) // Close the socket
					con_clients--                    // Decrease the client count
					log.Printf("Number of connected clients : %d", con_clients)
					continue
				}

				// Print the data received from the client
				log.Printf("Received from client %d: %s\n", int(events[i].Fd), string(buf[:n]))
			}
		}
	}
}
