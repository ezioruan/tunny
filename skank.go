/*
Package skank implements a simple pool for maintaining independant worker threads.
Here's a simple example of skank in action, creating a four threaded worker pool:

pool := skank.CreatePool(4, func( object interface{} ) ( interface{} ) {
	if w, ok := object.(int); ok {
		return w * 2
	}
	return "Not an int!"
}).Open()

defer pool.Close()

// pool.SendWork is thread safe, so it can be called from another pool of go routines.
// This call blocks until a worker is ready and has completed the job
out, err := pool.SendWork(50)
*/
package skank

import (
	"reflect"
	"errors"
	"time"
	"sync"
)

type SkankWorker interface {
	Job(interface{}) (interface{})
	Ready() bool
}

type workerWrapper struct {
	readyChan  chan int
	jobChan    chan interface{}
	outputChan chan interface{}
	worker     SkankWorker
}

func (wrapper *workerWrapper) Loop () {
	for !wrapper.worker.Ready() {
		time.Sleep(50 * time.Millisecond)
	}
	wrapper.readyChan <- 1
	for data := range wrapper.jobChan {
		wrapper.outputChan <- wrapper.worker.Job( data )
		for !wrapper.worker.Ready() {
			time.Sleep(50 * time.Millisecond)
		}
		wrapper.readyChan <- 1
	}
	close(wrapper.readyChan)
	close(wrapper.outputChan)
}

func (wrapper *workerWrapper) Close () {
	close(wrapper.jobChan)
}

type skankDefaultWorker struct {
	job *func(interface{}) (interface{})
}

func (worker *skankDefaultWorker) Job(data interface{}) interface{} {
	return (*worker.job)(data)
}

func (worker *skankDefaultWorker) Ready() bool {
	return true
}

/*
WorkPool allows you to contain and send work to your worker pool.
You must first indicate that the pool should run by calling Open(), then send work to the workers
through SendWork.
*/
type WorkPool struct {
	workers []*workerWrapper
	selects []reflect.SelectCase
	mutex   sync.RWMutex
	running bool
}

func (pool *WorkPool) SendWork (jobData interface{}) (interface{}, error) {
	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	if pool.running {

		if chosen, _, ok := reflect.Select(pool.selects); ok && chosen >= 0 {
			(*pool.workers[chosen]).jobChan <- jobData
			return <- (*pool.workers[chosen]).outputChan, nil
		}

		return nil, errors.New("Failed to find or wait for a worker")

	} else {
		return nil, errors.New("Pool is not running! Call Open() before sending work")
	}
}

func (pool *WorkPool) Open () (*WorkPool, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !pool.running {

		pool.selects = make( []reflect.SelectCase, len(pool.workers) )

		for i, worker := range pool.workers {
			(*worker).readyChan  = make (chan int)
			(*worker).jobChan    = make (chan interface{})
			(*worker).outputChan = make (chan interface{})

			pool.selects[i] = reflect.SelectCase {
				Dir: reflect.SelectRecv,
				Chan: reflect.ValueOf((*worker).readyChan),
			}

			go (*worker).Loop()
		}

		pool.running = true
		return pool, nil

	} else {
		return nil, errors.New("Pool is already running!")
	}
}

func (pool *WorkPool) Close() error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.running {

		for _, worker := range pool.workers {
			(*worker).Close()
		}
		pool.running = false
		return nil

	} else {
		return errors.New("Cannot close when the pool is not running!")
	}
}

/*
CreatePool is a helper function that creates a pool of workers.
Args: numWorkers int, job func(interface{}) (interface{})
Summary: number of threads, the closure to run for each job
*/
func CreatePool (numWorkers int, job func(interface{}) interface{}) *WorkPool {
	pool := WorkPool { running: false }

	pool.workers = make ([]*workerWrapper, numWorkers)
	for i, _ := range pool.workers {
		newWorker := workerWrapper {
			worker: &(skankDefaultWorker { &job }),
		}
		pool.workers[i] = &newWorker
	}

	return &pool
}

/*
CreateCustomPool is a helper function that creates a pool for an array of custom workers.
Args: customWorkers []SkankWorker
Summary: An array of workers to use in the pool, each worker gets its own thread
*/
func CreateCustomPool (customWorkers []SkankWorker) *WorkPool {
	pool := WorkPool { running: false }

	pool.workers = make ([]*workerWrapper, len(customWorkers))
	for i, _ := range pool.workers {
		newWorker := workerWrapper {
			worker: customWorkers[i],
		}
		pool.workers[i] = &newWorker
	}

	return &pool
}
