package labee

import (
	"bufio"
	"errors"
	"os"
)

var (
	ErrNoArgs = errors.New("no arguments provided")
)

func pipeArgsAvailable() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func readPipeArgs() []string {
	scanner := bufio.NewScanner(os.Stdin)

	var pipeArgs []string
	for scanner.Scan() {
		pipeArgs = append(pipeArgs, scanner.Text())
	}

	return pipeArgs
}
