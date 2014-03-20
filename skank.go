package skank

import (
	"reflect"
	"errors"
)

/* INTERFACES */

type SkankWorker interface {
	Work()
	ReadyChan()  chan int
	JobChan()    chan interface{}
	OutputChan() chan interface{}
}

/* Worker */

type Worker struct {
	readyChan  chan int
	jobChan    chan interface{}
	outputChan chan interface{}
	job        *func(interface{}) (interface{})
}

func (worker *Worker) Work () {
	worker.readyChan <- 1
	for data := range worker.jobChan {
		worker.outputChan <- (*worker.job)( data )
		worker.readyChan <- 1
	}
}

func (worker *Worker) ReadyChan() chan int {
	return worker.readyChan
}

func (worker *Worker) JobChan() chan interface{} {
	return worker.jobChan
}

func (worker *Worker) OutputChan() chan interface{} {
	return worker.outputChan
}

/* WorkPool */

type WorkPool struct {
	workers []*SkankWorker
	selects []reflect.SelectCase
	running bool
}

func CreatePool ( numWorkers int, job func(interface{}) (interface{}) ) *WorkPool {
	pool := WorkPool { running: false }

	pool.workers = make ([]*SkankWorker, numWorkers)
	for i, _ := range pool.workers {
		newWorker := Worker { make (chan int), make (chan interface{}), make (chan interface{}), &job }
		skankWorker := SkankWorker(&newWorker)
		pool.workers[i] = &skankWorker
	}

	pool.selects = make( []reflect.SelectCase, len(pool.workers) )

	for i, worker := range pool.workers {
		pool.selects[i] = reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: reflect.ValueOf((*worker).ReadyChan()) }
	}

	return &pool
}

func CreateCustomPool ( customWorkers []*SkankWorker ) *WorkPool {
	pool := WorkPool { running: false, workers: customWorkers }

	pool.selects = make( []reflect.SelectCase, len(pool.workers) )
	for i, worker := range pool.workers {
		pool.selects[i] = reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: reflect.ValueOf((*worker).ReadyChan()) }
	}

	return &pool
}

func (pool *WorkPool) SendWork ( jobData interface{} ) (interface{}, error) {
	if pool.running {

		if chosen, _, ok := reflect.Select(pool.selects); ok && chosen >= 0 {
			(*pool.workers[chosen]).JobChan() <- jobData
			return <- (*pool.workers[chosen]).OutputChan(), nil
		}

		return nil, errors.New("No workers or some stupid shit")

	} else {
		return nil, errors.New("Pool is not running! Call Begin() before sending work")
	}
}

func (pool *WorkPool) Begin () *WorkPool {
	for i, _ := range pool.workers {
		go (*pool.workers[i]).Work()
	}
	pool.running = true
	return pool
}

func (pool *WorkPool) ForEachWorker ( do func( *SkankWorker ) ) {
	for i, _ := range pool.workers {
		do(pool.workers[i])
	}
}

