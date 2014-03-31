package skank

import (
	"testing"
	"time"
	"runtime"
)

func TestTimeout (t *testing.T) {
	outChan  := make(chan int, 3)

	pool, errPool := CreatePool(1, func(object interface{}) interface{} {
		time.Sleep(500 * time.Millisecond)
		return nil
	}).Open()

	if errPool != nil {
		t.Errorf("Error starting pool: ", errPool)
		return
	}

	defer pool.Close()

	before := time.Now()

	go func() {
		if _, err := pool.SendWorkTimed(200, nil); err == nil {
			t.Errorf("No timeout triggered thread one")
		} else {
			taken := ( time.Since(before) / time.Millisecond )
			if taken > 210 {
				t.Errorf("Time taken at thread one: ", taken, ", with error: ", err)
			}
		}
		outChan <- 1

		go func() {
			if _, err := pool.SendWork(nil); err == nil {
			} else {
				t.Errorf("Error at thread three: ", err)
			}
			outChan <- 1
		}()
	}()

	go func() {
		if _, err := pool.SendWorkTimed(200, nil); err == nil {
			t.Errorf("No timeout triggered thread two")
		} else {
			taken := ( time.Since(before) / time.Millisecond )
			if taken > 210 {
				t.Errorf("Time taken at thread two: ", taken, ", with error: ", err)
			}
		}
		outChan <- 1
	}()

	for i := 0; i < 3; i++ {
		<-outChan
	}
}

func validateReturnInt (t *testing.T, expecting int, object interface{}) {
	if w, ok := object.(int); ok {
		if w != expecting {
			t.Errorf("Wrong, expected %v, got %v", expecting, w)
		}
	} else {
		t.Errorf("Wrong, expected int")
	}
}

func TestBasic (t *testing.T) {
	sizePool, repeats, sleepFor, margin := 16, 2, 250, 100
	outChan  := make(chan int, sizePool)

	runtime.GOMAXPROCS(runtime.NumCPU())

	pool, errPool := CreatePool(sizePool, func(object interface{}) interface{} {
		time.Sleep(time.Duration(sleepFor) * time.Millisecond)
		if w, ok := object.(int); ok {
			return w * 2
		}
		return "Not an int!"
	}).Open()

	if errPool != nil {
		t.Errorf("Error starting pool: ", errPool)
		return
	}

	defer pool.Close()

	for i := 0; i < sizePool * repeats; i++ {
		go func() {
			if out, err := pool.SendWork(50); err == nil {
				validateReturnInt (t, 100, out)
			} else {
				t.Errorf("Error returned: ", err)
			}
			outChan <- 1
		}()
	}

	before := time.Now()

	for i := 0; i < sizePool * repeats; i++ {
		<-outChan
	}

	taken    := float64( time.Since(before) ) / float64(time.Millisecond)
	expected := float64( sleepFor + margin ) * float64(repeats)

	if taken > expected {
		t.Errorf("Wrong, should have taken less than %v seconds, actually took %v", expected, taken)
	}
}

func TestExampleCase (t *testing.T) {
	outChan  := make(chan int, 10)
	runtime.GOMAXPROCS(runtime.NumCPU())

	pool, errPool := CreatePool(4, func(object interface{}) interface{} {
		if str, ok := object.(string); ok {
			return "job done: " + str
		}
		return nil
	}).Open()

	if errPool != nil {
		t.Errorf("Error starting pool: ", errPool)
		return
	}

	defer pool.Close()

	for i := 0; i < 10; i++ {
		go func() {
			if value, err := pool.SendWork("hello world"); err == nil {
				if _, ok := value.(string); ok {
				} else {
					t.Errorf("Not a string!")
				}
			} else {
				t.Errorf("Error returned: ", err)
			}
			outChan <- 1
		}()
	}

	for i := 0; i < 10; i++ {
		<-outChan
	}
}

type customWorker struct {
	// TODO: Put some state here
}

func (worker *customWorker) Ready() bool {
	return true
}

func (worker *customWorker) Job(data interface{}) interface{} {
	/* TODO: Use and modify state
	 * there's no need for thread safety paradigms here unless the data is being accessed from
	 * another goroutine outside of the pool.
	 */
	if outputStr, ok := data.(string); ok {
		return ("custom job done: " + outputStr )
	}
	return nil
}

func TestCustomWorkers (t *testing.T) {
	outChan  := make(chan int, 10)
	runtime.GOMAXPROCS(runtime.NumCPU())

	workers := make([]SkankWorker, 4)
    for i, _ := range workers {
        workers[i] = &(customWorker{})
    }

	pool, errPool := CreateCustomPool(workers).Open()

	if errPool != nil {
		t.Errorf("Error starting pool: ", errPool)
		return
	}

	defer pool.Close()

	for i := 0; i < 10; i++ {
		/* Calling SendWork is thread safe, go ahead and call it from any goroutine.
		 * The call will block until a worker is ready and has completed the job.
		 */
		go func() {
			if value, err := pool.SendWork("hello world"); err == nil {
				if str, ok := value.(string); ok {
					if str != "custom job done: hello world" {
						t.Errorf("Unexpected output from custom worker")
					}
				} else {
					t.Errorf("Not a string!")
				}
			} else {
				t.Errorf("Error returned: ", err)
			}
			outChan <- 1
		}()
	}

	for i := 0; i < 10; i++ {
		<-outChan
	}
}
