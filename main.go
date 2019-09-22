package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	//Trace logs
	Trace *log.Logger
	//Info logs
	Info *log.Logger
	//Warning logs
	Warning *log.Logger
	//Error logs
	Error *log.Logger
	//Inventory data about all peeers
	Inventory PeerInventory
	//LanEnc Enc traffic between nodes in same subnet
	lanEnc bool
)

func loginit(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Trace = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	target := flag.String("join", "0.0.0.0:0", "join to: IP:PORT")
	tcpPort := flag.String("port", "9080", "pecify binding port")
	debug := flag.Bool("debug", false, "turn on debug logs")
	srcIfname := flag.String("nic", "eth0", "source interface name or index")
	localEnc := flag.Bool("local", false, "encript traffic between peers in same subnet")

	flag.Parse()

	if *debug {
		loginit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		loginit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}

	lanEnc = *localEnc
	Inventory.Port = *tcpPort

	peer := nodeResolveTarget(*target)
	nodeCollectData(*srcIfname)

	sswanSession()

	sswanLoadKey()

	if peer != "0.0.0.0:0" {
		nodeJoin2cluster(peer)
	}

	APIServer := newAPIServer()
	Info.Println("listen port:", Inventory.Port)
	log.Fatal(http.ListenAndServe(":"+Inventory.Port, APIServer))
}
