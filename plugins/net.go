package plugins

import (
	"fmt"
	"github.com/chenyf/goeos/p2p"
	"io"
	"net"
)

type MsgHandler func(*net.TCPConn, *Peer, *p2p.Header, []byte) int

type NetPlugin struct {
	node_id         string
	chain_id        string
	network_version string

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
	peer, ret := HandleHandshake(this, conn)
	if ret != 0 {
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
	status               uint32
	syncing              bool
	node_id              string
	sent_handshake_count int
}

func (this *Peer) send_handshake() {
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

type HandshakeMsg struct {
	generation      int
	node_id         string
	chain_id        string
	head_id         string
	network_version string
	head_num        int
}

func HandleHandshake(plugin *NetPlugin, conn *net.TCPConn) (*Peer, int) {
	var msg HandshakeMsg
	var peer Peer
	if msg.generation == 1 {
		if msg.node_id == plugin.node_id {
			return nil, 1
		}

		if msg.chain_id != plugin.chain_id {
			return nil, 2
		}

		if msg.network_version != plugin.network_version {
		}
		if peer.node_id != msg.node_id {
			peer.node_id = msg.node_id
		}
		if peer.sent_handshake_count == 0 {
			peer.send_handshake()
		}
	}

	peer.syncing = false
	var cc ChainController
	head_num := cc.head_block_num()
	head_id := cc.head_block_id()
	if head_id == msg.head_id {
	}
	if head_num < 0 {
	}
	if msg.head_num < 100 {
	}
	if head_num <= msg.head_num {
		return nil, 0
	}
	if msg.generation > 1 {
		/*var NoticeMsg note
		peer.enqueue(note)*/
	}
	peer.syncing = true
	return nil, 0
}

func HandleGoAway(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}

func HandleTime(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}

func HandleNotice(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}
func HandleRequest(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}
func HandleSyncRequest(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}
func HandleSignedBlockSummary(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}
func HandleSignedBlock(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}
func HandlePackedTransaction(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}
func HandleSignedTransaction(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}
