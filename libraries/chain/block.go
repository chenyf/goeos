package chain

type BlockHeader struct {
}

func (this *BlockHeader) block_num() uint32 {
	return 0
}

type SignedBlockHeader struct {
	BlockHeader
}

type ShardLock struct {
	account string
	scope   string
}

type ShardSummary struct {
	read_locks   []ShardLock
	write_lock   []ShardLock
	Transactions []TransactionReceipt
}

type RegionSummary struct {
	region        uint16
	CyclesSummary [][]ShardSummary
}

type SignedBlockSummary struct {
	SignedBlockHeader
	Regions []RegionSummary
}

type SignedBlock struct {
	SignedBlockSummary
	input_transaction []PackedTransaction
}
