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
	Hostname string          `json:"hostname"`
	Intip    []string        `json:"intip"`
	Extip    []string        `json:"extip"`
	Port     string          `json:"port"`
	Remoteip string          `json:"remoteip"`
	Nodes    map[string]Node `json:"nodes"`
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
	/*pv4, netv4, err := net.ParseCIDR(fmt.Sprint(ip[0]))
	if err != nil {
		Error.Println(err)
	}
	Trace.Println("local IPv4 network:", netv4)*/
	ipv4 := fmt.Sprint(ip[0])
	Info.Println("local IPv4 addres:", ipv4)
	return fmt.Sprint(ipv4)
}

func nodeCollectData(ifname string) {
	Inventory.Hostname = getHostname()
	Inventory.Intip = []string{getLocalIP(ifname)}
	Inventory.Extip = []string{"0.0.0.0"}
	Inventory.Nodes = make(map[string]Node)
}

func nodeJoin2cluster(host string) {
	var node Node

	rInventory := getNodes(host)
	Trace.Println("cluster state:", rInventory)
	for i := range rInventory.Nodes {
		n := rInventory.Nodes[i]
		switch {
		case n.Hostname != Inventory.Hostname:
			Trace.Println("add node:", n.Hostname)
			Inventory.Nodes[i] = n
		case n.Hostname == Inventory.Hostname:
			Info.Println("rejoin to cluster")
		}
	}

	node.Hostname = rInventory.Hostname
	node.Port = rInventory.Port
	node.Extip = rInventory.Extip
	node.Intip = rInventory.Intip

	Inventory.Extip[0] = rInventory.Remoteip
	Trace.Println("add node:", node.Hostname)
	Inventory.Nodes[node.Hostname] = node

	for i := range Inventory.Nodes {
		n := Inventory.Nodes[i]
		if !nodeCheckLocalNet(n) {
			Trace.Println("load connection to:", n.Hostname)
			sswanLoadConn(n.Hostname)
		}
		Trace.Println("join to node:", n.Hostname)
		ip := n.Extip[0]
		host := net.JoinHostPort(ip, n.Port)
		postJoin(host)
	}
}

func nodeCheckLocalNet(n Node) bool {
	Trace.Println("compare loacal network for:", Inventory.Intip, "and", n.Intip)
	_, myNet, _ := net.ParseCIDR(Inventory.Intip[0])
	_, peerNet, _ := net.ParseCIDR(n.Intip[0])
	Trace.Println("compare loacal network for:", myNet, "and", peerNet)
	if fmt.Sprint(myNet) == fmt.Sprint(peerNet) {
		Info.Println("hostname:", n.Hostname, "is in my broadcast network")
		return true
	}
	return false
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
