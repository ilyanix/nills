package main

import (
	"log"
	"net"
	"net/http"
	"io"
	"os"
	"bufio"
	"regexp"
	"fmt"
)

type Node struct {
	Hostname string `json:"hostname"`
	Extip net.IP `json:"extip"`
	Intip net.IP `json:"intip"`
}

var (
	Trace		*log.Logger
	Info		*log.Logger
	Warning		*log.Logger
	Error		*log.Logger
	Nodes		[]Node
	MyHostname	string
)

func Loginit(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Trace = log.New(traceHandle,"TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func GetLocalIP() net.IP {
	re := regexp.MustCompile("^(\\w+)\\s+00000000(\\s+\\w+){5}\\s+00000000")
	var ipv4 net.IP

	file, _ := os.Open("/proc/net/route")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := re.FindStringSubmatch(scanner.Text())
		if match != nil {
			iface, _ := net.InterfaceByName(match[1])
			addrs, _ := iface.Addrs()
			Trace.Println("localIP addresses:", addrs)
			ipv4, _, _ = net.ParseCIDR(fmt.Sprint(addrs[0]))
		}
	}
	return ipv4
}
func main() {
	Loginit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	var node Node
	var err error

	MyHostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
	Info.Println("my hostname:", MyHostname)

	myIP := GetLocalIP()
	Info.Println("local IPv4:", myIP)

	node.Hostname = MyHostname
	node.Intip = myIP
	Nodes = append(Nodes, node)

	TCPPort := "9080"

	APIServer := NewAPIServer()
	Info.Println("listen port:", TCPPort)
	log.Fatal(http.ListenAndServe(":" + TCPPort, APIServer))
}
