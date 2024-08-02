package genericentities

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

type IEntityHandle interface {
	// Non-nil
	BlockChain() chainreadinterface.IBlockChain
	// None or exactly one of these four will be non-nil
	MaybeBlock() chainreadinterface.IBlockHandle
	MaybeTransaction() chainreadinterface.ITransHandle
	MaybeTxi() chainreadinterface.ITxiHandle
	MaybeTxo() chainreadinterface.ITxoHandle
}

type IFieldProvider interface {
	IntFieldNamesAvailable() []string
	StringFieldNamesAvailable() []string
	GetFieldTypeHint(fieldName string) string
	GetIntField(fieldName string) int64
	GetStringField(fieldName string) string
}

type IEntity interface {
	IEntityHandle
	IFieldProvider
	EntityTypeName() string
	IdentityRepresentations() []string
	PrevEntities() map[string]IEntityHandle
	NextEntities() map[string]IEntityHandle
	ParentEntities() map[string][]IEntityHandle
	ChildEntities() map[string][]IEntityHandle
}
