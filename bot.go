package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
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

	if strings.HasPrefix(msg.Message, "!") {
		return b.processCommand(msg)
	}

	// extract language
	lang := "en"
	text := strings.TrimSpace(msg.Message)

	if len(msg.Message) > 3 {
		if msg.Message[2] == ':' {
			lang = msg.Message[:2]
			text = strings.TrimSpace(msg.Message[2:])
		}
	}

	text = fmt.Sprintf("%s says %s", msg.Nick, text)
	fmt.Printf("tts: %s\n", text)

	text = strings.ReplaceAll(text, "judwhite", "judd white")

	mp3bytes, err := textToMP3(lang, text)
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

func (b *bot) processCommand(msg ircPRIVMSG) error {
	var cmd, text string

	idx := strings.Index(msg.Message, " ")
	if idx == -1 {
		cmd = msg.Message
	} else {
		cmd = strings.TrimSpace(msg.Message[:idx])
		text = strings.TrimSpace(msg.Message[idx:])
	}

	switch cmd {
	case "!github":
		return b.irc.Say(msg.Channel, "https://github.com/judwhite")
	case "!lichess":
		return b.irc.Say(msg.Channel, "https://lichess.org/@/bantercode")
	case "!chess":
		return b.irc.Say(msg.Channel, "https://chess.com/play/judwhite")
	case "!f2c":
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		c := (f - 32) * 5 / 9
		return b.irc.Say(msg.Channel, fmt.Sprintf("%d°F = %d°C", int(f), int(c)))
	case "!c2f":
		c, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return err
		}
		f := (c * 9 / 5) + 32
		return b.irc.Say(msg.Channel, fmt.Sprintf("%d°C = %d°F", int(c), int(f)))
	}

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
