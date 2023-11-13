package tinychain

// ------------------
// Block 0 (Genesis)
// Transaction 0 (Coinbase)
// TXOs
var transIndex_txo_b0_t0_txo0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 0}}, index: 0}
var handle_txo_b0_t0_txo0 = TxoHandle{TxxHandle{TransIndex: transIndex_txo_b0_t0_txo0, txxHeight: -1, txxHeightSpecified: false}}
var txo_b0_t0_txo0 = Txo{TxoHandle: handle_txo_b0_t0_txo0, satoshis: 50}

// Transaction
var trans_b0_t0 = Transaction{
	TransHandle: TransHandle{HashHeight{0}},
	txis:        []Txi{},
	txos:        []Txo{txo_b0_t0_txo0},
}

// Block
var block_b0 = Block{
	BlockHandle:  BlockHandle{HashHeight{height: 0}},
	transactions: []Transaction{trans_b0_t0},
}

// --------
// Block 1
// Transaction 0 (Coinbase)
// TXOs
var transIndex_txo_b1_t0_txo0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 1}}, index: 0}
var handle_txo_b1_t0_txo0 = TxoHandle{TxxHandle{TransIndex: transIndex_txo_b1_t0_txo0, txxHeight: -1, txxHeightSpecified: false}}
var txo_b1_t0_txo0 = Txo{TxoHandle: handle_txo_b1_t0_txo0, satoshis: 55}

// Transaction
var trans_b1_t0 = Transaction{
	TransHandle: TransHandle{HashHeight{1}},
	txis:        []Txi{},
	txos:        []Txo{txo_b1_t0_txo0},
}

// Transaction 1
// TXIs
var transIndex_txi_b1_t1_txi0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 2}}, index: 0}
var handle_txi_b1_t1_txi0 = TxiHandle{TxxHandle{TransIndex: transIndex_txi_b1_t1_txi0, txxHeight: -1, txxHeightSpecified: false}}
var txi_b1_t1_txi0 = Txi{TxiHandle: handle_txi_b1_t1_txi0, sourceTxo: handle_txo_b0_t0_txo0}

// TXOs
var transIndex_txo_b1_t1_txo0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 2}}, index: 0}
var handle_txo_b1_t1_txo0 = TxoHandle{TxxHandle{TransIndex: transIndex_txo_b1_t1_txo0, txxHeight: -1, txxHeightSpecified: false}}
var txo_b1_t1_txo0 = Txo{TxoHandle: handle_txo_b1_t1_txo0, satoshis: 20}

var transIndex_txo_b1_t1_txo1 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 2}}, index: 1}
var handle_txo_b1_t1_txo1 = TxoHandle{TxxHandle{TransIndex: transIndex_txo_b1_t1_txo1, txxHeight: -1, txxHeightSpecified: false}}
var txo_b1_t1_txo1 = Txo{TxoHandle: handle_txo_b1_t1_txo1, satoshis: 25}

// Transaction
var trans_b1_t1 = Transaction{
	TransHandle: TransHandle{HashHeight{2}},
	txis:        []Txi{txi_b1_t1_txi0},
	txos:        []Txo{txo_b1_t1_txo0, txo_b1_t1_txo1},
}

// Block
var block_b1 = Block{
	BlockHandle:  BlockHandle{HashHeight{height: 1}},
	transactions: []Transaction{trans_b1_t0, trans_b1_t1},
}

// Block 2
// Transactons
// Transaction 0 (Coinbase)
// TXOs
var transIndex_txo_b2_t0_txo0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 3}}, index: 0}
var handle_txo_b2_t0_txo0 = TxoHandle{TxxHandle{TransIndex: transIndex_txo_b2_t0_txo0, txxHeight: -1, txxHeightSpecified: false}}
var txo_b2_t0_txo0 = Txo{TxoHandle: handle_txo_b2_t0_txo0, satoshis: 50}

// Transaction
var trans_b2_t0 = Transaction{
	TransHandle: TransHandle{HashHeight{3}},
	txis:        []Txi{},
	txos:        []Txo{txo_b2_t0_txo0},
}

// Note that transaction 1 takes an input from transaction 2's output

// Transaction 1
// TXIs
var transIndex_txi_b2_t1_txi0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 4}}, index: 0}
var handle_txi_b2_t1_txi0 = TxiHandle{TxxHandle{TransIndex: transIndex_txi_b2_t1_txi0, txxHeight: -1, txxHeightSpecified: false}}
var txi_b2_t1_txi0 = Txi{TxiHandle: handle_txi_b2_t1_txi0, sourceTxo: handle_txo_b2_t1_txo0}

// TXOs
var transIndex_txo_b2_t1_txo0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 4}}, index: 0}
var handle_txo_b2_t1_txo0 = TxoHandle{TxxHandle{TransIndex: transIndex_txo_b2_t1_txo0, txxHeight: -1, txxHeightSpecified: false}}
var txo_b2_t1_txo0 = Txo{TxoHandle: handle_txo_b2_t1_txo0, satoshis: 40}

// Transaction
var trans_b2_t1 = Transaction{
	TransHandle: TransHandle{HashHeight{4}},
	txis:        []Txi{txi_b2_t1_txi0},
	txos:        []Txo{txo_b2_t1_txo0},
}

// Transaction 2
// TXIs
var transIndex_txi_b2_t2_txi0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 5}}, index: 0}
var handle_txi_b2_t2_txi0 = TxiHandle{TxxHandle{TransIndex: transIndex_txi_b2_t2_txi0, txxHeight: -1, txxHeightSpecified: false}}
var txi_b2_t2_txi0 = Txi{TxiHandle: handle_txi_b2_t2_txi0, sourceTxo: handle_txo_b1_t1_txo0}
var transIndex_txi_b2_t2_txi1 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 5}}, index: 1}
var handle_txi_b2_t2_txi1 = TxiHandle{TxxHandle{TransIndex: transIndex_txi_b2_t2_txi1, txxHeight: -1, txxHeightSpecified: false}}
var txi_b2_t2_txi1 = Txi{TxiHandle: handle_txi_b2_t2_txi1, sourceTxo: handle_txo_b1_t1_txo1}

// TXOs
var transIndex_txo_b2_t2_txo0 = TransIndex{TransHandle: TransHandle{HashHeight: HashHeight{height: 5}}, index: 0}
var handle_txo_b2_t2_txo0 = TxoHandle{TxxHandle{TransIndex: transIndex_txo_b2_t2_txo0, txxHeight: -1, txxHeightSpecified: false}}
var txo_b2_t2_txo0 = Txo{TxoHandle: handle_txo_b2_t2_txo0, satoshis: 45}

// Transaction
var trans_b2_t2 = Transaction{
	TransHandle: TransHandle{HashHeight{5}},
	txis:        []Txi{txi_b2_t2_txi0, txi_b2_t2_txi1},
	txos:        []Txo{txo_b2_t2_txo0},
}

// Block
var block_b2 = Block{
	BlockHandle:  BlockHandle{HashHeight{height: 2}},
	transactions: []Transaction{trans_b2_t0, trans_b2_t1, trans_b2_t2}, // Note t1,t2 reversed
}

// ------------
// Whole Thing
var theTinyChain = Blockchain{
	blocks:       []Block{block_b0, block_b1, block_b2},
	transactions: []Transaction{trans_b0_t0, trans_b1_t0, trans_b1_t1, trans_b2_t0, trans_b2_t1, trans_b2_t2},
}
