package main

import (
	"bufio"
	"errors"
	"os"
)

var (
	ErrNoArgs = errors.New("no arguments provided")
)

func Filter[I any](slice []I, predicate func(I) bool) (result []I) {
	for i := 0; i < len(slice); i++ {
		s := slice[i]
		if predicate(s) {
			result = append(result, s)
		}
	}

	return result
}

func Map[I any, O any](slice []I, m func(I) O) (result []O) {
	for i := 0; i < len(slice); i++ {
		result = append(result, m(slice[i]))
	}

	return result
}

func pipeArgsAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func readPipeArgs() (pipeArgs []string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		pipeArgs = append(pipeArgs, scanner.Text())
	}

	return
}
