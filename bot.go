package main

import (
	"errors"
	"fmt"
	"io"
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
	tts            chan ircPRIVMSG
	mp3Writer      io.WriteCloser
	exit           chan struct{}
}

func (b *bot) Start() error {
	b.publicMessages = make(chan ircPRIVMSG)
	b.joins = make(chan ircJOIN)
	b.parts = make(chan ircPART)
	b.tts = make(chan ircPRIVMSG)

	b.exit = make(chan struct{})

	if err := b.connectIRC(); err != nil {
		return err
	}

	if err := b.startFFPlay(); err != nil {
		return err
	}

	go b.ttsLoop()
	go b.readLoop()

	return nil
}

func (b *bot) ttsLoop() {
	for msg := range b.tts {
		fmt.Printf("tts received: %s '%s'\n", msg.Nick, msg.Message)
		nick := strings.ReplaceAll(msg.Nick, "_", " ")
		text := msg.Message

		lang := "en"
		if len(text) > 3 {
			if text[2] == ':' {
				lang = text[:2]
				text = strings.TrimSpace(text[2:])
			}
		}

		says := "says"
		switch lang {
		case "es":
			says = "dice"
		case "fr":
			says = "dit"
		case "it":
			says = "dice"
		case "de":
			says = "sagt"
		case "tr":
			says = "viski iç"
		}

		text = fmt.Sprintf("%s %s %s", nick, says, text)

		if err := b.doTTS(lang, text); err != nil {
			log.Printf("ERR: %v\n", err)
			continue
		}
	}
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
	b.tts <- msg

	return nil
}

func (b *bot) doTTS(lang, text string) error {
	// cleanup
	text = strings.ReplaceAll(text, "judwhite", "judd white")
	text = strings.ReplaceAll(text, "soup steward", "soupy")
	text = strings.ReplaceAll(text, "degen", "di-jenn")
	text = strings.ReplaceAll(text, "elo", "e-low")
	text = strings.ReplaceAll(text, "@", " @ ")
	text = strings.ReplaceAll(text, "  ", " ")

	var (
		mp3 []byte
		err error
		ok  bool
	)

	for tries := 0; tries < 3 && !ok; tries++ {
		mp3, err, ok = textToMP3(lang, text)
		if tries >= 2 {
			if err != nil {
				return err
			} else if !ok {
				return errors.New("3 failed attempts at textToMP#")
			}
		}
	}

	return b.playMP3(mp3)
}

func (b *bot) startFFPlay() error {
	cmd := exec.Command("ffplay", "-nodisp", "-af", "atempo=1.1", "-")
	wc, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	b.mp3Writer = wc

	go func() {
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func (b *bot) playMP3(mp3 []byte) error {
	_, err := b.mp3Writer.Write(mp3)
	if err != nil {
		return fmt.Errorf("writeMP3: %w", err)
	}
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
