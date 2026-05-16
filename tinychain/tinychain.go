package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// A tiny read only hard coded blockchain useful for testing
// Implements the chainreadinterface collection of interfaces.

// Access to the top level interface to the chain
var TheTinyChain = (chainreadinterface.IBlockChain)(&theTinyChain)
