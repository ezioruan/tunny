package skank

import (
	"testing"
	"time"
	"runtime"
	"fmt"
)

func validateReturnInt ( t *testing.T, expecting int, object interface{} ) {
	if w, ok := object.(int); ok {
		if w != expecting {
			t.Errorf("Wrong, expected %v, got %v", expecting, w)
		}
	} else {
		t.Errorf("Wrong, expected int")
	}
}

func TestBasic (t *testing.T) {
	sizePool, repeats, sleepFor, margin := 16, 2, 1, 0.1
	outChan  := make(chan int, sizePool)

	runtime.GOMAXPROCS(runtime.NumCPU())

	pool := CreatePool(sizePool, func( object interface{} ) ( interface{} ) {
		time.Sleep(time.Duration(sleepFor) * time.Second)
		if w, ok := object.(int); ok {
			return w * 2
		}
		return "Not an int!"
	}).Begin()

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

	taken    := float64( time.Since(before) ) / float64(time.Second)
	expected := ( float64(sleepFor) + margin ) * float64(repeats)

	if taken > expected {
		t.Errorf("Wrong, should have taken less than %v seconds, actually took %v", expected, taken)
	}
}

func TestExampleCase (t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	pool := CreatePool(4, func( object interface{} ) ( interface{} ) {

		if str, ok := object.(string); ok {
			return "job done: " + str
		}
		return nil

	}).Begin()

	for i := 0; i < 10; i++ {
		go func() {
			if value, err := pool.SendWork("hello world"); err == nil {
				if str, ok := value.(string); ok {
					fmt.Println(str)
				} else {
					t.Errorf("Not a string!")
				}
			} else {
				t.Errorf("Error returned: ", err)
			}
		}()
	}
}
