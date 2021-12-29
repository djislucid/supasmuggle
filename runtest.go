package main

// so you should be calling on WaitGroup.Add(1) for each go routine you are calling
// then once you are done with that task you should be calling WaitGroup.Done()
// 
// So one go routine requires the use of 1 wg.Add(1) and 1 wg.Done()
// then, wg.Wait() should be called in the main thread to wait for all go routines to finish
// hm...

import (
	"sync"
	"time"
	"os"
	"fmt"
	"log"
	"bufio"
)

func main() {
	var taskWG sync.WaitGroup
	concurrency := 10
	tasks := make(chan string)
	output := make(chan string)

	for i := 0; i < concurrency; i++ {
		// this is where you process tasks from the channel
		// maybe it's ok if this happens in the main thread?
		processStuff(i)
	}

	taskWG.Add(1)
	go func() {
		defer taskWG.Done()
		f, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}

		sc := bufio.NewScanner(f)
		for sc.Scan() {
			tasks <- sc.Text()
		}
	}()
	taskWG.Wait()
	close(tasks)
}

func processStuff(i int) {
	time.Sleep(3 * time.Second)
	fmt.Println(i)
}
