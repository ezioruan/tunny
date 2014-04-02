![Skank](http://www.creepybit.co.uk/images/skank_logo_small.png "Skank")

Skank is a golang library for creating and managing a thread pool, aiming to be simple, intuitive, ground breaking, revolutionary, world dominating and also trashy.

Use cases for skank are any situation where a large flood of jobs are imminent, potentially from different threads, and you need to bottleneck those jobs through a fixed number of dedicated worker threads. The most obvious example is as an easy wrapper for limiting the hard work done in your software to the number of CPU's available, preventing the threads from foolishly competing with each other for CPU time.

##How to install:

```bash

go get github.com/jeffail/skank

```

##How to use:

Here's a simple example of skank being used to distribute a batch of calculations to a pool of workers that matches the number of CPU's:

```go
...

import "github.com/jeffail/skank"

...

func CalcRoots (inputs []float64) []float64 {
    numCPUs  := runtime.NumCPU()
    numJobs  := len(inputs)
    doneChan := make( chan int,  numJobs )
    outputs  := make( []float64, numJobs )

    runtime.GOMAXPROCS(numCPUs)

    /* Create the pool, and specify the job each worker should perform,
	 * if each worker needs to carry its own state then this can also
	 * be accomplished, read on.
     */
    pool, err := skank.CreatePool(numCPUs, func( object interface{} ) ( interface{} ) {
        if value, ok := object.(float64); ok {
            // Hard work here
            return math.Sqrt(value)
        }
        return nil
    }).Open()

    if err != nil {
        fmt.Fprintln(os.Stderr, "Error starting pool: ", err)
        return nil
    }

    defer pool.Close()

    /* Creates a go routine for all jobs, these will be blocked until
	 * a worker is available and has finished the request.
     */
    for i := 0; i < numJobs; i++ {
        go func(index int) {
            // SendWork is thread safe. Go ahead and call it from any go routine
            if value, err2 := pool.SendWork(inputs[index]); err2 == nil {
                if result, ok := value.(float64); ok {
                    outputs[index] = result
                }
            }
            doneChan <- 1
        }(i)
    }

    // Wait for all jobs to be completed before closing the pool
    for i := 0; i < numJobs; i++ {
        <-doneChan
    }

    return outputs
}

...

```

This particular example, since it all resides in the one func, could actually be done with less code by simply spawning numCPU's go routines that gobble up a shared channel of float64's. This would probably also be quicker since you waste cycles here boxing and unboxing the job values, but at least you don't have to write it all yourself you lazy scum.

##Can I specify the job for each work call?

Yes, the arguments passed to the worker are boxed as interface{}, so this can actually be a func, you can implement this yourself, or if you're not bothered about return values you can use:

```go
...

exampleChannel := make(chan int)

pool, _ := skank.CreatePoolGeneric(numCPUs).Open()

pool.SendWork(func() {
	/* Do your hard work here, usual rules of enclosures apply here,
	 * so you can return values like so:
	 */
	exampleChannel <- 10
})

...
```

##Specify a time out period

To make pool calls adhere to a timeout period of your choice simply swap the call to SendWork with SendWorkTimed, like so:

```go
...

// SendWorkTimed takes an argument for a timeout in milliseconds.
// If this timeout triggers the call will return with an error
if value, err := pool.SendWorkTimed(500, inputs[index]); err == nil {
	if result, ok := value.(float64); ok {
		outputs[index] = result
	}
} else {
/* A timeout most likely occured, I haven't checked this specifically because
 * I am a lazy garbage mongler.
 */
}

...
```

This snippet will send the job, and wait for up to 500 milliseconds for an answer. You could optionally implement a timeout yourself by starting a new go routine that returns the output through a channel, and having that channel compete with time.After().

You'd be an idiot for doing that though because you would be forcing the pool to send work to a worker even if the timeout occured whilst waiting for a worker to become available, you muppet!

##How do I give my workers state?

The call to skank.CreatePool will generate a pool of SkankWorkers for you, and then assign each worker the closure argument to run for each job. You can, however, create these workers yourself, thereby allowing you to also give them their own state and methods.

Here is a short example:

```go
...

type customWorker struct {
    // TODO: Put some state here
}

// Use this call to block further jobs if necessary
func (worker *customWorker) Ready() bool {
    return true
}

// This is where the work actually happens
func (worker *customWorker) Job(data interface{}) interface{} {
    /* TODO: Use and modify state
     * there's no need for thread safety paradigms here unless the
	 * data is being accessed from another go routine outside of
	 * the pool.
     */
    if outputStr, ok := data.(string); ok {
        return ("custom job done: " + outputStr )
    }
    return nil
}

func TestCustomWorkers (t *testing.T) {
    outChan  := make(chan int, 10)

    workers := make([]skank.SkankWorker, 4)
    for i, _ := range workers {
        workers[i] = &(customWorker{})
    }

    pool, errPool := skank.CreateCustomPool(workers).Open()

    if errPool != nil {
        t.Errorf("Error starting pool: ", errPool)
        return
    }

    defer pool.Close()

    for i := 0; i < 10; i++ {
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

...
```

You'll notice that as well as the important Job(data interface{}) interface{} call to implement there is also the call Ready() bool. Ready is potentially an important part of the SkankWorker that allows you to use your state to determine whether or not this worker should take on another job yet, and answer true or false accordingly.

For example, your worker could hold a counter of how many jobs it has done, and perhaps after a certain amount it should perform another act before taking on more work, it's important to use Ready for these occasions since blocking the Job call will hold up the waiting client.

You can block Ready whilst you wait for some condition to change, or alternatively you can return a true/false straight away, in this case the call will be repeated at 50 millisecond intervals until you answer true.

##So where do I actually benefit from using skank?

You don't, I'm not a god damn charity.
