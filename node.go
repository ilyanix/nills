package main

import (
	"regexp"
	"net"
	"net/http"
	"os"
	"bufio"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"bytes"
)

func getHostname() string {
        re := regexp.MustCompile("\\.")
        hostname, err := os.Hostname()
        match := re.Split(hostname, -1)
        if err != nil {
                panic(err)
        }
        Info.Println("my hostname:", Inventory.Hostname)
	return string(match[0])
}

func getLocalIP() net.IP {
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
	Trace.Println("local IPv4:", ipv4)
        return ipv4
}

func nodeCollectData() {
	Inventory.Hostname = getHostname()
	Inventory.Intip = getLocalIP()
}

func nodeJoin2cluster(host string) {
	var node Node
	var nodes []string
	var rejoin bool

	rInventory := getNodes(host)
	Trace.Println("cluster state:", rInventory)
	for _, v := range rInventory.Nodes {
		if v.Hostname != Inventory.Hostname {
			Trace.Println("add node:", v.Hostname)
			Inventory.Nodes = append(Inventory.Nodes, v)
		}
	}

	node.Hostname = rInventory.Hostname
	node.Port = rInventory.Port
	node.Extip = rInventory.Extip
	node.Intip = rInventory.Intip

	Inventory.Extip = rInventory.Remoteip
	Trace.Println("add node:", node.Hostname)
	rInventory.Nodes = append(rInventory.Nodes, node)
	nodes = append(nodes, node.Hostname)

	for _, v := range rInventory.Nodes {
		Trace.Println("loop nodes:", v.Hostname, "vs", Inventory.Hostname)
		switch {
		case v.Hostname != Inventory.Hostname:
			Trace.Println("join to node:", v.Hostname)
			ip := fmt.Sprint(v.Extip)
			host := net.JoinHostPort(ip, v.Port)
			nodes = append(nodes, v.Hostname)
			postJoin(host)
                        Inventory.Nodes = append(Inventory.Nodes, v)
		case v.Hostname == Inventory.Hostname:
			Info.Println("rejoin to cluster")
			rejoin = true
		}
	}
	Trace.Println("load connection", Inventory.Nodes)
	sswanLoadConn()
	if rejoin {
		Info.Println("rejoin to nodes:", nodes)
		sswanInitConn(nodes)
		rejoin = false
	}
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
		//r := PeerData{}
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


