package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"sync"
)

func textToMP3(lang, text string) ([]byte, error) {
	const cmdName = "gtts-cli"
	cmdArgs := []string{"-l", lang, "-o", "-", "-f", "-"}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	cmd.Stdin = bytes.NewReader([]byte(text))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderrpipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdoutpipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("!! ERROR: cmd.Start: %T %v\n", err, err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// stderr loop
	go func() {
		defer wg.Done()
		r := bufio.NewScanner(stderr)
		for r.Scan() {
		}
		if err := r.Err(); err != nil {
			if strings.Contains(err.Error(), "file already closed") {
				// no-op
			} else {
				fmt.Printf("!! ERROR: stderr loop: %T %v\n", err, err)
			}
		}
	}()

	// stdout loop
	var mp3 []byte
	go func() {
		defer wg.Done()

		var err2 error
		mp3, err2 = ioutil.ReadAll(stdout)
		if err2 != nil {
			log.Fatalf("stdout: ioutil.ReadAll: %v", err2)
		}
	}()

	if err := cmd.Wait(); err != nil {
		errMsg := err.Error()
		if errMsg == "exit status 1" || errors.Is(err, context.DeadlineExceeded) || errMsg == "signal: killed" || errMsg == "signal: interrupt" {
			// no-op
		} else {
			return nil, fmt.Errorf("!! ERROR: cmd.Wait: %T %v\n", err, err)
		}
	}

	wg.Wait()

	return mp3, nil
}
