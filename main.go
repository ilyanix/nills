package main

import (
	"log"
	"net"
	"net/http"
	"io"
	"os"
//	"bufio"
	"regexp"
//	"fmt"
	"flag"
//	"io/ioutil"
)

type Node struct {
	Hostname	string	`json:"hostname"`
	Extip		net.IP	`json:"extip"`
	Intip		net.IP	`json:"intip"`
	Port		string	`json:"port"`
}

type PeerInventory struct {
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
	Inventory	PeerInventory
)

func Loginit(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Trace = log.New(traceHandle,"TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}


func main() {
	Loginit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	joinFlag := flag.String("join","", "join to: IP:PORT")
	tcpPort := flag.String("port", "9080", "bind port, default 9080")
	flag.Parse()

	Inventory.Port = *tcpPort
	nodeCollectData()

	sswanLoadKey()

	re := regexp.MustCompile("\\d+(\\.\\d+){3}")
	if joinFlag != nil && re.MatchString(*joinFlag) {
		nodeJoin2cluster(*joinFlag)
	}


	APIServer := NewAPIServer()
	Info.Println("listen port:", Inventory.Port)
	log.Fatal(http.ListenAndServe(":" + Inventory.Port, APIServer))
}
