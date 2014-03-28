Skank
=====

Skank is a golang library for creating and managing a thread pool of fixed size, aiming to be simple, intuitive, ground breaking, revolutionary and world dominating.

Use cases for skank are any situation where a large flood of jobs are imminent, potentially from different threads, and you need to bottleneck those jobs through a fixed number of dedicated worker threads. The most obvious example is as an easy wrapper for limiting the hard work done in your software to the number of CPU's available, preventing the threads from foolishly competing with each other for CPU time.

##How to use:

Here's a simple example of skank being used to distribute a batch of calculations to a pool of workers that matches the number of CPU's:

```golang
...

func CalcRoots (inputs []float64) []float64 {
    numCPUs  := runtime.NumCPU()
    numJobs  := len(inputs)
    doneChan := make( chan int,  numJobs )
    outputs  := make( []float64, numJobs )

    runtime.GOMAXPROCS(numCPUs)

    /* Create the pool, and specify the job each worker should perform, if each worker needs
     * to carry its own state then this can also be accomplished, read on.
     */
    pool, errPool := skank.CreatePool(numCPUs, func( object interface{} ) ( interface{} ) {
        if value, ok := object.(float64); ok {
            // Hard work here
            return math.Sqrt(value)
        }
        return nil
    }).Open()

    if errPool != nil {
        fmt.Fprintln(os.Stderr, "Error starting pool: ", errPool)
        return nil
    }

    defer pool.Close()

    /* Creates a go routine for all jobs, these will be blocked until a worker is available
     * and has finished the request.
     */
    for i := 0; i < numJobs; i++ {
        go func(index int) {
            // SendWork is thread safe. Go ahead and call it from any goroutine
            if value, err := pool.SendWork(inputs[index]); err == nil {
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

This particular example, since it all resides in the one func, could actually be done with less code by simply spawning numCPU's goroutines that gobble up a shared channel of float64's. This would probably also be quicker since you waste cycles here boxing and unboxing the job values, but at least you don't have to write it all yourself you lazy scum.

##So where do I actually benefit from using skank?

You don't, I'm not a god damn charity.
