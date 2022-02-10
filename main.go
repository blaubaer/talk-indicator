package main

import (
	"context"
	_ "embed"
	"github.com/blaubaer/talk-indicator/pkg/app"
	log "github.com/echocat/slf4g"
	"github.com/echocat/slf4g/native"
	_ "github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
	"github.com/echocat/slf4g/native/formatter"
	"github.com/getlantern/systray"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	lv := value.NewProvider(native.DefaultProvider)
	lv.Consumer.Formatter.Codec = value.MappingFormatterCodec{
		"text": formatter.NewText(func(v *formatter.Text) {
			bv := true
			v.AllowMultiLineMessage = &bv
			v.MultiLineMessageAfterFields = &bv
		}),
		"json": formatter.NewJson(),
	}

	var app app.App

	cmd := kingpin.New(os.Args[0], "").
		Action(func(*kingpin.ParseContext) (rErr error) {
			if err := app.Initialize(); err != nil {
				return err
			}
			defer func() {
				if err := app.Dispose(); err != nil && rErr == nil {
					rErr = err
				}
			}()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go func() {
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Interrupt, syscall.SIGTERM)
				<-c
				log.Info("Terminated. Going down...")
				cancel()
			}()

			return app.Run(ctx)
		})
	app.SetupConfiguration(cmd)

	cmd.Flag("log.level", "").
		SetValue(lv.Level)
	cmd.Flag("log.format", "").
		Default("text").
		SetValue(lv.Consumer.Formatter)

	kingpin.MustParse(cmd.Parse(os.Args[1:]))
}

//go:embed iconwin.ico
var icon []byte

func onExit() {
	log.Info("Bye!")
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("MicRedLight")
	systray.SetTooltip("MicRedLight")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}
