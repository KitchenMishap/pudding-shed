package jsonblock

const hello string = `Hello "World"`

// Note these strings aren't fixed. For example, "confirmations" can change!
const block0json string = `{"hash":"000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f","confirmations":788250,"height":0,"version":1,"versionHex":"00000001","merkleroot":"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b","time":1231006505,"mediantime":1231006505,"nonce":2083236893,"bits":"1d00ffff","difficulty":1,"chainwork":"0000000000000000000000000000000000000000000000000000000100010001","nTx":1,"nextblockhash":"00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048","strippedsize":285,"size":285,"weight":1140,"tx":[{"txid":"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b","hash":"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b","version":1,"size":204,"vsize":204,"weight":816,"locktime":0,"vin":[{"coinbase":"04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73","sequence":4294967295}],"vout":[{"value":50.00000000,"n":0,"scriptPubKey":{"asm":"04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f OP_CHECKSIG","hex":"4104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac","type":"pubkey"}}],"hex":"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4d04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73ffffffff0100f2052a01000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"}]}`
const block1json string = `{"hash":"00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048","confirmations":788865,"height":1,"version":1,"versionHex":"00000001","merkleroot":"0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098","time":1231469665,"mediantime":1231469665,"nonce":2573394689,"bits":"1d00ffff","difficulty":1,"chainwork":"0000000000000000000000000000000000000000000000000000000200020002","nTx":1,"previousblockhash":"000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f","nextblockhash":"000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd","strippedsize":215,"size":215,"weight":860,"tx":[{"txid":"0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098","hash":"0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098","version":1,"size":134,"vsize":134,"weight":536,"locktime":0,"vin":[{"coinbase":"04ffff001d0104","sequence":4294967295}],"vout":[{"value":50.00000000,"n":0,"scriptPubKey":{"asm":"0496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858ee OP_CHECKSIG","hex":"410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac","type":"pubkey"}}],"hex":"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0704ffff001d0104ffffffff0100f2052a0100000043410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac00000000"}]}`
const block2json string = `{"hash":"000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd","confirmations":789054,"height":2,"version":1,"versionHex":"00000001","merkleroot":"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5","time":1231469744,"mediantime":1231469665,"nonce":1639830024,"bits":"1d00ffff","difficulty":1,"chainwork":"0000000000000000000000000000000000000000000000000000000300030003","nTx":1,"previousblockhash":"00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048","nextblockhash":"0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449","strippedsize":215,"size":215,"weight":860,"tx":[{"txid":"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5","hash":"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5","version":1,"size":134,"vsize":134,"weight":536,"locktime":0,"vin":[{"coinbase":"04ffff001d010b","sequence":4294967295}],"vout":[{"value":50.00000000,"n":0,"scriptPubKey":{"asm":"047211a824f55b505228e4c3d5194c1fcfaa15a456abdf37f9b9d97a4040afc073dee6c89064984f03385237d92167c13e236446b417ab79a0fcae412ae3316b77 OP_CHECKSIG","hex":"41047211a824f55b505228e4c3d5194c1fcfaa15a456abdf37f9b9d97a4040afc073dee6c89064984f03385237d92167c13e236446b417ab79a0fcae412ae3316b77ac","type":"pubkey"}}],"hex":"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0704ffff001d010bffffffff0100f2052a010000004341047211a824f55b505228e4c3d5194c1fcfaa15a456abdf37f9b9d97a4040afc073dee6c89064984f03385237d92167c13e236446b417ab79a0fcae412ae3316b77ac00000000"}]}`
const block3json string = `{"hash":"0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449","confirmations":789123,"height":3,"version":1,"versionHex":"00000001","merkleroot":"999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644","time":1231470173,"mediantime":1231469744,"nonce":1844305925,"bits":"1d00ffff","difficulty":1,"chainwork":"0000000000000000000000000000000000000000000000000000000400040004","nTx":1,"previousblockhash":"000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd","nextblockhash":"000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485","strippedsize":215,"size":215,"weight":860,"tx":[{"txid":"999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644","hash":"999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644","version":1,"size":134,"vsize":134,"weight":536,"locktime":0,"vin":[{"coinbase":"04ffff001d010e","sequence":4294967295}],"vout":[{"value":50.00000000,"n":0,"scriptPubKey":{"asm":"0494b9d3e76c5b1629ecf97fff95d7a4bbdac87cc26099ada28066c6ff1eb9191223cd897194a08d0c2726c5747f1db49e8cf90e75dc3e3550ae9b30086f3cd5aa OP_CHECKSIG","hex":"410494b9d3e76c5b1629ecf97fff95d7a4bbdac87cc26099ada28066c6ff1eb9191223cd897194a08d0c2726c5747f1db49e8cf90e75dc3e3550ae9b30086f3cd5aaac","type":"pubkey"}}],"hex":"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0704ffff001d010effffffff0100f2052a0100000043410494b9d3e76c5b1629ecf97fff95d7a4bbdac87cc26099ada28066c6ff1eb9191223cd897194a08d0c2726c5747f1db49e8cf90e75dc3e3550ae9b30086f3cd5aaac00000000"}]}`
const block4json string = `{"hash":"000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485","confirmations":789552,"height":4,"version":1,"versionHex":"00000001","merkleroot":"df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a","time":1231470988,"mediantime":1231469744,"nonce":2850094635,"bits":"1d00ffff","difficulty":1,"chainwork":"0000000000000000000000000000000000000000000000000000000500050005","nTx":1,"previousblockhash":"0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449","nextblockhash":"000000009b7262315dbf071787ad3656097b892abffd1f95a1a022f896f533fc","strippedsize":215,"size":215,"weight":860,"tx":[{"txid":"df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a","hash":"df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a","version":1,"size":134,"vsize":134,"weight":536,"locktime":0,"vin":[{"coinbase":"04ffff001d011a","sequence":4294967295}],"vout":[{"value":50.00000000,"n":0,"scriptPubKey":{"asm":"04184f32b212815c6e522e66686324030ff7e5bf08efb21f8b00614fb7690e19131dd31304c54f37baa40db231c918106bb9fd43373e37ae31a0befc6ecaefb867 OP_CHECKSIG","hex":"4104184f32b212815c6e522e66686324030ff7e5bf08efb21f8b00614fb7690e19131dd31304c54f37baa40db231c918106bb9fd43373e37ae31a0befc6ecaefb867ac","type":"pubkey"}}],"hex":"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0704ffff001d011affffffff0100f2052a01000000434104184f32b212815c6e522e66686324030ff7e5bf08efb21f8b00614fb7690e19131dd31304c54f37baa40db231c918106bb9fd43373e37ae31a0befc6ecaefb867ac00000000"}]}`

var hardCodedJsonBlocks = []string{block0json, block1json, block2json, block3json, block4json}

func HardCodedJsonBlock(height int64) string {
	return hardCodedJsonBlocks[height]
}

func HardCodedJsonBlockCount() int64 {
	return int64(len(hardCodedJsonBlocks))
}

// implements IBlockJsonFetcher
var _ IBlockJsonFetcher = (*HardCodedBlockFetcher)(nil) // Check that implements
type HardCodedBlockFetcher struct {
}

func (fbf *HardCodedBlockFetcher) CountBlocks() (int64, error) {
	return HardCodedJsonBlockCount(), nil
}
func (fbf *HardCodedBlockFetcher) FetchBlockJsonBytes(height int64) ([]byte, error) {
	return []byte(HardCodedJsonBlock(height)), nil
}
