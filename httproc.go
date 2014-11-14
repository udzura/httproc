package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/pipe.v2"
)

func runScanLoop(scanner *bufio.Scanner) {
	for scanner.Scan() {
		if text := scanner.Text(); text != "\"\"" {
			fmt.Printf("%v\n", text)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading command input:", err)
	}
}

func runWatchCmdLoop(cmd *exec.Cmd, term chan bool) {
	cmd.Wait()
	fmt.Println("Process exit detected")
	term <- true
}

func main() {
	ruby := "puts 'Start'; while str = gets; exit if str.start_with?('quit'); p str.chomp.reverse; end"
	cmd := exec.Command("ruby", "-e", ruby)
	stdout, err := cmd.StdoutPipe()
	stdin, err2 := cmd.StdinPipe()
	if err != nil || err2 != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// Invoke first input
	fmt.Fprint(stdin, "\n")
	scanner := bufio.NewScanner(stdout)
	go runScanLoop(scanner)

	term := make(chan bool, 1)
	go runWatchCmdLoop(cmd, term)

	go func() {
		<-term
		fmt.Println("Exited...")
		os.Exit(0)
	}()

	for {
		p := pipe.Line(
			pipe.Read(os.Stdin),
			pipe.Write(stdin),
		)
		if err := pipe.Run(p); err != nil {
			panic(err)
		}
		if s, err := pipe.Output(p); err != nil {
			fmt.Println(s)
			panic(err)
		}
	}
}
