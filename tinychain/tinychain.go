package tinychain

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// A tiny read only hard coded blockchain useful for testing
// Implements the chainreadinterface collection of interfaces.

// TheTinyChain is the top level interface to the chain
var TheTinyChain = (chainreadinterface.IBlockChain)(&theTinyChain)

// TheHandles is the top level interface for handling handles
var TheHandles = (chainreadinterface.IHandles)(&theHandles)
