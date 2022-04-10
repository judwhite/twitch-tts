# twitch-tts

Twitch Text-to-Speech bot

# Compile instructions

1. Download the latest version of Go https://go.dev/dl/
2. Follow the installation instructions for your OS https://go.dev/doc/install.
3. Clone this repository `git clone https://github.com/judwhite/twitch-tts`
4. Go to the new directory `cd twitch-tts`
5. Compile the binary using `go build`

# Installation

## Install gTTS

Run `sudo apt install python3-gtts` to use your distribution's version of gTTS.

For a more recent version of gTTS, run:

```
sudo apt install python3-pip
sudo pip install gTTS
```

After installing you may need to create a new terminal session for `gtts-cli` to show up in the `PATH`.

## Install VLC command-line utilities

```
sudo apt install vlc-bin
```

Test that the `nvlc` binary is found.

## Create a Twitch Chat OAuth Token

1. The easy way: https://www.twitchapps.com/tmi/
2. `export TWITCH_IRC_OAUTH="<your_token>"` (add this to ~/.profile and run `source ~/.profile`)


# Running

```
./twitch-tts --nick=<bot_name> --channel=<#your_channel>
```

# Fork of a previous bot
Fork of https://github.com/judwhite/go-cah Cards Against Humanity IRC/Twitch Bot (2015).

## Resources
- [Twitch Chat OAuth Password Generator](http://www.twitchapps.com/tmi/)
- [Whisper Rate Limiting](https://discuss.dev.twitch.tv/t/whisper-rate-limiting/2836)
- [Twitch IRC](http://help.twitch.tv/customer/portal/articles/1302780-twitch-irc)
- [IRC - RFC2812](https://tools.ietf.org/html/rfc2812)
