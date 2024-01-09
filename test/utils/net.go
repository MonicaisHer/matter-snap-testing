package utils

import (
	"fmt"
	"log"
	"net"
	"strings"
	"testing"
	"time"
)

type Net struct {
	StartSnap        bool // should be set to true if services aren't started by default
	TestOpenPorts    []string
	TestBindLoopback []string
}

const dialTimeout = 2 * time.Second

var portService = map[string]string{
	// matter controller
	"5550": "chip-tool",
}

// servicePort looks up the service port by app name
func ServicePort(serviceName string) string {
	for p, s := range portService {
		if s == serviceName {
			return p
		}
	}
	panic("Found no port number for service: " + serviceName)
}

func TestNet(t *testing.T, snapName string, conf Net) {
	t.Run("net", func(t *testing.T) {
		if conf.StartSnap {
			t.Cleanup(func() {
				SnapStop(t, snapName)
			})
			SnapStart(t, snapName)
		}

		if len(conf.TestOpenPorts) > 0 {
			testOpenPorts(t, snapName, conf.TestOpenPorts)
		}
		if len(conf.TestBindLoopback) > 0 {
			testBindLoopback(t, snapName, conf.TestBindLoopback)
		}

	})
}

func testOpenPorts(t *testing.T, snapName string, ports []string) {
	t.Run("ports open", func(t *testing.T) {
		WaitServiceOnline(t, 60, ports...)
	})
}

func testBindLoopback(t *testing.T, snapName string, ports []string) {
	WaitServiceOnline(t, 60, ports...)

	t.Run("ports not listening on all interfaces", func(t *testing.T) {
		requireListenAllInterfaces(t, false, ports...)
	})

	t.Run("ports listening on localhost", func(t *testing.T) {
		requireListenLoopback(t, ports...)
		// requirePortOpen(t, ports...)
	})
}

// WaitServiceOnline waits for a service to come online by dialing its port(s)
// up to a maximum number
func WaitServiceOnline(t *testing.T, maxRetry int, ports ...string) error {
	closedPorts := make([]string, len(ports))
	copy(closedPorts, ports)

	prettyPorts := func(ports []string) string {
		prettyList := make([]string, len(ports))
		for i, p := range ports {
			if s, found := portService[p]; found {
				prettyList[i] = fmt.Sprintf("%s (%s)", p, s)
			} else {
				prettyList[i] = p
			}
		}
		return strings.Join(prettyList, ", ")
	}

	var returnErr error
	for i := 1; i <= maxRetry; i++ {

		msg := fmt.Sprintf("Retry %d/%d: Waiting for ports: %s", i, maxRetry, prettyPorts(closedPorts))
		if t != nil {
			t.Log(msg)
		} else {
			log.Print(msg)
		}

		var closedPortsTemp []string
		for _, port := range closedPorts {
			conn, err := net.DialTimeout("tcp", ":"+port, dialTimeout)
			if conn == nil {
				closedPortsTemp = append(closedPortsTemp, port)
			}
			returnErr = err
		}
		closedPorts = closedPortsTemp

		if len(closedPorts) == 0 {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	var err error
	if returnErr != nil {
		err = fmt.Errorf("Time out: reached max %d retries. Error: %v", maxRetry, returnErr)
	} else {
		err = fmt.Errorf("Time out: reached max %d retries.", maxRetry)
	}
	if t != nil {
		t.Fatal(err)
	} else {
		return err
	}

	return nil
}

// requirePortOpen checks if the local port(s) accepts connections
func requirePortOpen(t *testing.T, ports ...string) {
	if len(ports) == 0 {
		panic("No ports given as input")
	}

	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", ":"+port, dialTimeout)
		if err != nil {
			conn.Close()
			t.Errorf("Port %s is not open: %s", port, err)
		}

		if conn == nil {
			t.Errorf("Port %s is not open", port)
		}

		if conn != nil {
			t.Logf("Port %v is open.", port)
			conn.Close()
		}
	}
	if t.Failed() {
		t.FailNow()
	}
}

func requireListenAllInterfaces(t *testing.T, mustListen bool, ports ...string) {
	if len(ports) == 0 {
		panic("No ports given as input")
	}

	for _, port := range ports {
		isListening := isListenInterface(t, "*", port)

		if mustListen && !isListening {
			t.Errorf("Port %v not listening to all interfaces", port)
		} else if !mustListen && isListening {
			t.Errorf("Port %v is listening to all interfaces", port)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
}

// requireListenLoopback checks if the port(s) listen on the loopback interface
// It does not check whether port(s) listen on interfaces other than the loopback
func requireListenLoopback(t *testing.T, ports ...string) {
	if len(ports) == 0 {
		panic("No ports given as input")
	}

	for _, port := range ports {
		if !isListenInterface(t, "127.0.0.1", port) {
			t.Errorf("Port %s is not restricted to listen on loopback interface", port)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
}

// RequirePortAvailable checks if a port is available (not open) locally
func RequirePortAvailable(t *testing.T, port string) {
	stdout := lsof(t, port)
	if stdout != "" {
		t.Fatalf("Port %s is not available", port)
	}
	t.Logf("Port %s is available.", port)
}

func isListenInterface(t *testing.T, addr string, port string) bool {
	list := filterOpenPorts(t, port)

	// look for LISTEN explicitly to exclude ESTABLISHED connections
	substr := fmt.Sprintf("%s:%s (LISTEN)", addr, port)
	t.Logf("Looking for '%s'", substr)

	return strings.Contains(list, substr)
}

func filterOpenPorts(t *testing.T, port string) string {
	stdout := lsof(t, port)
	if stdout == "" {
		t.Fatalf("Port %s is not open", port)
	}
	return stdout
}

func lsof(t *testing.T, port string) string {
	// The chained true command is to make sure execution succeeds even if
	// 	the first command fails when list is empty
	stdout, _, _ := Exec(t, fmt.Sprintf("sudo lsof -nPi :%s || true", port))
	return stdout
}
