package corereader

import "github.com/KitchenMishap/pudding-shed/jsonblock"

// corereader.Pool implements jsonblock.IBlockJsonFetcher
var _ jsonblock.IBlockJsonFetcher = (*Pool)(nil) // Check that implements

func NewPool(poolSize int) *Pool {
	result := Pool{}
	result.inChan = make(chan *Task)
	result.counter = CoreReader{}
	for i := 0; i < poolSize; i++ {
		reader := CoreReader{}
		go result.worker(reader)
	}
	return &result
}

type Pool struct {
	inChan  chan *Task
	counter CoreReader
}

func (p *Pool) CountBlocks() (int64, error) {
	return p.counter.CountBlocks()
}
func (p *Pool) FetchBlockJsonBytes(height int64) ([]byte, error) {
	resultChan := make(chan *Task)
	task := NewTask(height, &resultChan)
	p.inChan <- task
	resultTask := <-resultChan
	return resultTask.resultBytes, resultTask.resultErr
}

func NewTask(height int64, completionChan *chan *Task) *Task {
	result := Task{}
	result.blockHeight = height
	result.completionChan = completionChan
	return &result
}

type Task struct {
	blockHeight    int64
	resultBytes    []byte
	resultErr      error
	completionChan *chan *Task
}

func (pool *Pool) worker(reader CoreReader) {
	for task := range pool.inChan {
		task.resultBytes, task.resultErr = reader.FetchBlockJsonBytes(task.blockHeight)
		*task.completionChan <- task
	}
}
