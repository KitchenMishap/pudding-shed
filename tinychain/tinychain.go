package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// A tiny read only hard coded blockchain useful for testing
// Implements the chainreadinterface collection of interfaces.
// Does not involve hashes (for blocks nor transactions)

// Access to the top level interface to the chain
var TheTinyChain = (chainreadinterface.IBlockChain)(&theTinyChain)
var TheHandles = (chainreadinterface.IHandles)(&theHandles)
