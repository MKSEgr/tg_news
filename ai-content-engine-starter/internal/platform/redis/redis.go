package redis

import (
	"fmt"
	"net"
	"strconv"
)

// ValidateAddr checks whether Redis address has host:port form.
func ValidateAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("parse redis addr: %w", err)
	}
	if host == "" {
		return fmt.Errorf("redis host is required")
	}
	if port == "" {
		return fmt.Errorf("redis port is required")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse redis port: %w", err)
	}
	if portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("redis port out of range: %d", portNumber)
	}

	return nil
}
