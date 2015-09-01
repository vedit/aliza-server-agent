package proxy

import (
	"fmt"
	"io"
	"os"

	"net"
)

var connid = uint64(0)

//A Proxy represents a pair of connections and their state
type Proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  *net.TCPAddr
	lconn, rconn  *net.TCPConn
	erred         bool
	errsig        chan bool
}

func NewProxy(localAddr, remoteAddr string) *Proxy {
	laddr, err := net.ResolveTCPAddr("tcp", localAddr)
	check(err)
	raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	check(err)
	listener, err := net.ListenTCP("tcp", laddr)
	check(err)
	conn, err := listener.AcceptTCP()
	if err != nil {
		fmt.Printf("Failed to accept connection '%s'\n", err)
	}
	return &Proxy{
		lconn:  conn,
		laddr:  laddr,
		raddr:  raddr,
		erred:  false,
		errsig: make(chan bool),
	}
}

func (p *Proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		log(s, err)
	}
	p.errsig <- true
	p.erred = true
}

func (p *Proxy) Start() {
	defer p.lconn.Close()
	//connect to remote
	rconn, err := net.DialTCP("tcp", nil, p.raddr)
	if err != nil {
		p.err("Remote connection failed: %s", err)
		return
	}
	p.rconn = rconn
	defer p.rconn.Close()

	//display both ends
	// p.log("Opened %s >>> %s", p.lconn.RemoteAddr().String(), p.rconn.RemoteAddr().String())
	//bidirectional copy
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)
	//wait for close...
	<-p.errsig
	// p.log("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
}

func (p *Proxy) pipe(src, dst *net.TCPConn) {
	//data direction
	islocal := src == p.lconn

	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b := buff[:n]
		//show output
		n, err = dst.Write(b)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
		if islocal {
			p.sentBytes += uint64(n)
		} else {
			p.receivedBytes += uint64(n)
		}
	}
}

//helper functions

func check(err error) {
	if err != nil {
		log(err.Error())
		os.Exit(1)
	}
}

func log(f string, args ...interface{}) {
	fmt.Printf(f+"\n", args...)
}
