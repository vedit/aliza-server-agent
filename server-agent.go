package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vedit/proxy"
)

var localAddr1 = flag.String("l1", "0.0.0.0:6801", "local address")
var remoteAddr1 = flag.String("r1", "127.0.0.1:6801", "remote address")
var localAddr2 = flag.String("l2", "0.0.0.0:6901", "local address")
var remoteAddr2 = flag.String("r2", "127.0.0.1:6901", "remote address")

func main() {
	flag.Parse()
	ps1 := NewProxyServer(*localAddr1, *remoteAddr1)
	ps2 := NewProxyServer(*localAddr2, *remoteAddr2)
	go ps1.setup()
	go ps2.setup()
}

type ProxyServer struct {
	localAddr, remoteAddr string
}

func NewProxyServer(local string, remote string) *ProxyServer {
	return &ProxyServer{localAddr: local, remoteAddr: remote}
}

func (ps *ProxyServer) setup() {
	fmt.Printf("Proxying from %v to %v\n", ps.localAddr, ps.remoteAddr)

	for {
		p := proxy.NewProxy(ps.localAddr, ps.remoteAddr)
		go p.Start()
	}
}

func check(err error) {
	if err != nil {
		log(err.Error())
		os.Exit(1)
	}
}

func log(f string, args ...interface{}) {
	fmt.Printf(f+"\n", args...)
}
