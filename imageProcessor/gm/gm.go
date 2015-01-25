package gm

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"
)

const GM_COMMAND = "convert"

func ConvertToJpeg(filename string) (string, error) {
	outfile := fmt.Sprintf("%s_jpg", filename)

	args := []string{
		filename,
		"-flatten",
		"JPEG:" + outfile,
	}

	err := runConvertCommand(args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}

func Quality(filename string, quality int) (string, error) {
	outfile := fmt.Sprintf("%s_q", filename)

	args := []string{
		filename,
		"-quality",
		string(quality),
		"-density",
		"72x72",
		outfile,
	}

	err := runConvertCommand(args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}

func ResizePercent(filename string, percent int) (string, error) {
	outfile := fmt.Sprintf("%s_rp", filename)

	args := []string{
		filename,
		"-resize",
		fmt.Sprintf("%i%", percent),
		outfile,
	}

	err := runConvertCommand(args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}

func runConvertCommand(args []string) error {
	cmd := exec.Command(GM_COMMAND, args...)

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
