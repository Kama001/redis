package test

import (
	"fmt"
	"net"
	"redis/config"
	"time"
)

// pingTCP attempts to connect to a TCP port on the specified host
func pingTCP(host string, port int) bool {
	// Create the address string (host:port)
	address := fmt.Sprintf("%s:%d", host, port)

	// Attempt to connect to the target address within a timeout period
	_, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		// If error occurs, the port is not open
		return false
	}
	// If no error, the port is open
	return true
}

func StartTest() {
	// Host and port to check
	host := config.Host
	port := config.Port

	// Check if the port is open
	for i := 0; i < 50; i++ {
		if pingTCP(host, port) {
			fmt.Printf("Port %d on %s is open.\n", port, host)
		} else {
			fmt.Printf("Port %d on %s is closed.\n", port, host)
		}
	}
}
