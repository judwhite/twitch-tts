package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
)

func textToMP3(text string) ([]byte, error) {
	const cmdName = "gtts-cli"
	cmdArgs := []string{"-l", "en", "-o", "-", "-f", "-"}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	cmd.Stdin = bytes.NewReader([]byte(text))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("%v", err)
	}

	var mp3 []byte
	go func() {
		var err2 error
		mp3, err2 = ioutil.ReadAll(stdout)
		if err2 != nil {
			log.Fatal(err2)
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("!! ERROR: cmd.Start: %T %v\n", err, err)
	}

	if err := cmd.Wait(); err != nil {
		errMsg := err.Error()
		if errMsg == "exit status 1" || errors.Is(err, context.DeadlineExceeded) || errMsg == "signal: killed" || errMsg == "signal: interrupt" {
			// no-op
		} else {
			return nil, fmt.Errorf("!! ERROR: cmd.Wait: %T %v\n", err, err)
		}
	}

	return mp3, nil
}
