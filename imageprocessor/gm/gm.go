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

func FixOrientation(filename string) (string, error) {
	outfile := fmt.Sprintf("%s_ort", filename)

	args := []string{
		filename,
		"-auto-orient",
		outfile,
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
		fmt.Sprintf("%d", quality),
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
		fmt.Sprintf("%d%%", percent),
		outfile,
	}

	err := runConvertCommand(args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}

func SquareThumb(filename, name string, size int) (string, error) {
	outfile := fmt.Sprintf("%s_%s", filename, name)

	args := []string{
		fmt.Sprintf("%s[0]", filename),
		"-quality",
		"94",
		"-resize",
		fmt.Sprintf("%dx%d^", size, size),
		"-gravity",
		"center",
		"-crop",
		fmt.Sprintf("%dx%d+0+0", size, size),
		"-density",
		"72x72",
		"-unsharp",
		"0.5",
		fmt.Sprintf("JPG:%s", outfile),
	}

	err := runConvertCommand(args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}

func Thumb(filename, name string, width, height int) (string, error) {
	outfile := fmt.Sprintf("%s_%s", filename, name)

	args := []string{
		fmt.Sprintf("%s[0]", filename),
		"-quality",
		"83",
		"-resize",
		fmt.Sprintf("%dx%d>", width, height),
		"-density",
		"72x72",
		fmt.Sprintf("JPG:%s", outfile),
	}

	err := runConvertCommand(args)
	if err != nil {
		return "", err
	}

	return outfile, nil
}

func CircleThumb(filename, name string, width int) (string, error) {
	outfile := fmt.Sprintf("%s_%s", filename, name)

	filename, err := Thumb(filename, name, width*2, width*2)
	if err != nil {
		return "", err
	}

	args := []string{
		"-size",
		fmt.Sprintf("%dx%d", width, width),
		"xc:none",
		"-fill",
		filename,
		"-quality",
		"83",
		"-density",
		"72x72",
		"-draw",
		fmt.Sprintf("circle %d,%d %d,1", width/2, width/2, width/2),
		fmt.Sprintf("PNG:%s", outfile),
	}

	err = runConvertCommand(args)
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
		return errors.New("Command timed out")
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
