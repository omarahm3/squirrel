package client

import (
	"os"
	"reflect"
	"testing"

	"github.com/omarahm3/squirrel/internal/pkg/common"
	"go.uber.org/zap/zapcore"
)

const command = "squirrel"
const defaultPublicUrl = "https://localhost:3000"
const defaultWebsocketUrl = "wss://localhost:3000"

func TestInitOptions(t *testing.T) {
	tests := []struct {
		description string
		args        []string
		options     *ClientOptions
	}{
		{
			description: "default options",
			args:        []string{command},
			options: &ClientOptions{
				Env:          "prod",
				LogLevel:     zapcore.ErrorLevel,
				PeerId:       "",
				Listen:       false,
				Output:       false,
				UrlClipboard: false,
				Domain: &common.Domain{
					Public:    defaultPublicUrl,
					Websocket: defaultWebsocketUrl,
				},
			},
		},
		{
			description: "change log level",
			args:        []string{command, "-log", "debug"},
			options: &ClientOptions{
				Env:          "prod",
				LogLevel:     zapcore.DebugLevel,
				PeerId:       "",
				Listen:       false,
				Output:       false,
				UrlClipboard: false,
				Domain: &common.Domain{
					Public:    defaultPublicUrl,
					Websocket: defaultWebsocketUrl,
				},
			},
		},
		{
			description: "change domain prod",
			args:        []string{command, "-domain", "mrg.sh"},
			options: &ClientOptions{
				Env:          "prod",
				LogLevel:     zapcore.DebugLevel, // Debug level since it was changed it in the test before
				PeerId:       "",
				Listen:       false,
				Output:       false,
				UrlClipboard: false,
				Domain: &common.Domain{
					Public:    "https://mrg.sh",
					Websocket: "wss://mrg.sh",
				},
			},
		},
		{
			description: "change domain dev",
			args:        []string{command, "-domain", "mrg.sh", "-env", "dev"},
			options: &ClientOptions{
				Env:          "dev",
				LogLevel:     zapcore.DebugLevel, // Debug level since it was changed it in the test before
				PeerId:       "",
				Listen:       false,
				Output:       false,
				UrlClipboard: false,
				Domain: &common.Domain{
					Public:    "http://mrg.sh",
					Websocket: "ws://mrg.sh",
				},
			},
		},
		{
			// This is needed since the default values of env, domain, and logLevel
			// is determined based on global variables that we need to reset
			description: "reset defaults",
			args:        []string{command, "-log", "error", "-domain", "localhost:3000", "-env", "prod"},
			options: &ClientOptions{
				Env:          "prod",
				LogLevel:     zapcore.ErrorLevel,
				PeerId:       "",
				Listen:       false,
				Output:       false,
				UrlClipboard: false,
				Domain: &common.Domain{
					Public:    defaultPublicUrl,
					Websocket: defaultWebsocketUrl,
				},
			},
		},
		{
			description: "listen without peerId",
			args:        []string{command, "-listen"},
			options: &ClientOptions{
				Env:          "prod",
				LogLevel:     zapcore.ErrorLevel,
				PeerId:       "",
				Listen:       true,
				Output:       false,
				UrlClipboard: false,
				Domain: &common.Domain{
					Public:    defaultPublicUrl,
					Websocket: defaultWebsocketUrl,
				},
			},
		},
		{
			description: "shorthand listen with peerId and url copied to clipboard",
			args:        []string{command, "-l", "-u", "-peer", "some-id"},
			options: &ClientOptions{
				Env:          "prod",
				LogLevel:     zapcore.ErrorLevel,
				PeerId:       "some-id",
				Listen:       true,
				Output:       false,
				UrlClipboard: true,
				Domain: &common.Domain{
					Public:    defaultPublicUrl,
					Websocket: defaultWebsocketUrl,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			oldArgs := os.Args
			defer ResetTesting(oldArgs)

			os.Args = test.args
			options := InitOptions()

			if !reflect.DeepEqual(options, test.options) {
				t.Errorf("expected %v, actual %v", test.options, options)
			}
		})
	}
}
