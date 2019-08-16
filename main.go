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

var (
	Trace		*log.Logger
	Info		*log.Logger
	Warning		*log.Logger
	Error		*log.Logger
)

func Loginit(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Trace = log.New(traceHandle,"TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func GetLocalIP() net.Addr {
	re := regexp.MustCompile("^(\\w+)\\s+00000000(\\s+\\w+){5}\\s+00000000")
	var localip net.Addr

	file, _ := os.Open("/proc/net/route")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := re.FindStringSubmatch(scanner.Text())
		if match != nil {
			iface, _ := net.InterfaceByName(match[1])
			addrs, _ := iface.Addrs()
			Trace.Println("localIP addresses:", addrs)
			ip := fmt.Sprint(addrs[0])
			ipv4, prefix, _ := net.ParseCIDR(ip)
			Trace.Println("local ipv4:", ipv4, "Prefix:", prefix)
		}
	}
	return localip
}
func main() {
	Loginit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	localip := GetLocalIP()
	Info.Println("local IPv4:", localip)
	Trace.Println("Create API router")
	APIServer := NewAPIServer()
	Trace.Println("Start WeB API Server")
	log.Fatal(http.ListenAndServe(":9080", APIServer))
}
