package main

import (
	"log"
	"net/http"
	"io"
	"os"
	"flag"
	"io/ioutil"
)

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
	joinTarget := flag.String("join","0.0.0.0:0", "join to: IP:PORT")
	tcpPort := flag.String("port", "9080", "pecify binding port")
	debug := flag.Bool("debug", false, "turn on debug logs")
	srcIfname := flag.String("nic", "eth0", "source interface name or index")

	flag.Parse()

	if *debug {
		Loginit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		Loginit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}

	Inventory.Port = *tcpPort

	target := nodeResolveTarget(*joinTarget)
	nodeCollectData(*srcIfname)

	sswanLoadKey()

	if target != "0.0.0.0:0" {
		nodeJoin2cluster(target)
	}

	APIServer := NewAPIServer()
	Info.Println("listen port:", Inventory.Port)
	log.Fatal(http.ListenAndServe(":" + Inventory.Port, APIServer))
}
