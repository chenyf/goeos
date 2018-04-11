package plugins

import (
	"fmt"
	"github.com/chenyf/goeos/p2p"
	"io"
	"net"
)

type MsgHandler func(*net.TCPConn, *Peer, *p2p.Header, []byte) int

type NetPlugin struct {
	name                  string
	node_id               string
	chain_id              string
	network_version       int16
	network_version_match bool

	addr               string
	listener           *net.TCPListener
	connection_num     uint32
	max_connection_num uint32
	max_body_len       uint32

	handlerMap map[uint8]MsgHandler

	sync_master *SyncManager
}

func NewNetPlugin() *NetPlugin {
	return &NetPlugin{
		name:               "net_plugin",
		connection_num:     0,
		max_connection_num: 100,
		max_body_len:       20480,
		sync_master:        &SyncManager{},
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
	dst                  uint64
	org                  uint64
	xmt                  uint64
	rec                  uint64
	offset               uint64
}

func (this *Peer) send_handshake() {
}
func (this *Peer) flush_queues() {
}
func (this *Peer) get_time() uint64 {
	return 0
}
func (this *Peer) Close() {
}

func (this *NetPlugin) handleInit(conn *net.TCPConn) (*Peer, error) {
	return nil, nil
}

func (this *NetPlugin) Startup() error {
	fmt.Printf("%s startup\n", this.name)
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
	fmt.Printf("%s shutdown\n", this.name)
}

func (this *NetPlugin) Name() string {
	return this.name
}

type HandshakeMsg struct {
	generation      int16
	node_id         string
	chain_id        string
	head_id         string
	network_version int16
	head_num        int
	p2p_address     string
	os              string
	agent           string

	last_irreversible_block_num uint32
	last_irreversible_block_id  uint32
}

type GoAwayMsg struct {
	reason  string
	node_id string
}

func HandleHandshake(plugin *NetPlugin, conn *net.TCPConn) (*Peer, int) {
	var cc ChainController
	var msg HandshakeMsg
	var peer Peer

	lib_num := cc.last_irreversible_block_num()
	peer_lib := msg.last_irreversible_block_num

	if msg.generation == 1 {
		if msg.node_id == plugin.node_id {
			return nil, 1
		}

		if msg.chain_id != plugin.chain_id {
			return nil, 2
		}

		if msg.network_version != plugin.network_version {
			if plugin.network_version_match {
				return nil, 3
			}
		}
		if peer.node_id != msg.node_id {
			peer.node_id = msg.node_id
		}
		if !authenticate_peer(&msg) {
			return nil, 4
		}
		on_fork := false
		if peer_lib <= lib_num && peer_lib > 0 {
			peer_lib_id := cc.get_block_id_for_num(peer_lib)
			if msg.last_irreversible_block_id != peer_lib_id {
				on_fork = true
			}
			if on_fork {
				return nil, 4
			}
		}

		if peer.sent_handshake_count == 0 {
			peer.send_handshake()
		}
	}

	plugin.sync_master.recv_handshake(&peer, &msg)
	return nil, 0
}

func HandleGoAway(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	var msg GoAwayMsg
	reason := msg.reason
	if reason == "duplicate" {
		peer.node_id = msg.node_id
	}
	peer.flush_queues()
	peer.Close()
	return 0
}

type TimeMsg struct {
	org uint64
	rec uint64
	xmt uint64
	dst uint64
}

func HandleTime(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	var msg TimeMsg
	msg.dst = peer.get_time()
	if msg.xmt == 0 {
		return 1
	}
	if msg.xmt == peer.xmt {
		return 1
	}
	peer.xmt = msg.xmt
	peer.rec = msg.rec
	peer.dst = msg.dst
	if msg.org == 0 {
		return 1
	}
	peer.offset = ((peer.rec - peer.org) + (msg.xmt - peer.dst)) / 2
	peer.org = 0
	peer.rec = 0
	return 0
}

func HandleNotice(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}

type RequestMsg struct {
	req_blocks struct {
		mode int
	}
	req_trx struct {
		mode int
	}
}

func HandleRequest(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	var msg RequestMsg
	switch msg.req_blocks.mode {
	}
	switch msg.req_trx.mode {
	}

	return 0
}

type SyncRequestMsg struct {
}

func HandleSyncRequest(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	var msg SyncRequestMsg
	if msg.end_block == 0 {
		peer.flush_queues()
	} else {
		peer.enqueue_sync_block()
	}
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

type SignedTransactionMsg struct {
}

func HandleSignedTransaction(conn *net.TCPConn, peer *Peer, header *p2p.Header, body []byte) int {
	var msg SignedTransactionMsg
	peer.cancel_wait()
	return 0
}

type SyncManager struct {
}

func (this *SyncManager) recv_handshake(peer *Peer, msg *HandshakeMsg) {
	var cc ChainController
	//lib_num := cc.last_irreversible_block_num()
	//peer_lib := msg.last_irreversible_block_num

	peer.syncing = false
	head_num := cc.head_block_num()
	head_id := cc.head_block_id()
	if head_id == msg.head_id {
	}
	if head_num < 0 {
	}
	if msg.head_num < 100 {
	}
	if head_num <= msg.head_num {
		return
	}
	if msg.generation > 1 {
		/*var NoticeMsg note
		peer.enqueue(note)*/
	}
	peer.syncing = true
}

func authenticate_peer(msg *HandshakeMsg) bool {
	return true
}
