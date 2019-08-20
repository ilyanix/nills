package main

import (
        "net"
        "net/http"
        "github.com/julienschmidt/httprouter"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"fmt"
)

func ListNodes(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("nodes list request from host:", remoteIP)
	myExt, _, _ := net.SplitHostPort(r.Host)
	Trace.Println("myExt is:", myExt)
	Nodes.Extip = net.ParseIP(myExt)
	err := json.NewEncoder(w).Encode(Nodes)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
}

func Join(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	Trace.Println("Join Start")
	var node	Node
	var err		error

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("New join request from host:", remoteIP)
	myExt, _, _ := net.SplitHostPort(r.Host)
	Trace.Println("myExt is:", myExt)
	Nodes.Extip = net.ParseIP(myExt)
	err = json.NewDecoder(r.Body).Decode(&node)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}

	node.Extip = net.ParseIP(remoteIP)
	for i, _ := range Nodes.Nodes {
		n := &Nodes.Nodes[i]
		if bytes.Equal(n.Extip, node.Extip) && n.Port == node.Port {
			Info.Println("the node:", node.Hostname, node.Extip, node.Port, "is already known")
			Info.Println("update data for node with external IP:", node.Extip)
			n.Intip = node.Intip
			return
		}
	}

	Nodes.Nodes = append(Nodes.Nodes, node)
	Trace.Println("nodes:", Nodes.Nodes)
	err = json.NewEncoder(w).Encode(Nodes)
	if err != nil {
		w.WriteHeader(500)
		Trace.Println(err)
		return
	}
}

func getNodes(host string) PeerData {
	var r PeerData

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


func postJoin(host string) PeerData {
	var r PeerData

	url := "http://" + host + "/v1/join"
	Info.Println("join to node:", url)
	data, err := json.Marshal(Nodes)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		Error.Println(err)
	}
	req.Header.Set("Content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		r := PeerData{}
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

func Join2cluster(host string) {
	var node Node

	nodes := getNodes(host)
	Trace.Println("clusterNodes:", nodes)
	Nodes.Nodes = append(Nodes.Nodes, nodes.Nodes...)

	node.Hostname = nodes.Hostname
	node.Port = nodes.Port
	node.Extip = nodes.Extip
	node.Intip = nodes.Intip
	Nodes.Nodes = append(Nodes.Nodes, node)

	for _, v := range Nodes.Nodes {
		ip := fmt.Sprint(v.Extip)
		host := net.JoinHostPort(ip, v.Port)
		postJoin(host)
	}
}
