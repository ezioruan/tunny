/*
Package skank implements a simple pool for maintaining independant worker threads.
Here's a simple example of the skank in action, creating a four threaded worker pool:

pool := skank.CreatePool(4, func( object interface{} ) ( interface{} ) {
	if w, ok := object.(int); ok {
		return w * 2
	}
	return "Not an int!"
}).Begin()

go func() {
	out, err := pool.SendWork(50)
}()
*/
package skank

import (
	"reflect"
	"errors"
)

type SkankWorker interface {
	Job(interface{}) (interface{})
}

type workerWrapper struct {
	readyChan  chan int
	jobChan    chan interface{}
	outputChan chan interface{}
	worker     SkankWorker
}

func (wrapper *workerWrapper) Work () {
	wrapper.readyChan <- 1
	for data := range wrapper.jobChan {
		wrapper.outputChan <- wrapper.worker.Job( data )
		wrapper.readyChan <- 1
	}
}

type skankDefaultWorker struct {
	job *func(interface{}) (interface{})
}

func (worker *skankDefaultWorker) Job(data interface{}) interface{} {
	return (*worker.job)(data)
}

/*
WorkPool allows you to contain and send work to your worker pool.
You must first indicate that the pool should run by calling Begin(), then send work to the workers
through SendWork.
*/
type WorkPool struct {
	workers []*workerWrapper
	selects []reflect.SelectCase
	running bool
}

func (pool *WorkPool) SendWork ( jobData interface{} ) (interface{}, error) {
	if pool.running {

		if chosen, _, ok := reflect.Select(pool.selects); ok && chosen >= 0 {
			(*pool.workers[chosen]).jobChan <- jobData
			return <- (*pool.workers[chosen]).outputChan, nil
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

/*
CreatePool is a helper function that creates a pool of workers.
Args: numWorkers int, job func(interface{}) (interface{})
Summary: number of threads, the closure to run for each job
*/
func CreatePool ( numWorkers int, job func(interface{}) (interface{}) ) *WorkPool {
	pool := WorkPool { running: false }

	pool.workers = make ([]*workerWrapper, numWorkers)
	for i, _ := range pool.workers {
		newWorker := workerWrapper {
			make (chan int),
			make (chan interface{}),
			make (chan interface{}),
			&(skankDefaultWorker { &job }),
		}
		pool.workers[i] = &newWorker
	}

	pool.selects = make( []reflect.SelectCase, len(pool.workers) )

	for i, worker := range pool.workers {
		pool.selects[i] = reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: reflect.ValueOf((*worker).readyChan) }
	}

	return &pool
}

/*func CreateCustomPool ( numWorkers int, customWorker SkankWorker ) *WorkPool {
	pool := WorkPool { running: false }

	pool.workers = make ([]*workerWrapper, numWorkers)
	for i, _ := range pool.workers {
		newWorker := workerWrapper {
			make (chan int),
			make (chan interface{}),
			make (chan interface{}),
			customWorker,
		}
		pool.workers[i] = &newWorker
	}

	pool.selects = make( []reflect.SelectCase, len(pool.workers) )

	for i, worker := range pool.workers {
		pool.selects[i] = reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: reflect.ValueOf((*worker).readyChan) }
	}

	return &pool
}*/

/*
CreateCustomPool is a helper function that creates a pool for an array of custom workers.
Args: customWorkers []SkankWorker
Summary: An array of workers to use in the pool, each worker gets its own thread
*/
func CreateCustomPool ( customWorkers []SkankWorker ) *WorkPool {
	pool := WorkPool { running: false }

	pool.workers = make ([]*workerWrapper, len(customWorkers))
	for i, _ := range pool.workers {
		newWorker := workerWrapper {
			make (chan int),
			make (chan interface{}),
			make (chan interface{}),
			customWorkers[i],
		}
		pool.workers[i] = &newWorker
	}

	pool.selects = make( []reflect.SelectCase, len(pool.workers) )

	for i, worker := range pool.workers {
		pool.selects[i] = reflect.SelectCase{ Dir: reflect.SelectRecv, Chan: reflect.ValueOf((*worker).readyChan) }
	}

	return &pool
}
