package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type bot struct {
	botCfg         *botConfig
	irc            *ircClient
	publicMessages chan ircPRIVMSG
	joins          chan ircJOIN
	parts          chan ircPART
	exit           chan struct{}
}

func (b *bot) Start() error {
	b.publicMessages = make(chan ircPRIVMSG)
	b.joins = make(chan ircJOIN)
	b.parts = make(chan ircPART)

	b.exit = make(chan struct{})

	if err := b.connectIRC(); err != nil {
		return err
	}

	go b.readLoop()

	return nil
}

func (b *bot) readLoop() {
loop:
	for {
		select {
		case msg := <-b.publicMessages:
			if err := b.processPRIVMSG(msg); err != nil {
				log.Println(err)
			}
		case _ = <-b.joins:
		//	b.processJOIN(join)
		case _ = <-b.parts:
		//	b.processPART(part)
		case <-b.exit:
			break loop
		}
	}
}

func (b *bot) processPRIVMSG(msg ircPRIVMSG) error {
	if !strings.HasPrefix(msg.Channel, "#") {
		fmt.Printf("!! channel = %s\n", msg.Channel)
		return nil
	}
	text := fmt.Sprintf("%s says %s", msg.Nick, msg.Message)
	mp3bytes, err := textToMP3(text)
	if err != nil {
		return err
	}

	go func() {
		cmd := exec.Command("nvlc", "--no-interact", "--play-and-exit", "-")
		cmd.Stdin = bytes.NewReader(mp3bytes)
		err := cmd.Start()
		if err != nil {
			log.Println(err)
		}
		_ = cmd.Wait()
	}()

	return nil
}

func (b *bot) connectIRC() error {
	whisperServerAddr, err := getWhisperServerAddress()
	if err != nil {
		return err
	}

	irc := &ircClient{
		ServerAddress:        "irc.twitch.tv:6667",
		WhisperServerAddress: whisperServerAddr,
		Nick:                 b.botCfg.nick,
		ServerPassword:       b.botCfg.serverPassword,
		PublicMessages:       b.publicMessages,
		Joins:                b.joins,
		Parts:                b.parts,
	}

	if err = irc.Connect(); err != nil {
		return err
	}

	for _, channel := range b.botCfg.channels {
		if err = irc.Join(channel); err != nil {
			// TODO: disconnect
			return err
		}
	}

	b.irc = irc
	return nil
}
