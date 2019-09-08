package main

import (
	"fmt"

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
	conn := "to_" + hostname
	m.Set("child", conn)
	m.Set("ike", conn)
	m.Set("timeout", "-1")
	Info.Println("initiate conn:", conn)
	r, err := sswanSock.CommandRequest("initiate", m)
	if err != nil {
		Error.Println("initiate error:", err)
	}
	Info.Println("initiate conn", conn, "success:", r)
}

func sswanTerminateConn(hostname string) {
	m := vici.NewMessage()
	conn := "to_" + hostname
	m.Set("child", conn)
	m.Set("ike", conn)
	m.Set("timeout", "-1")
	Info.Println("terminate conn:", conn)
	r, err := sswanSock.CommandRequest("terminate", m)
	if err != nil {
		Error.Println("terminate error:", err)
	}
	Info.Println("terminate conn", conn, "success:", r)
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
	conn := "to_" + hostname
	m.Set("name", conn)
	Info.Println("unload conn:", conn)
	r, err := sswanSock.CommandRequest("unload-conn", m)
	if err != nil {
		Error.Println("unload error:", err)
	}
	Info.Println("unload conn result:", r)
}

func sswanLoadConn() {
	myIntip := fmt.Sprint(Inventory.Intip)
	myExtip := fmt.Sprint(Inventory.Extip)
	for _, n := range Inventory.Nodes {
		nodeIntip := fmt.Sprint(n.Intip)
		nodeExtip := fmt.Sprint(n.Extip)
		local := ikeLocal{myExtip, "psk"}
		remote := ikeRemote{nodeExtip, "psk"}

		lts := []string{myIntip}
		rts := []string{nodeIntip}
		ep := []string{"aes256-sha2_256"}
		child := ikeChildren{lts, rts, ep, "clear", "start", "1h"}
		mChild, err := vici.MarshalMessage(child)
		if err != nil {
			Error.Println(err)
		}
		cName := "to_" + n.Hostname
		childName := vici.NewMessage()
		childName.Set(cName, mChild)

		la := []string{myIntip}
		ra := []string{nodeExtip}
		ps := []string{"aes256-sha2_256-modp1024", "default"}
		ike := ike{la, ra, ps, local, remote, "replace", "600", childName}
		mIke, e := vici.MarshalMessage(ike)

		c := vici.NewMessage()
		c.Set(cName, mIke)
		check := c.Err()
		if check != nil {
			Error.Println(check)
		}
		m, e := sswanSock.CommandRequest("load-conn", c)
		if e != nil {
			Error.Println(e)
		}
		Info.Println("connection", cName, "loaded:", m.Get("success"))
	}
}
