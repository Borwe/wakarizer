# <html>**<center>:clock1:<u>WAKARIZER</u>:clock1:**</center></html>

## Reason:

In my country Kenya it's quite common for people to post the wakatimes (time spent coding on editors), and came across this [X post](https://twitter.com/SamProgramiz/status/1696655932661387414) and one of it's [comments](https://twitter.com/Shifu_the_great/status/1696730247910109406) and was like, won't it be a fun project to make it possible for people to clock 24 hours coding times a day, LOL!

## Goal: 

Hacking wakatime :laughing::sweat_smile: and learning Go.

## How it Works:

It is a TUI application written in go using [Bubbletea](https://github.com/charmbracelet/bubbletea) library that once executed will produce heartbeats of the programming languages the user inputs when prompted and keep running forever until user explicitly closes the application.

With Go awesomeness, was able to use the [wakatime-cli](https://github.com/wakatime/wakatime-cli/) as if it was a library, and calling functions from inside it.



# Similarities with wakatime-cli

- Everything is pretty much similar, except this executable doesn't take any arguments, and interactions is via the TUI.
- Default cfg of wakatime is `~/.wakatime.cfg` but this one is `~/.wakarizer.cfg`



## Build from Source:

Just do the simple commands bellow.

```sh
go mod tidy && go build
```





## Contributions:

This is an opensource project, be free to fork, contribute Pull Requests, or even post on the issue tracker.