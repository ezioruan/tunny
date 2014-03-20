package skank

import (
	"reflect"
	"errors"
)

type Worker struct {
	readyChan  chan int
	jobChan    chan interface{}
	outputChan chan interface{}
	job        *func(interface{}) (interface{})
}

type WorkPool struct {
	workers []*Worker
	selects []reflect.SelectCase
	running bool
}

func CreatePool ( numWorkers int, job func(interface{}) (interface{}) ) *WorkPool {
	var pool WorkPool

	pool.running = false

	pool.workers = make ([]*Worker, numWorkers)
	for i, _ := range pool.workers {
		pool.workers[i] = new (Worker)
		pool.workers[i].readyChan  = make (chan int)
		pool.workers[i].jobChan    = make (chan interface{})
		pool.workers[i].outputChan = make (chan interface{})
		pool.workers[i].job = &job
	}

	pool.selects = make( []reflect.SelectCase, len(pool.workers) )

	for i, worker := range pool.workers {
		pool.selects[i] = reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: reflect.ValueOf(worker.readyChan) }
	}

	return &pool
}

func (pool *WorkPool) SendWork ( jobData interface{} ) (interface{}, error) {
	if pool.running {

		if chosen, _, ok := reflect.Select(pool.selects); ok && chosen >= 0 {
			pool.workers[chosen].jobChan <- jobData
			return <- pool.workers[chosen].outputChan, nil
		}

		return nil, errors.New("No workers or some stupid shit")

	} else {
		return nil, errors.New("Pool is not running! Cannot take work")
	}
}

func (pool *WorkPool) Begin () *WorkPool {
	for i, _ := range pool.workers {
		go pool.workers[i].Work()
	}
	pool.running = true
	return pool
}

func (worker *Worker) Work () {
	worker.readyChan <- 1
	for data := range worker.jobChan {
		worker.outputChan <- (*worker.job)( data )
		worker.readyChan <- 1
	}
}
