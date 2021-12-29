package main
// compile: go build -o supasmuggle supasmuggle.go
// run: ./supasmuggle <file.txt>
// OR
// go run supasmuggle.go -f <file.txt> -t 10000000 (but not that many) 

/*
 * Bugs:
 *	- Not 100% sure the CommandContext timeout is working
 * 	- It's not better
 * 
 * To Do:
 *	Add flags to enable exhaustive mode in smuggler.py as well as not stop after finding one vuln (should be real easy)
 */

import (
	"fmt"
	"flag"
	"os"
	"os/exec"
	"log"
	"context"
	"bufio"
	"sync"
	"time"
	"strings"
	"encoding/json"

	"github.com/fatih/color"
)

type Results struct {
	Host string
	Payload string
	Error string
}

// pretty colors
var success = color.New(color.FgMagenta).SprintFunc()
var successmsg = color.New(color.FgMagenta).PrintfFunc()
var report = color.New(color.FgGreen).SprintFunc()
var fail = color.New(color.FgRed).SprintFunc()
var warn = color.New(color.FgYellow).PrintfFunc()

// format scan time
func timer(t time.Duration, o string) {
	tps_report := fmt.Sprintf("\nScan completed in %s seconds\n", report(t.Round(time.Millisecond)))
	fmt.Println(tps_report)

	f, err := os.OpenFile(o, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Fprintln(f, tps_report)
}

func main() {
	// args parsing duh
	var sec int
	flag.IntVar(&sec, "s", 360, "Specify the time (in seconds) to wait before moving on to next host")

	var concurrency int
	flag.IntVar(&concurrency, "t", 50, "Specify the size of the resource pool")

	var debug bool
	flag.BoolVar(&debug, "d", false, "Show the actual output of smuggler.py")

	var outfile string
        t := time.Now()
	fname := fmt.Sprintf("supa_%v-%v-%v_%v.json", t.Hour(), t.Minute(), t.Second(), t.Year())
	flag.StringVar(&outfile, "o", fname, "Specify an output file")

	var file string
	flag.StringVar(&file, "f", "THERE IS NO SPOON", "File containing URLs to look HRS vulnerabilities on")
	flag.Parse()

	// training wheels
	if (sec < 60) {
		warn("ERROR: It's not recommended to reduce timeout below 1 minute as you'll miss potential vulns!\n")
		os.Exit(1)
	}

	// begin supafast stuff
	// thanks @tomnomnom for httprobe, where I learned how to concurrency
	/* ok let's start from the top. Here we create a waitgroup for our tasks, and an input and output channel. ok */
	t1 := time.Now()
	var tasksWG sync.WaitGroup
	tasks := make(chan string)
	output := make(chan Results)

	/* then we spawn 50 go routines and 50 wg counters. so far so good */
	for i := 0; i < concurrency; i++ {
		tasksWG.Add(1)

		// one goroutine per wg counter. nothing wrong here
		go func() {
			// check out line 154 real quick. That's where we are pushing tasks into the input channel
			// so in the main thread we are adding tasks into the channel and in a separate thread we range through the tasks being added
			// I suppose I see an issue perhaps. maybe this needs to be in the main thread and adding them needs to be in a go routine?
			for t := range tasks {
				// nothing wrong here. This is the processing part. 
				resp, err := smuggler(t, sec, debug)
				if err != nil {
					continue
				}
				output <- resp // in the same thread we are pulling tasks from the input channel, we add the processed response to an output channel
				continue 	// which is totally fine
			}
			tasksWG.Done() /* this is fine as we are decrementing the counter as the goroutine finished */
		}()
	}

	// wait for all the output to come through the channel
	// There shouldn't be anything wrong here?
	var outputWG sync.WaitGroup
	outputWG.Add(1)
	go func() {
		// Final output
		f, err := os.OpenFile(outfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		for o := range output {
			if o.Payload == "" {
				fmt.Printf("Scanned %s %s\n", report(o.Host), fail(o.Error))
			} else{
				successmsg("Potential vulnerability found: %s\n", success(o.Host))
				fmt.Printf("Payload: %s\n", o.Payload)

				// format output to JSON
				js, err := json.Marshal(o)
				if err != nil {
					log.Fatal(err)
				}

				// Then write it to file (whether you like it or not)
				fmt.Fprintf(f, "%s\r\n", string(js))
				//if _, err = f.WriteString(string(js)+"\n"); err != nil {
				//	log.Fatal(err)
				//}
			}
		}
		outputWG.Done()
	}()


	// waiting for tasks to complete
	go func() {
		tasksWG.Wait()
		close(output)
	}()

	// open the file
	f, err := os.Open(file)
	if err != nil {
		log.Panic(err)
	}

	// here in the main thread of execution we scan our file line by line and add each line to the input channel. ok now go back to line 93
	s := bufio.NewScanner(f)
	for s.Scan() {
		tasks <- s.Text()
	}

	// because the I was waiting on tasks in a go routine this runs almost immediately which means I'm not executing any of the tasks
	// even though it does indeed run everything? I don't get it
	close(tasks)
	outputWG.Wait()
	timer(time.Since(t1), outfile)

}

// I lika... do... dah cha cha
func smuggler(t string, sec int, debug bool) (Results, error) {
	var r Results
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	pybin := fmt.Sprintf("%s/resources/smuggler/smuggler.py", pwd)
	// time out smuggler.py if it takes too long
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(sec) * time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, pybin, "-x", "-u", t)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return r, err
	}
	// start smuggler.py
	cmd.Start()

	// run it line by line
	s := bufio.NewScanner(stdout)
	for s.Scan() {
		l := s.Text()
		// show all the output of smuggler.py (you psycho)
		if debug {
			fmt.Println(l)
		}

		// check for connection error 
		if strings.Contains(l, "Unable to connect to host") {
			r.Error = strings.Split(l, ":")[1]
		}

		// otherwise, if we found something
		if strings.Contains(l, "CRITICAL") {
			f := strings.Fields(l)
			r.Payload = f[5]
		}
		r.Host = t
	}
	cmd.Wait()
	return r, err
}