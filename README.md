# twitch-tts

Twitch Text-to-Speech bot

# Install instruction for Ubuntu 20+

```
export TWITCH_OAUTH_TOKEN=<your_token>

sudo apt install pip
sudo pip install gTTS
sudo apt install vlc-bin

git clone https://github.com/judwhite/twitch-tts
cd twitch-tts
go build

./twitch-tts --nick=<bot_name> --channel=<#your_channel>
```

# Fork of a previous bot
Fork of https://github.com/judwhite/go-cah Cards Against Humanity IRC/Twitch Bot (2015).

## Resources
- [Twitch Chat OAuth Password Generator](http://www.twitchapps.com/tmi/)
- [Whisper Rate Limiting](https://discuss.dev.twitch.tv/t/whisper-rate-limiting/2836)
- [Twitch IRC](http://help.twitch.tv/customer/portal/articles/1302780-twitch-irc)
- [IRC - RFC2812](https://tools.ietf.org/html/rfc2812)
