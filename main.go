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
	"flag"
//	"io/ioutil"
)

type Node struct {
	Hostname	string	`json:"hostname"`
	Extip		net.IP	`json:"extip"`
	Intip		net.IP	`json:"intip"`
	Port		string	`json:"port"`
}

type PeerData struct {
	Hostname	string	`json:"hostname"`
	Intip		net.IP	`json:"intip"`
	Extip		net.IP	`json:"extip"`
	Port		string	`json:"port"`
	Remoteip	net.IP	`json:"remoteip"`
	Nodes		[]Node	`json:"nodes"`
}

var (
	Trace		*log.Logger
	Info		*log.Logger
	Warning		*log.Logger
	Error		*log.Logger
	Nodes		PeerData
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
	//var node Node
	var err error

	re := regexp.MustCompile("\\.")
	hostname, _ := os.Hostname()
	match := re.Split(hostname, -1)
	Nodes.Hostname = string(match[0])
	if err != nil {
		panic(err)
	}
	Info.Println("my hostname:", Nodes.Hostname)

	Nodes.Intip = GetLocalIP()
	Info.Println("local IPv4:", Nodes.Intip)

	/*node.Hostname = Nodes.Hostname
	node.Intip = Nodes.Intip
	*/
	cf := flag.String("join","", "")
	tcpPort := flag.String("port", "9080", "")
	flag.Parse()

	Nodes.Port = *tcpPort
	//Nodes.Nodes = append(Nodes.Nodes, node)

	re = regexp.MustCompile("\\d+(\\.\\d+){3}")
	if cf != nil && re.MatchString(*cf) {
		Join2cluster(*cf)
	}
	Loadkey()
	APIServer := NewAPIServer()
	Info.Println("listen port:", Nodes.Port)
	log.Fatal(http.ListenAndServe(":" + Nodes.Port, APIServer))
}
