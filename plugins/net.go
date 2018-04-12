package plugins

import (
	"fmt"
	"github.com/chenyf/goeos/libraries/chain"
	"github.com/chenyf/goeos/p2p"
	"io"
	"net"
)

type MsgHandler func(*NetPlugin, *Peer, *p2p.Header, []byte) int

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

	chain_plugin   *ChainPlugin
	sync_master    *SyncManager
	big_msg_master *BigMsgManager
}

func NewNetPlugin() *NetPlugin {
	return &NetPlugin{
		name:               "net_plugin",
		connection_num:     0,
		max_connection_num: 100,
		max_body_len:       20480,
		chain_plugin:       nil,
		sync_master:        &SyncManager{},
		big_msg_master:     &BigMsgManager{},
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
		handler(this, peer, &header, dataBuf)
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

func (this *Peer) cancel_wait() {
}

func (this *Peer) enqueue_sync_block() bool {
	return false
}

func (this *Peer) Close() {
}

func (this *Peer) EnqueueMsg(msg interface{}) {
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

func (this *NetPlugin) is_valid(msg *HandshakeMsg) bool {
	if msg.last_irreversible_block_num > (uint32)(msg.head_num) {
		return false
	}

	if msg.p2p_address == "" {
		return false
	}
	if msg.os == "" {
		return false
	}
	return true
}

func HandleHandshake(plugin *NetPlugin, conn *net.TCPConn) (*Peer, int) {
	var cc ChainController
	var msg HandshakeMsg
	var peer Peer

	if !plugin.is_valid(&msg) {
		peer.EnqueueMsg(&GoAwayMsg{reason: "fatal other"})
		return nil, 1
	}

	lib_num := cc.last_irreversible_block_num()
	peer_lib := msg.last_irreversible_block_num

	if msg.generation == 1 {
		if msg.node_id == plugin.node_id {
			peer.EnqueueMsg(&GoAwayMsg{reason: "self"})
			return nil, 1
		}

		if msg.chain_id != plugin.chain_id {
			peer.EnqueueMsg(&GoAwayMsg{reason: "wrong chain"})
			return nil, 2
		}

		if msg.network_version != plugin.network_version {
			if plugin.network_version_match {
				peer.EnqueueMsg(&GoAwayMsg{reason: "wrong version"})
				return nil, 3
			}
		}
		if peer.node_id != msg.node_id {
			peer.node_id = msg.node_id
		}
		if !authenticate_peer(&msg) {
			peer.EnqueueMsg(&GoAwayMsg{reason: "authentication"})
			return nil, 4
		}
		on_fork := false
		if peer_lib <= lib_num && peer_lib > 0 {
			peer_lib_id := cc.get_block_id_for_num(peer_lib)
			if msg.last_irreversible_block_id != peer_lib_id {
				on_fork = true
			}
			if on_fork {
				peer.EnqueueMsg(&GoAwayMsg{reason: "forked"})
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

func HandleGoAway(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
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

func HandleTime(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
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

func HandleNotice(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
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

func HandleRequest(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
	var msg RequestMsg
	switch msg.req_blocks.mode {
	}
	switch msg.req_trx.mode {
	}

	return 0
}

type SyncRequestMsg struct {
	start_block uint32
	end_block   uint32
}

func HandleSyncRequest(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
	var msg SyncRequestMsg
	if msg.end_block == 0 {
		peer.flush_queues()
	} else {
		peer.enqueue_sync_block()
	}
	return 0
}
func HandleSignedBlockSummary(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
	return 0
}

type SignedBlockMsg struct {
}

func HandleSignedBlock(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
	var msg chain.SignedBlock
	var cc ChainController
	//blk_id := msg.id()
	//blk_num := msg.block_num()
	var blk_id uint32 = 0
	var blk_num uint32 = 0
	peer.cancel_wait()

	if cc.is_known_block(blk_id) {
		net_plugin.sync_master.recv_block(peer, blk_id, blk_num, true)
		return 0
	}

	reason := "fatal other"
	ok := net_plugin.chain_plugin.accept_block(&msg, net_plugin.sync_master.is_active(peer))
	if ok {
		reason = "no reason"
	}
	if reason == "no reason" {
		for _, region := range msg.Regions {
			for _, cycle_sum := range region.CyclesSummary {
				for _, shared := range cycle_sum {
					for _, recpt := range shared.Transactions {
						switch recpt.Status {
						}
					}
				}
			}
		}
	}
	net_plugin.sync_master.recv_block(peer, blk_id, blk_num, reason == "no reason")
	return 0
}

type PackedTransactionMsg struct {
}

func HandlePackedTransaction(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
	var msg PackedTransactionMsg
	if net_plugin.sync_master.is_active(peer) {
		return 1
	}
	peer.cancel_wait()
	net_plugin.big_msg_master.recv_transaction(peer)
	ok := net_plugin.chain_plugin.accept_transaction(&msg)
	if !ok {
		net_plugin.big_msg_master.bcast_rejected_transaction(&msg)
	}
	return 0
}

type SignedTransactionMsg struct {
}

func HandleSignedTransaction(net_plugin *NetPlugin, peer *Peer, header *p2p.Header, body []byte) int {
	//var msg SignedTransactionMsg
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

func (this *SyncManager) is_active(peer *Peer) bool {
	return false
}

func (this *SyncManager) recv_block(peer *Peer, id uint32, num uint32, flag bool) bool {
	return false
}

type BigMsgManager struct {
}

func (this *BigMsgManager) recv_transaction(peer *Peer) {
}

func (this *BigMsgManager) bcast_rejected_transaction(msg *PackedTransactionMsg) {
}

func authenticate_peer(msg *HandshakeMsg) bool {
	return true
}
