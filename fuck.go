package main

import (
	"fmt"
	"os"
	"log"
	"bufio"
	"sync"
	"time"
)

func main() {
	concurrency := 50
	var tasksWG sync.WaitGroup
	tasks := make(chan string)
	output := make(chan string)

	// ok, so we spawn 50 go routines ands increment the waitgroup counter by 50. Good to go
	for i := 0; i < concurrency; i++ {
		tasksWG.Add(1)

		go func() {
			// inside of each goroutine we loop through the tasks queue and execute a process, pushing the result into an output chan
			// it does seem weird to loop through tasks inside of a loop
			for t := range tasks {
				output <- smuggler(t)
				continue
			}
			// otherwise though, this is all good. Overall the waitgroup structure in the concurrency loop is fine
			tasksWG.Done()
		}()
	}

	// we create another waitgroup to wait for all the output to finish
	var outputWG sync.WaitGroup
	// we create a singular goroutine and increment the counter by one
	outputWG.Add(1)
	go func() {
		// in the separate thread we print the output
		for o := range output {
			fmt.Println(o)
		}
		// and once the output thread is totally complete we decrement. This is fine
		outputWG.Done()
	}()

	// some weirdness here. Why do we wait for the taskgroup to finish in a separate thread
	go func() {
		tasksWG.Wait()
		close(output)
	}()

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Panic(err)
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		tasks <- s.Text()
	}

	// we close the tasks in the main thread. presumably after the tasksWG is done
	// what I don't get is how that works if we wait in a separate thread as well...
	close(tasks)
	// this waits in the main thread for all the output to be printed. This is fine
	outputWG.Wait()
}

func smuggler(t string) string {
	time.Sleep(3 * time.Second)
	return t
}
