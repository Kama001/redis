package server

import (
	"fmt"
	"log"
	"net"
	"redis/config"
	"redis/core"
	"syscall"
//	"errors"
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
		fmt.Println("looking for events.....")
		nevents, e := syscall.EpollWait(epollFD, events[:], -1)
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
				fmt.Println("Started receiving data.....")
				comm := core.FDComm{Fd: int(events[i].Fd)}
				cmds, err := readCommands(comm)
				if err != nil {
					if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
						// No data available or socket would block, skip this iteration
						continue
					} else if err.Error() == "no data" {
						log.Printf("Client socket %d closed the connection", comm.Fd)
					} else {
						log.Println("Error reading from client:", err)
					}
					syscall.Close(int(events[i].Fd)) // Close the connection on error
					continue
				}
				fmt.Println(cmds)
			}
		}
	}
}

func toArrayString(ai []interface{}) ([]string, error) {
	var as []string = make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as, nil
}

func readCommands(f core.FDComm) (core.RedisCmds, error) {
	buf := make([]byte, 512)
	n, err := f.Read(buf[:])
	if err != nil {
		return nil, err
	}
	fmt.Println("Received data is:", string(buf))
	args, err := core.Parser(buf[:n])
	fmt.Println("Parsed data is: ", args)
	//if n == 0 {
	//	return nil, errors.New("no data")
	//}
	if err != nil {
		// fmt.Println("connection closed error")
		// fmt.Println(err.Error())
		return nil, err
	}
	var cmds []*core.RedisCmd = make([]*core.RedisCmd, 0)
	for _, arg := range args {
		tokens, err := toArrayString(arg.([]interface{}))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, &core.RedisCmd{
			Cmd:  tokens[0],
			Args: tokens[1:],
		})
	}
	fmt.Println("fetched the cmds.....")
	// log.Printf("Received from client %d: %s", f.Fd, string(buf[:n]))
	return cmds, nil
}
