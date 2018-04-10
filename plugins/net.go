package plugins

import (
	"fmt"
	"github.com/chenyf/goeos/p2p"
	"io"
	"net"
)

type MsgHandler func(*net.TCPConn, *Peer, *p2p.Header, []byte) int

type NetPlugin struct {
	addr               string
	listener           *net.TCPListener
	connection_num     uint32
	max_connection_num uint32
	max_body_len       uint32

	handlerMap map[uint8]MsgHandler
}

func NewNetPlugin() *NetPlugin {
	return &NetPlugin{
		connection_num:     0,
		max_connection_num: 100,
		max_body_len:       20480,
	}
}

func (this *NetPlugin) Init() {
}

func (this *NetPlugin) listenLoop() {
	for {
		conn, err := this.listener.AcceptTCP()
		if err != nil {
			continue
		}
		if this.connection_num >= this.max_connection_num {
			conn.Close()
			continue
		}
		this.connection_num++
		go this.startSession(conn)
	}
}

func (this *NetPlugin) startSession(conn *net.TCPConn) {
	peer, err := this.handleInit(conn)
	if err != nil {
		conn.Close()
		return
	}

	var (
		headBuf []byte = make([]byte, p2p.HEADER_SIZE)
		dataBuf []byte
		header  p2p.Header
	)

	close_code := 0
	for {
		if peer.status > 0 {
			close_code = 1
			break
		}

		//now := time.Now()
		//conn.SetReadDeadline(now.Add(hbTimeout))

		// read header
		if _, err := io.ReadFull(conn, headBuf[0:]); err != nil {
			close_code = 3
			break
		}
		if err := header.Deserialize(headBuf[0:p2p.HEADER_SIZE]); err != nil {
			close_code = 4
			break
		}
		if header.Len > this.max_body_len {
			close_code = 5
			break
		}
		if header.Len != 0 {
			// read body
			dataBuf = make([]byte, header.Len)
			if _, err := io.ReadFull(conn, dataBuf); err != nil {
				close_code = 6
				break
			}
		} else {
			dataBuf = nil
		}

		// handle message
		handler, ok := this.handlerMap[header.Type]
		if !ok {
			continue
		}
		handler(conn, peer, &header, dataBuf)
	}
	fmt.Printf("close connection, reason %d\n", close_code)
	this.closePeer(peer)
}

func (this *NetPlugin) closePeer(peer *Peer) {
}

type Peer struct {
	status uint32
}

func (this *NetPlugin) handleInit(conn *net.TCPConn) (*Peer, error) {
	return nil, nil
}

func (this *NetPlugin) Startup() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", this.addr)
	l, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		return err
	}
	this.listener = l

	go this.listenLoop()

	//this.start_monitors()
	/*for _, seed := range this.supplied_peers {
		//connect(seed)
	}*/
	return nil
}

func (this *NetPlugin) Shutdown() {
}

func (this *NetPlugin) Name() string {
	return "network plugin"
}
