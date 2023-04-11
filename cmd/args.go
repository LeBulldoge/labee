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

func readPipeArgs() (pipeArgs []string) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		pipeArgs = append(pipeArgs, scanner.Text())
	}

	return
}
