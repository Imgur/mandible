package imageprocessor

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"
)

const GM_COMMAND = "convert"

func convertToJpeg(filename string) (string, error) {
	outfile := strings.Sprintf("%s_q", filename)

	args := []string{
		filename,
		"-flatten",
		"JPEG:" + outfile,
	}

	err := runConvertCommand(args)
	if err != nil {
		return nil, err
	}
}

func quality(filename string, quality int) error {
	outfile := strings.Sprintf("%s_q", filename)

	args := []string{
		filename,
		"-quality",
		quality,
		"-density",
		"72x72",
		outfile,
	}

	err := runConvertCommand(args)
	if err != nil {
		return nil, err
	}
}

func resize_percent(filename string, percent int) (string, error) {
	args := []string{
		filename,
		"-resize",
		strings.Sprintf("%i%", percent),
		outfile,
	}

	err := runConvertCommand(args)
	if err != nil {
		return nil, err
	}
}

func runConvertCommand(args []string) error {
	cmd := exec.Command(GM_COMMAND, args)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Start()

	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Duration(500000) * time.Millisecond):
		killCmd(cmd)
		<-cmdDone
		return errors.New("Command command timed out")
	case err := <-cmdDone:
		if err != nil {
			log.Println(stderr.String())
		}
		return err
	}

	return nil
}

func killCmd(cmd *exec.Cmd) {
	if err := cmd.Process.Kill(); err != nil {
		log.Printf("Failed to kill command: %v", err)
	}
}
