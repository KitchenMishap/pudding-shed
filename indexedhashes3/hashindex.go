package indexedhashes3

type hashIndex int64

// Count at 15 years:
const bitsFor2milBlocks = int64(21) // 2^21 = 2,097,152 blocks				There were 824,204 blocks
const bitsFor2bilTrans = int64(31)  // 2^31 = 2,147,483,648 transactions	There were 947,337,057 transactions
const bitsFor4bilAddrs = int64(32)  // 2^32 = 4,294,967,296 addresses		There must be fewer addresses than txos
//										(There were 2,652,374,369 txos, including spent)
