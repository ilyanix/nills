package main

import (
        "net"
        "net/http"
        "github.com/julienschmidt/httprouter"
	"encoding/json"
	"bytes"
)

func ListNodes(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var err		error
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("nodes list request from host:", remoteIP)
	err = json.NewEncoder(w).Encode(Nodes)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
	Trace.Println("host:", r.Host)
}

func Join(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	Trace.Println("Join Start")
	var node	Node
	var err		error

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("New join request from host:", remoteIP)
	myExt, _, _ := net.SplitHostPort(r.Host)
	for i, _ := range Nodes {
		n := &Nodes[i]
		if n.Hostname == MyHostname {
			ip := net.ParseIP(myExt)
			n.Extip = ip
		}
	}
	err = json.NewDecoder(r.Body).Decode(&node)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
	Trace.Println("nodes:", Nodes)
	err = json.NewEncoder(w).Encode(Nodes)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
	for i, _ := range Nodes {
		n := &Nodes[i]
		if bytes.Equal(n.Extip, node.Extip) {
			Info.Println("node with IP:", node.Extip, "already known")
			n.Hostname = node.Hostname
			n.Intip = node.Intip
			return
		}
	}
	Nodes = append(Nodes, node)
}
