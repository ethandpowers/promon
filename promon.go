package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shirou/gopsutil/v4/process"
)

func findProcessByName(name string) *process.Process {
	processes, err := process.Processes()
	if err != nil {
		panic(err)
	}

	var matches []*process.Process
	for _, p := range processes {
		pName, err := p.Name()
		if err != nil {
			continue
		}
		var strippedString = name

		if len(name) > 2 {
			firstRune, _ := utf8.DecodeRuneInString(name)
			lastRune, _ := utf8.DecodeLastRuneInString(name)

			if firstRune == lastRune && (firstRune == '"' || firstRune == '\'') {
				strippedString = name[1 : len(name)-1]
			}
		}
		if strings.Contains(strings.ToLower(pName), strings.ToLower(strippedString)) {
			matches = append(matches, p)
		}
	}

	if len(matches) == 0 {
		fmt.Println("No such process")
		os.Exit(1)
	}

	if len(matches) > 1 {
		fmt.Printf("Matched %d processes\n", len(matches))
	}

	return matches[0]
}

func main() {
	pidFlag := flag.Int("pid", -1, "Process ID to monitor")
	flag.Parse()

	var match *process.Process
	if *pidFlag != -1 {
		proc, err := process.NewProcess(int32(*pidFlag))
		if err != nil {
			fmt.Println("No such process")
			os.Exit(1)
		}
		match = proc
	} else if flag.NArg() > 0 {
		processName := flag.Arg(0)
		match = findProcessByName(processName)
	} else {
		panic("You must provide either a --pid or a process name")
	}

	pid := match.Pid
	matchName, err := match.Name()
	if err != nil {
		fmt.Printf("Error getting process name: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Watching \x1b[32m%s (%d)\x1b[0m\n", matchName, pid)

	for {
		proc, err := process.NewProcess(pid)
		if err != nil {
			fmt.Println("Process went away")
			os.Exit(1)
		}

		memPercent, err := proc.MemoryPercent()
		if err != nil {
			fmt.Printf("Error getting memory info: %v\n", err)
		}

		cpuPercent, err := proc.CPUPercent()
		if err != nil {
			fmt.Printf("Error getting CPU info: %v\n", err)
		}

		fmt.Printf("Memory usage: %.2f%%	CPU usage: %.2f\n", memPercent, cpuPercent)
		time.Sleep(1 * time.Second)
		fmt.Printf("\033[1A\033[K") // Erase the previous print
	}
}
