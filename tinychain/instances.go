package tinychain

var theHandles handles

// ------------------
// Block 0 (Genesis)
// Transaction 0 (Coinbase)
// TXOs
var txo_b0_t0_txo0 = txo{
	satoshis: 50,
}

// Transaction
var trans_b0_t0 = transaction{
	height: 0,
	txis:   []txi{},
	txos:   []txo{txo_b0_t0_txo0},
}

// Block
var block_b0 = block{
	height:       0,
	transactions: []transaction{trans_b0_t0},
}

// --------
// Block 1
// Transaction 0 (Coinbase)
// TXOs
var txo_b1_t0_txo0 = txo{
	satoshis: 55, // Includes fee of 5
}

// Transaction
var trans_b1_t0 = transaction{
	height: 1,
	txis:   []txi{},
	txos:   []txo{txo_b1_t0_txo0},
}

// Transaction 1
// TXIs
var txi_b1_t1_txi0 = txi{
	sourceTransactionHeight: trans_b0_t0.height,
	sourceIndex:             0,
}

// TXOs
var txo_b1_t1_txo0 = txo{
	satoshis: 20,
}
var txo_b1_t1_txo1 = txo{
	satoshis: 25,
}

// Transaction
var trans_b1_t1 = transaction{
	height: 2,
	txis:   []txi{txi_b1_t1_txi0},
	txos:   []txo{txo_b1_t1_txo0, txo_b1_t1_txo1},
}

// Block
var block_b1 = block{
	height:       1,
	transactions: []transaction{trans_b1_t0, trans_b1_t1},
}

// Block 2
// Transactons
// Transaction 0 (Coinbase)
// TXOs
var txo_b2_t0_txo0 = txo{
	satoshis: 50,
}

// Transaction
var trans_b2_t0 = transaction{
	height: 3,
	txis:   []txi{},
	txos:   []txo{txo_b2_t0_txo0},
}

// Transaction 1
// TXIs
var txi_b2_t1_txi0 = txi{
	sourceTransactionHeight: trans_b1_t1.height,
	sourceIndex:             0,
}
var txi_b2_t1_txi1 = txi{
	sourceTransactionHeight: trans_b1_t1.height,
	sourceIndex:             1,
}

// TXOs
var txo_b2_t1_txo0 = txo{
	satoshis: 45,
}

// Transaction
var trans_b2_t1 = transaction{
	height: 4,
	txis:   []txi{txi_b2_t1_txi0, txi_b2_t1_txi1},
	txos:   []txo{txo_b2_t1_txo0},
}

// Block
var block_b2 = block{
	height:       2,
	transactions: []transaction{trans_b2_t0, trans_b2_t1},
}

// ------------
// Whole Thing
var theTinyChain = blockchain{
	blocks:       []block{block_b0, block_b1, block_b2},
	transactions: []transaction{trans_b0_t0, trans_b1_t0, trans_b1_t1, trans_b2_t0, trans_b2_t1},
}
