package main

import (
	"fmt"
        "net"
        "net/http"
        "github.com/julienschmidt/httprouter"
)

func Join(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	Trace.Println("Join Start")
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	Info.Println("New join request from host:", remoteIP)
	fmt.Fprintln(w, "your IP address is:", remoteIP)
}
