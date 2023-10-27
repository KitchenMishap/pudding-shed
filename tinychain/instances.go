package tinychain

// ------------------
// Block 0 (Genesis)
// Transaction 0 (Coinbase)
// TXOs
var txo_b0_t0_txo0 = Txo{
	satoshis: 50,
}

// Transaction
var trans_b0_t0 = Transaction{
	txis: []Txi{},
	txos: []Txo{txo_b0_t0_txo0},
}

// Block
var block_b0 = Block{
	transactions: []Transaction{trans_b0_t0},
}

// --------
// Block 1
// Transaction 0 (Coinbase)
// TXOs
var txo_b1_t0_txo0 = Txo{
	satoshis: 55, // Includes fee of 5
}

// Transaction
var trans_b1_t0 = Transaction{
	txis: []Txi{},
	txos: []Txo{txo_b1_t0_txo0},
}

// Transaction 1
// TXIs
var txi_b1_t1_txi0 = Txi{
	sourceTxo: txo_b0_t0_txo0.TxoHandle,
}

// TXOs
var txo_b1_t1_txo0 = Txo{
	satoshis: 20,
}
var txo_b1_t1_txo1 = Txo{
	satoshis: 25,
}

// Transaction
var trans_b1_t1 = Transaction{
	txis: []Txi{txi_b1_t1_txi0},
	txos: []Txo{txo_b1_t1_txo0, txo_b1_t1_txo1},
}

// Block
var block_b1 = Block{
	transactions: []Transaction{trans_b1_t0, trans_b1_t1},
}

// Block 2
// Transactons
// Transaction 0 (Coinbase)
// TXOs
var txo_b2_t0_txo0 = Txo{
	satoshis: 50,
}

// Transaction
var trans_b2_t0 = Transaction{
	txis: []Txi{},
	txos: []Txo{txo_b2_t0_txo0},
}

// Transaction 1
// TXIs
var txi_b2_t1_txi0 = Txi{
	sourceTxo: txo_b1_t1_txo0.TxoHandle,
}
var txi_b2_t1_txi1 = Txi{
	sourceTxo: txo_b1_t1_txo1.TxoHandle,
}

// TXOs
var txo_b2_t1_txo0 = Txo{
	satoshis: 45,
}

// Transaction
var trans_b2_t1 = Transaction{
	txis: []Txi{txi_b2_t1_txi0, txi_b2_t1_txi1},
	txos: []Txo{txo_b2_t1_txo0},
}

// Block
var block_b2 = Block{
	transactions: []Transaction{trans_b2_t0, trans_b2_t1},
}

// ------------
// Whole Thing
var theTinyChain = Blockchain{
	blocks:       []Block{block_b0, block_b1, block_b2},
	transactions: []Transaction{trans_b0_t0, trans_b1_t0, trans_b1_t1, trans_b2_t0, trans_b2_t1},
}
