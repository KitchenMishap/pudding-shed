package corereader

import (
	"github.com/KitchenMishap/pudding-shed/jsonblock"
	"sync"
)

// corereader.Pool implements jsonblock.IBlockJsonFetcher
var _ jsonblock.IBlockJsonFetcher = (*Pool)(nil) // Check that implements

func NewPool(poolSize int) *Pool {
	result := Pool{}
	result.InChan = make(chan *Task)
	result.counter = CoreReader{}
	result.wg.Add(poolSize)
	for i := 0; i < poolSize; i++ {
		reader := CoreReader{}
		go result.worker(reader)
	}
	return &result
}

type Pool struct {
	InChan  chan *Task
	counter CoreReader
	wg      sync.WaitGroup
}

func (p *Pool) CountBlocks() (int64, error) {
	return p.counter.CountBlocks()
}
func (p *Pool) FetchBlockJsonBytes(height int64) ([]byte, error) {
	resultChan := make(chan *Task)
	task := NewTask(height, &resultChan)
	p.InChan <- task
	resultTask := <-resultChan
	return resultTask.ResultBytes, resultTask.ResultErr
}

func NewTask(height int64, completionChan *chan *Task) *Task {
	result := Task{}
	result.BlockHeight = height
	result.CompletionChan = completionChan
	return &result
}

type Task struct {
	BlockHeight    int64
	ResultBytes    []byte
	ResultErr      error
	CompletionChan *chan *Task
}

func (pool *Pool) worker(reader CoreReader) {
	for task := range pool.InChan {
		task.ResultBytes, task.ResultErr = reader.FetchBlockJsonBytes(task.BlockHeight)
		*task.CompletionChan <- task
	}
	pool.wg.Done()
}

func (pool *Pool) Flush() {
	pool.wg.Wait()
}
