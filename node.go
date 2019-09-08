package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

//Node data about node
type Node struct {
	Hostname string   `json:"hostname"`
	Extip    []string `json:"extip"`
	Intip    []string `json:"intip"`
	Port     string   `json:"port"`
}

//PeerInventory all nodes data
type PeerInventory struct {
	Hostname string   `json:"hostname"`
	Intip    []string `json:"intip"`
	Extip    []string `json:"extip"`
	Port     string   `json:"port"`
	Remoteip string   `json:"remoteip"`
	Nodes    []Node   `json:"nodes"`
}

func getHostname() string {
	re := regexp.MustCompile("\\.")
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	Trace.Println("os hostname is:", hostname)
	match := re.Split(hostname, -1)
	if len(match) > 1 {
		hostname = match[0]
	}
	Info.Println("my hostname:", hostname)
	return hostname
}

func nodeResolveTarget(target string) string {
	dstip, port, _ := net.SplitHostPort(target)
	re := regexp.MustCompile("\\d+(\\.\\d+)")
	if !re.MatchString(dstip) {
		ip, err := net.LookupHost(dstip)
		if err != nil {
			Error.Println("can't resolve name:", dstip)
		}
		dstip = ip[0]
	}

	Trace.Println("destination is:", dstip, "and port:", port)
	return net.JoinHostPort(dstip, port)
}

func getLocalIP(ifname string) string {
	var ip []net.Addr

	re := regexp.MustCompile("^\\d+$")
	if re.MatchString(ifname) {
		ifindex, err := strconv.Atoi(ifname)
		if err != nil {
			Error.Println(err)
		}
		iface, err := net.InterfaceByIndex(ifindex)
		if err != nil {
			Error.Println("can't find interface with index:", ifindex)
		}
		ip, _ = iface.Addrs()
	} else {
		iface, err := net.InterfaceByName(ifname)
		if err != nil {
			Error.Println("can't find interface name:", ifname)
		}
		ip, _ = iface.Addrs()
	}
	Trace.Println("local ip addresses:", ip)
	ipv4, netv4, err := net.ParseCIDR(fmt.Sprint(ip[0]))
	if err != nil {
		Error.Println(err)
	}
	Trace.Println("local IPv4 network:", netv4)
	Info.Println("local IPv4 addres:", ipv4)
	return fmt.Sprint(ipv4)
}

func nodeCollectData(ifname string) {
	Inventory.Hostname = getHostname()
	Inventory.Intip = append(Inventory.Intip, getLocalIP(ifname))
	Inventory.Extip = []string{"0.0.0.0"}
}

func nodeJoin2cluster(host string) {
	var node Node
	var rejoin bool

	rInventory := getNodes(host)
	Trace.Println("cluster state:", rInventory)
	for _, v := range rInventory.Nodes {
		switch {
		case v.Hostname != Inventory.Hostname:
			Trace.Println("add node:", v.Hostname)
			Inventory.Nodes = append(Inventory.Nodes, v)
		case v.Hostname == Inventory.Hostname:
			Info.Println("rejoin to cluster")
			rejoin = true
		}
	}

	node.Hostname = rInventory.Hostname
	node.Port = rInventory.Port
	node.Extip = rInventory.Extip
	node.Intip = rInventory.Intip

	Inventory.Extip = append(Inventory.Extip, rInventory.Remoteip)
	Trace.Println("add node:", node.Hostname)
	Inventory.Nodes = append(Inventory.Nodes, node)

	for _, v := range Inventory.Nodes {
		Trace.Println("load connection to:", v.Hostname)
		sswanLoadConn(v.Hostname)
		Trace.Println("join to node:", v.Hostname)
		ip := v.Extip[0]
		host := net.JoinHostPort(ip, v.Port)
		postJoin(host)
	}

	if rejoin {
		Info.Println("rejoin to all nodes:")
		for _, n := range Inventory.Nodes {
			sswanInitConn(n.Hostname)
		}
		rejoin = false
	}
}

func getNodeWipe(host string, node string) {

	url := "http://" + host + "/v1/nodewipe/" + node
	Info.Println("Trying to connect to:", url)

	resp, err := http.Get(url)
	if err != nil {
		Error.Println(err)
	}
	Trace.Println("getwipe resp:", resp)
}

func getNodes(host string) PeerInventory {
	var r PeerInventory

	url := "http://" + host + "/v1/nodes"
	Info.Println("Trying to connect to:", url)

	resp, err := http.Get(url)
	if err != nil {
		Error.Println(err)
		return r
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Error.Println(err)
		return r
	}
	Trace.Println("getNodes body:", body)
	err = json.Unmarshal(body, &r)
	if err != nil {
		Error.Println(err)
		return r
	}
	Trace.Println("getnodes res:", r)
	return r
}

func postJoin(host string) PeerInventory {
	var r PeerInventory

	url := "http://" + host + "/v1/join"
	Info.Println("join to url:", url)
	data, err := json.Marshal(Inventory)

	Trace.Println("join post data:", bytes.NewBuffer(data))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		Error.Println(err)
	}
	req.Header.Set("Content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Error.Println(err)
		return r
	}
	Info.Println("response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Error.Println(err)
	}
	Trace.Println("response body:", string(body))
	err = json.Unmarshal(body, &r)
	if err != nil {
		Error.Println(err)
	}
	return r
}
