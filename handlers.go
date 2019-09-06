package main

import (
	"fmt"
        "net"
        "net/http"
        "github.com/julienschmidt/httprouter"
	"encoding/json"
	"bytes"
)

func handlListNodes(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("nodes list request from host:", remoteIP)
	Inventory.Remoteip = net.ParseIP(remoteIP)
	myExt, _, _ := net.SplitHostPort(r.Host)
	Trace.Println("myExt is:", myExt)
	Inventory.Extip = net.ParseIP(myExt)
	err := json.NewEncoder(w).Encode(Inventory)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
}

func handlJoin(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	Trace.Println("Join Start")
	var rInventory	PeerInventory
	var node	Node

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("New join request from host:", remoteIP)
	myExt, _, _ := net.SplitHostPort(r.Host)
	Trace.Println("myExt is:", myExt)
	Inventory.Extip = net.ParseIP(myExt)
	err := json.NewDecoder(r.Body).Decode(&rInventory)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
	Trace.Println("Join data from node:", rInventory)
	node = Node{
		Hostname: rInventory.Hostname,
		Intip: rInventory.Intip,
		Extip: net.ParseIP(remoteIP),
		Port: rInventory.Port }

	for i, _ := range Inventory.Nodes {
		n := &Inventory.Nodes[i]
		if bytes.Equal(n.Extip, node.Extip) && n.Port == node.Port {
			Info.Println("the node:", node.Hostname, node.Extip, node.Port, "is already known")
			Info.Println("update data for node with external IP:", node.Extip)
			n.Intip = node.Intip
			sswanLoadConn()
			sswanTerminateConn(node.Hostname)
			sswanInitConn(node.Hostname)
			err = json.NewEncoder(w).Encode(Inventory)
			if err != nil {
				w.WriteHeader(500)
				Trace.Println(err)
				return
			}
			return
		}
	}
	Inventory.Nodes = append(Inventory.Nodes, node)
	Trace.Println("nodes:", Inventory.Nodes)
	err = json.NewEncoder(w).Encode(Inventory)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
	sswanLoadConn()
}

func handlNodeShow(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	var res Node

	hostname := p.ByName("hostname")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	Info.Println("request info about node", hostname, "from", remoteIP)
	if hostname == Inventory.Hostname {
		res.Hostname = Inventory.Hostname
		res.Intip = Inventory.Intip
		res.Extip = Inventory.Extip
		res.Port = Inventory.Port
	}
	for i, _ := range Inventory.Nodes {
		n := Inventory.Nodes[i]
		if n.Hostname == hostname {
			res.Hostname = n.Hostname
			res.Intip = n.Intip
			res.Extip = n.Extip
			res.Port = n.Port
		}
	}
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
}

func handlNodeWipe(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	hostname := p.ByName("hostname")
	rIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	remoteIP := net.ParseIP(rIP)
	Trace.Println("wipe node:", hostname)
	if hostname == Inventory.Hostname {
		for _, v := range Inventory.Nodes {
			sswanTerminateConn(v.Hostname)
			sswanUnloadConn(v.Hostname)
			host := net.JoinHostPort(fmt.Sprint(v.Extip), v.Port)
			if !bytes.Equal(v.Extip, remoteIP) {
				getNodeWipe(host, hostname)
			}
		}
		Inventory.Nodes = nil
		return
	} else {
		for i, v := range Inventory.Nodes {
			if hostname == v.Hostname {
				sswanTerminateConn(v.Hostname)
				sswanUnloadConn(v.Hostname)
				Inventory.Nodes = append(Inventory.Nodes[:i], Inventory.Nodes[i+1:]...)
				if !bytes.Equal(v.Extip, remoteIP) {
					host := net.JoinHostPort(fmt.Sprint(v.Extip), v.Port)
					getNodeWipe(host, hostname)
				}
			}
		}

	}
}
