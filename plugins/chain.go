package plugins

import (
	"fmt"
	"github.com/chenyf/goeos/libraries/appbase"
	"github.com/chenyf/goeos/libraries/chain"
	"github.com/chenyf/goeos/libraries/fc"
	"time"
)

type GenesisStateType struct {
}

func LoadGenesis(file string) *GenesisStateType {
	return nil
}

type ChainConfig struct {
	block_log_dir      string
	shared_memory_dir  string
	shared_memory_size uint64
	readonly           bool
	genesis            *GenesisStateType
	limits             struct {
		max_push_block_us       uint32
		max_push_transaction_us uint32
	}
}

func (this *ChainConfig) Reset() {
}

type ChainController struct {
}

func (this *ChainController) Reset() {
}
func (this *ChainController) AddCheckpoints() {
}
func (this *ChainController) Emplace(config *ChainConfig) {
}
func (this *ChainController) head_block_num() int {
	return 0
}
func (this *ChainController) head_block_id() string {
	return ""
}

func (this *ChainController) is_known_block(blk_id uint32) bool {
	return false
}

func (this *ChainController) last_irreversible_block_num() uint32 {
	return 0
}
func (this *ChainController) get_block_id_for_num(uint32) uint32 {
	return 0
}

type ChainPlugin struct {
	name                            string
	addr                            string
	genesis_file                    string
	block_log_dir                   string
	skip_flags                      uint32
	readonly                        bool
	genesis_timestamp               int64
	loaded_checkpoints              map[uint32]fc.BlockIdType
	max_reversible_block_time_ms    uint32
	max_pending_transaction_time_ms uint32
	chain_config                    *ChainConfig
	chain                           *ChainController
}

func NewChainPlugin() *ChainPlugin {
	return &ChainPlugin{
		name:               "chain_plugin",
		skip_flags:         fc.SkipNothing,
		loaded_checkpoints: map[uint32]fc.BlockIdType{},
		chain_config:       &ChainConfig{},
		chain:              &ChainController{},
	}
}

func (this *ChainPlugin) Init(options map[string]interface{}) {
	if genesis, ok := options["genesis-file"]; ok {
		this.genesis_file = genesis.(string)
	}
	if tstr, ok := options["genesis-timestamp"]; ok {
		tt := tstr.(string)
		if tt == "now" {
			now := time.Now().Unix()
			this.genesis_timestamp = now
		} else {
			tm, _ := time.Parse("01/02/2006", tt)
			this.genesis_timestamp = tm.Unix()
		}
	}
	if bld, ok := options["block-log-dir"]; ok {
		this.block_log_dir = bld.(string)
	}
	replay, _ := options["replay-blockchain"]
	if replay.(bool) {
		fmt.Printf("Replay requested: wiping database\n")
		fc.RemoveAll(appbase.Instance().DataDir() + "/" + chain.DefaultSharedMemoryDir)
	}

	resync, _ := options["resync-blockchain"]
	if resync.(bool) {
		fmt.Printf("Resync requested: wiping database and blocks\n")
		fc.RemoveAll(appbase.Instance().DataDir() + "/" + chain.DefaultSharedMemoryDir)
		fc.RemoveAll(this.block_log_dir)
	}

	skipts, _ := options["skip-transaction-signatures"]
	if skipts.(bool) {
		fmt.Printf("Setting skip_transaction_signatures\n")
		this.skip_flags |= fc.SkipTransactionSignatures
	}

	if _, ok := options["checkpoint"]; ok {
	}
	this.max_reversible_block_time_ms = options["max-reversible-block-time"].(uint32)
	this.max_pending_transaction_time_ms = options["max_pending_transaction_time_ms"].(uint32)

}

func (this *ChainPlugin) Startup() error {
	fmt.Printf("%s startup\n", this.name)

	this.chain_config.block_log_dir = this.block_log_dir
	this.chain_config.shared_memory_dir = appbase.Instance().DataDir() + "/" + chain.DefaultSharedMemoryDir
	this.chain_config.readonly = this.readonly
	this.chain_config.genesis = LoadGenesis(this.genesis_file)

	if this.max_reversible_block_time_ms > 0 {
		this.chain_config.limits.max_push_block_us = 0
	}
	if this.max_pending_transaction_time_ms > 0 {
		this.chain_config.limits.max_push_transaction_us = 0
	}
	/*if this.wasm_runtime != nil {
		this.chain_config.wasm_runtime = this.wasm_runtime
	}*/

	this.chain.Emplace(this.chain_config)
	if !this.readonly {
		this.chain.AddCheckpoints()
	}
	this.chain_config.Reset()
	return nil
}

func (this *ChainPlugin) Shutdown() {
	fmt.Printf("%s shutdown\n", this.name)
	this.chain.Reset()
}

func (this *ChainPlugin) Name() string {
	return this.name
}

func (this *ChainPlugin) accept_block(sb *chain.SignedBlock, flag bool) bool {
	return false
}

func (this *ChainPlugin) accept_transaction(trx *PackedTransactionMsg) bool {
	return false
}
func (this *ChainPlugin) block_is_on_preferred_chain() bool {
	return false
}

func (this *ChainPlugin) is_skipping_transaction_signatures() bool {
	flag := this.skip_flags & fc.SkipTransactionSignatures
	if flag == 0 {
		return false
	} else {
		return true
	}
}
