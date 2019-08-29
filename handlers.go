package main

import (
        "net"
        "net/http"
        "github.com/julienschmidt/httprouter"
	"encoding/json"
	"bytes"
	//"io/ioutil"
	//"fmt"
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
	var node	Node

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("New join request from host:", remoteIP)
	myExt, _, _ := net.SplitHostPort(r.Host)
	Trace.Println("myExt is:", myExt)
	Inventory.Extip = net.ParseIP(myExt)
	err := json.NewDecoder(r.Body).Decode(&node)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
	Trace.Println("Join from node:", node)
	node.Extip = net.ParseIP(remoteIP)
	for i, _ := range Inventory.Nodes {
		n := &Inventory.Nodes[i]
		if bytes.Equal(n.Extip, node.Extip) && n.Port == node.Port {
			Info.Println("the node:", node.Hostname, node.Extip, node.Port, "is already known")
			Info.Println("update data for node with external IP:", node.Extip)
			n.Intip = node.Intip
			//UpdateIKE()
			//return
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
	sswanUpdateIKE()
}


