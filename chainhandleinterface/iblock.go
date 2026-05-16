package chainhandleinterface

type IBlock interface {
	TransactionCount() (int64, error)
}
