package main

import (
	vici "github.com/strongswan/govici"
)

var (
	sswanSock *vici.Session
)

type ikeLocal struct {
	ID   string `vici:"id"`
	Auth string `vici:"auth"`
}

type ikeRemote struct {
	ID   string `vici:"id"`
	Auth string `vici:"auth"`
}

type ikeChildren struct {
	LocalTs      []string `vici:"local_ts"`
	RemotelTs    []string `vici:"remote_ts"`
	EspProposals []string `vici:"esp_proposals"`
	Dpdaction    string   `vici:"dpd_action"`
	Startaction  string   `vici:"start_action"`
	Lifetime     string   `vici:"life_time"`
}

type ike struct {
	LocalAddrs  []string      `vici:"local_addrs"`
	RemoteAddrs []string      `vici:"remote_addrs"`
	Proposals   []string      `vici:"proposals"`
	Local       ikeLocal      `vici:"local"`
	Remote      ikeRemote     `vici:"remote"`
	Unique      string        `vici:"unique"`
	ReauthTime  string        `vici:"reauth_time"`
	Children    *vici.Message `vici:"children"`
}

type ikeKey struct {
	ID     string   `vici:"id"`
	Type   string   `vici:"type"`
	Data   string   `vici:"data"`
	Owners []string `vici:"owners"`
}

func sswanSession() {
	s, err := vici.NewSession()
	if err != nil {
		panic(err)
	}
	sswanSock = s
}

func sswanInitConn(hostname string) {
	m := vici.NewMessage()
	m.Set("child", "childTo_"+hostname)
	m.Set("ike", "to_"+hostname)
	//m.Set("timeout", "1000")
	//m.Set("init-limits", 1)
	Info.Println("initiate conn to host:", hostname)
	r, err := sswanSock.CommandRequest("initiate", m)
	if err != nil {
		Error.Println("initiate error:", err)
	}
	Info.Println("initiate conn to host:", hostname, "success:", r)
}

func sswanTerminateConn(hostname string) {
	m := vici.NewMessage()
	m.Set("child", "childTo_"+hostname)
	m.Set("ike", "to_"+hostname)
	m.Set("timeout", "0")
	Info.Println("terminate conn to host:", hostname)
	r, err := sswanSock.CommandRequest("terminate", m)
	if err != nil {
		Error.Println("terminate error:", err)
	}
	Info.Println("terminate conn to host:", hostname, "success:", r)
}

func sswanLoadKey() {
	var psk ikeKey
	psk.ID = "qwe"
	psk.Type = "ike"
	psk.Data = "supersecretkey12345678"

	m, err := vici.MarshalMessage(psk)
	if err != nil {
		Error.Println("message key:", err)
	}
	r, err := sswanSock.CommandRequest("load-shared", m)
	if err != nil {
		Error.Println("lad key:", err)
	}
	Info.Println("load key success:", r.Get("success"))
}

func sswanUnloadConn(hostname string) {
	m := vici.NewMessage()
	m.Set("name", "to_"+hostname)
	Info.Println("unload conn to hostname:", hostname)
	r, err := sswanSock.CommandRequest("unload-conn", m)
	if err != nil {
		Error.Println("unload error:", err)
	}
	Info.Println("unload conn result:", r)
}

func sswanLoadConn(hostname string) {
	n := Inventory.Nodes[hostname]

	local := ikeLocal{Inventory.Extip[0], "psk"}
	remote := ikeRemote{n.Extip[0], "psk"}
	lts := Inventory.Intip
	rts := n.Intip
	ep := []string{"aes256-sha2_256"}

	child := ikeChildren{lts, rts, ep, "clear", "none", "1h"}
	mChild, err := vici.MarshalMessage(child)
	if err != nil {
		Error.Println(err)
	}

	childName := vici.NewMessage()
	childName.Set("childTo_"+n.Hostname, mChild)
	Trace.Println("sswan child messages:", mChild)

	la := Inventory.Intip
	ra := n.Extip
	ps := []string{"aes256-sha2_256-modp1024", "default"}

	ike := ike{la, ra, ps, local, remote, "no", "0", childName}
	mIke, e := vici.MarshalMessage(ike)
	if err != nil {
		Error.Println(err)
	}
	Trace.Println("sswan ike messages:", ike)

	c := vici.NewMessage()
	c.Set("to_"+hostname, mIke)

	check := c.Err()
	if check != nil {
		Error.Println(check)
	}

	m, e := sswanSock.CommandRequest("load-conn", c)
	if e != nil {
		Error.Println(e)
	}

	Info.Println("connection to host:", hostname, "loaded:", m)
}
