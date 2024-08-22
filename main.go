package main

import (
	"context"
	_ "embed"
	"github.com/alecthomas/kingpin/v2"
	"github.com/blaubaer/talk-indicator/pkg/app"
	log "github.com/echocat/slf4g"
	"github.com/echocat/slf4g/native"
	_ "github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
	"github.com/echocat/slf4g/native/formatter"
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

	var a app.App

	cmd := kingpin.New(os.Args[0], "").
		Action(func(*kingpin.ParseContext) (rErr error) {
			if err := a.Initialize(); err != nil {
				return err
			}
			defer func() {
				if err := a.Dispose(); err != nil && rErr == nil {
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

			return a.Run(ctx)
		})
	a.SetupConfiguration(cmd)

	cmd.Flag("log.level", "").
		SetValue(lv.Level)
	cmd.Flag("log.format", "").
		Default("text").
		SetValue(lv.Consumer.Formatter)
	cmd.Flag("log.color", "").
		Default("auto").
		SetValue(lv.Consumer.Formatter.ColorMode)

	kingpin.MustParse(cmd.Parse(os.Args[1:]))
}
