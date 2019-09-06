package main

import (
	"fmt"
	"github.com/strongswan/govici"
)

type IKELocal struct {
	Id	string	`vici:"id"`
	Auth	string	`vici:"auth"`
}

type IKERemote struct {
	Id	string	`vici:"id"`
	Auth	string	`vici:"auth"`
}

type IKEChildren struct {
	Local_ts	[]string	`vici:"local_ts"`
	Remote_ts	[]string	`vici:"remote_ts"`
	Esp_proposals	[]string	`vici:"esp_proposals"`
	Dpd_action	string		`vici:"dpd_action"`
	Start_action	string		`vici:"start_action"`
	Life_time	string		`vici:"life_time"`
}


type IKE struct {
	Local_addrs	[]string	`vici:"local_addrs"`
	Remote_addrs	[]string	`vici:"remote_addrs"`
	Proposals	[]string	`vici:"proposals"`
	Local		IKELocal	`vici:"local"`
	Remote		IKERemote	`vici:"remote"`
        Unique		string		`vici:"unique"`
	Reauth_time	string		`vici:"reauth_time"`
	Children        *vici.Message   `vici:"children"`
}

type IKEKey struct {
	Id		string		`vici:"id"`
	Type		string		`vici:"type"`
	Data		string		`vici:"data"`
	Owners		[]string	`vici:"owners"`
}

func sswanInitConn(hostname string) {
	session, err := vici.NewSession()
        if err != nil {
                Error.Println("can't connect to strongswan daemon:", err)
                return
        }
	m := vici.NewMessage()
	conn := "to_" + hostname
	m.Set("child", conn)
	m.Set("ike", conn)
	m.Set("timeout", "-1")
	Info.Println("initiate conn:", conn)
	r, err := session.CommandRequest("initiate", m)
	if err != nil {
		Error.Println("initiate error:", err)
	}
	Info.Println("initiate conn", conn, "success:", r)
}

func sswanTerminateConn(hostname string) {
	session, err := vici.NewSession()
        if err != nil {
                Error.Println("can't connect to strongswan daemon:", err)
                return
        }
	m := vici.NewMessage()
	conn := "to_" + hostname
	m.Set("child", conn)
	m.Set("ike", conn)
	m.Set("timeout", "-1")
	Info.Println("terminate conn:", conn)
	r, err := session.CommandRequest("terminate", m)
	if err != nil {
		Error.Println("terminate error:", err)
	}
	Info.Println("terminate conn", conn, "success:", r)
}
func sswanLoadKey() {
	session, err := vici.NewSession()
        if err != nil {
                Error.Println("can't connect to strongswan daemon:", err)
                return
        }
	var psk IKEKey
	psk.Id = "qwe"
	psk.Type = "IKE"
	psk.Data = "supersecretkey12345678"

	m, err := vici.MarshalMessage(psk)
	if err != nil {
		Error.Println("message key:", err)
	}
	r, err := session.CommandRequest("load-shared", m)
	if err != nil {
		Error.Println("lad key:", err)
	}
	Info.Println("load key success:", r.Get("success"))
}

func sswanUnloadConn(hostname string) {
	session, err := vici.NewSession()
        if err != nil {
                Error.Println("can't connect to strongswan daemon:", err)
                return
        }
	m := vici.NewMessage()
	conn := "to_" + hostname
	m.Set("name", conn)
	Info.Println("unload conn:", conn)
	r, err := session.CommandRequest("unload-conn", m)
	if err != nil {
		Error.Println("unload error:", err)
	}
	Info.Println("unload conn result:", r)
}

func sswanLoadConn() {
	session, err := vici.NewSession()
        if err != nil {
                Error.Println("can't connect to strongswan daemon:", err)
                return
        }

	myIntip := fmt.Sprint(Inventory.Intip)
	myExtip := fmt.Sprint(Inventory.Extip)
	for _, n := range Inventory.Nodes {
		nodeIntip := fmt.Sprint(n.Intip)
		nodeExtip := fmt.Sprint(n.Extip)
		local := IKELocal{myExtip, "psk"}
		remote := IKERemote{nodeExtip, "psk"}

		lts := []string{myIntip}
		rts := []string{nodeIntip}
		ep := []string{"aes256-sha2_256"}
		child := IKEChildren{lts, rts, ep, "clear", "start", "1h"}
		m_child, err := vici.MarshalMessage(child)
		if err != nil {
			Error.Println(err)
		}
		c_name := "to_" + n.Hostname
		child_name := vici.NewMessage()
		child_name.Set(c_name, m_child)

		la := []string{myIntip}
		ra := []string{nodeExtip}
		ps := []string{"aes256-sha2_256-modp1024" ,"default"}
		ike := IKE{la, ra, ps, local, remote, "replace", "600", child_name}
		m_ike, e := vici.MarshalMessage(ike)

		c := vici.NewMessage()
		c.Set(c_name, m_ike)
		check := c.Err()
		if check != nil {
			Error.Println(check)
		}
		m, e := session.CommandRequest("load-conn", c)
		if e != nil {
			Error.Println(e)
		}
		Info.Println("connection", c_name, "loaded:", m.Get("success"))
	}
}
