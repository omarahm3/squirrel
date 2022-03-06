# 🐿️ Squirrel
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)[![Github: CI](https://img.shields.io/github/workflow/status/omarahm3/switcher/CI/master)](https://github.com/omarahm3/switcher/actions)[![Twitter: omarahm3](https://img.shields.io/twitter/follow/omarahm3.svg?style=social)](https://twitter.com/omarahm3)

Realtime logs (or any stdout basically!) sharing by just piping squirrel.

![web-view](https://user-images.githubusercontent.com/8606113/156921849-2ae60d76-be21-4271-adf3-fe20fd9a75f5.gif)
![terminal-view](https://user-images.githubusercontent.com/8606113/156922080-5e881e4c-6999-4f18-a04b-5454bba7965f.gif)

## Why?
- I'm lazy
- I wanted to learn Go
- I wanted to implement a websocket server using Go
- I wanted to be used on crossplatform without installing any extra dependencies (that's why not Nodejs or ink)
- I'm tired of requesting people to send me snippets of logs, so that i can analyze, or grep the lines
- I love CLI apps
- Probably people will find this useful?
- I wanted to use squirrel as a name of something xD
- Because why not?

## Setup
Currently can head over to [releases](https://github.com/omarahm3/squirrel/releases), on the latest release, and choose the suitable package for you. Currently this was tested on Ubuntu and OSX Monterey only though

If you're on Linux, then you can choose the package based on your distribution and install it, for example on Ubuntu/Debian:

```bash
curl -fLO https://github.com/omarahm3/squirrel/releases/download/v<VERSION>/squirrel_<VERSION>_linux_amd64.deb
sudo apt install -y ./squirrel_<VERSION>_linux_amd64.deb
```

If you're on a different OS, then you can use download the compressed files, and move these files to your bin folder accordingly, for example on Ubuntu/Debian:

```bash
curl -fLO https://github.com/omarahm3/squirrel/releases/download/v<VERSION>/squirrel_<VERSION>_Linux_x86_64.tar.gz
tar xf squirrel_<VERSION>_Linux_x86_64.tar.gz
mv ./squirrel ~/.local/bin/
```

Or if you have go installed, you can always install squirrel client as a Go package by doing so:

```bash
go install github.com/omarahm3/squirrel/main/squirrel
```

***Note**: Installing squirrel as a go package will install it as a development ready package, which is going to need to be configured either by Environment variables or passing options to the CLI, Refer to [configuration](#Configuration) section for more details*

## Squirreling
Squirrel has no commands, the only way it can be used is via piping it to stdout of another command, for example 

```bash
$ for i in $(seq 1 50); do echo "Example Message #$i"; sleep 1; done | squirrel -o -u

➜ ID: [ 315c77cd-7ac1-4487-adf8-d205471f0771 ]
➜ Link: [ https://squirrel-jwls9.ondigitalocean.app/client/315c77cd-7ac1-4487-adf8-d205471f0771 ]
➜ Url is copied to your clipboard
📢 Squirrel is waiting for listeners to begin piping stdout...

```

This will print a shareable link and ID that you can use to send to the person you want to share the output of this for loop with, however squirrel will not start reading stdout till the other end is connected and ready to receive the actual output.

The other end (or maybe yourself) can then open the link and you'll begin to see that messages are coming in, which are the output of the for loop we just piped. Or if person prefer to use terminal, then another squirrel can be used in listening mode, but then `peer` option is must be supplied with the ID of the broadcaster:

```bash
squirrel -l --peer=315c77cd-7ac1-4487-adf8-d205471f0771
```

## Configuration
Squirrel can be configured by passing options/flags to the CLI, or for some options you can use ENV variables as well. Just note that ENV variables have more priority over flags.
Squirrel can be run in 2 different modes too:
- Broadcasting: which is the default behavior and the initial mode of squirrel where _broadcaster_ is piping squirrel to stdout
- Listening: in which the other end _subscriber_ which is the one who is listening for _broadcaster_ messages 

### Client configuration
**ENV Variables:**
- `APP_ENV` - Set the app environment mode (`prod` or `dev` default is `prod`)
- `DOMAIN` - Set the server domain in which CLI is going to send events to
- `LOG_LEVEL` - Set the current log level of the CLI (default is `error`)
	- Log levels are:
		- error
		- warn
		- info
		- debug

**Flags:**
- `--env` - Set app environment mode (same as `APP_ENV`)
- `--domain` - Set the server domain in which CLI is going to send events to (same as `DOMAIN`)
- `--log` - Set the current log level of the CLI (same as `DOMAIN`)
- `--peer` - Peer (broadcaster) ID that squirrel is going to listen to (must be supplied in listen mode `-l/--listen`)
- `-l` or `--listen` - Set the current mode of the CLI to listen instead of broadcasting
- `-o` or `--show-output` - Show the output of what is being piped to squirrel on the current session as well
- `-u` or `--copy-url` - Copy shareable link to the clipboard

You can always run:
```bash
squirrel -h
```
and see all of the available options

## Squirreld (Server)
This package is built into 2 different applications, the main one and probably most of the users will be interested in is _squirrel_ which is the client/CLI, and the other one is _squirreld_ which is the server daemon that will be responsible of managing and routing broadcasters messages to their corresponding subscribers.

Squirreld is a websocket server, that each of the broadcasters and subscribers is connecting to, so that they can exchange events between each other, that's is how we're making sure that stdout messages are exchanged in realtime from broadcasters to subscribers.
Currently server is hosted by me on one of Digitalocean servers, but this is subject to change indeed.

## Squirreld Configuration
All of server configuration can be tweaked using ENV variables or passing flags to squirreld, here is the detailed options and ENV variables list:
- `--env` or `APP_ENV` - Set server environment mode (`prod` or `dev` default is `prod`)
- `--domain` or `DOMAIN` - Set the current server domain
- `--log` or `LOG_LEVEL` - Set the current log level of the server (same as squirrel log levels)
- `--port` or `PORT` - Set the current port that server is going to listen to (default is `3000`)
- `--read-buffer-size` or `READ_BUFFER_SIZE` - Websocket server read buffer size (default is `0`)
- `--write-buffer-size` or `WRITE_BUFFER_SIZE` - Websocket server write buffer size (default is `0`)
- `--max-message-size` or `MAX_MESSAGE_SIZE` - Websocket server maximum message size (default is `1024`)

Same as squirrel, ENV variables have more priority than flags as well.

## Note
This is pretty immature Go project, i'm still learning Go by actually doing and maintaining this project, it gave me the opportunity to explore various topics that i want to get familiar with using Go such as backend (http, and websocket), templates, CLI, Go routines and channels ..etc
Contributions are more than welcome, i indeed would like to see how this project will scale.
