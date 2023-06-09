package labee

import (
	"errors"
	"os"
	"os/exec"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/gookit/color"
)

var errFzfReturn = errors.New("exit status 130")

func openInteractiveMode(strs []string, preview string) error {
	cmd := exec.Command("fzf", "-m", "--ansi", "--height", "40%", "--border", "--preview", preview)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		color.Danger.Println(err)
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return errors.New("interactive mode requires fzf to be installed and located in $PATH")
	}

	for _, s := range strs {
		color.Fprintln(pipe, s)
	}

	err = cmd.Wait()
	// !TODO: better way to check against the exit code
	if err != nil && err.Error() != errFzfReturn.Error() {
		return err
	}

	return nil
}

func openInteractiveFileMode(files []database.File) error {
	strs := []string{}
	for _, f := range files {
		strs = append(strs, f.Path)
	}
	return openInteractiveMode(strs, "labee info {}")
}
