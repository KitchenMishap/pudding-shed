package concurrency

type Sequencable interface {
	SequenceNumber() int64
}
