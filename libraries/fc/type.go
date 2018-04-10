package fc

type sha256 uint64

type BlockIdType sha256

const SkipNothing = 0
const SkipProducerSignature = 1 << 0
const SkipTransactionSignatures = 1 << 1
const skip_transaction_dupe_check = 1 << 2
const skip_fork_db = 1 << 3
const skip_block_size_check = 1 << 4
const skip_tapos_check = 1 << 5
const skip_authority_check = 1 << 6
const skip_merkle_check = 1 << 7
const skip_assert_evaluation = 1 << 8
const skip_undo_history_check = 1 << 9
const skip_producer_schedule_check = 1 << 10
const skip_validate = 1 << 11
const skip_scope_check = 1 << 12
const skip_output_check = 1 << 13
const pushed_transaction = 1 << 14
const created_block = 1 << 15
const received_block = 1 << 16
const genesis_setup = 1 << 17
const skip_missed_block_penalty = 1 << 18
