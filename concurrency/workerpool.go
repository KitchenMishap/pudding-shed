package concurrency

func NewWorkerPool(concurrency int) *WorkerPool {
	result := WorkerPool{}
	result.InChan = make(chan Task)
	for i := 0; i < concurrency; i++ {
		go result.worker()
	}
	return &result
}

type Task interface {
	Process() error // For the worker to call
	SetError(error) // For the worker to call with return value of Process()
	Done()
	GetError() error // For the client to call later
}

type WorkerPool struct {
	InChan chan Task
}

func (wp *WorkerPool) worker() {
	for task := range wp.InChan {
		task.SetError(task.Process())
		task.Done()
	}
}
