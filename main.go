package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// ./twitch-tts --nick=<botname> --channel=<#yourchannel>

	args := os.Args[1:] // for live play
	err := startBot(args)
	if err != nil {
		log.Fatal(err)
	}
}

func startBot(args []string) error {
	botCfg, err := parseArgs(args)
	if err != nil {
		return err
	}

	bot := bot{botCfg: botCfg}
	if err = bot.Start(); err != nil {
		return err
	}

	// wait for ^C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	return nil
}
