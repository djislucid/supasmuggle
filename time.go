package main

import (
	"time"
	"log"
	"fmt"
	"os/exec"
	"bufio"
)

func main() {
	t1 := time.Now()

	cmd := exec.Command("sleep", "5")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Start()
	s := bufio.NewScanner(stdout)
	for s.Scan() {
		fmt.Println("hello")
	}

	cmd.Wait()
	final := (time.Since(t1))
	fmt.Printf("Time elapsed: %s\n", final.Round(time.Millisecond))

}
