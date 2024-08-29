package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	log "github.com/echocat/slf4g"
	"github.com/echocat/slf4g/native"
	_ "github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/consumer"
	"github.com/echocat/slf4g/native/facade/value"
	"github.com/echocat/slf4g/native/formatter"
	"github.com/getlantern/systray"

	"github.com/blaubaer/talk-indicator/pkg/app"
	"github.com/blaubaer/talk-indicator/pkg/common"
	"github.com/blaubaer/talk-indicator/pkg/console"
	ps "github.com/blaubaer/talk-indicator/pkg/signal"
)

func main() {
	wf := &writerFacade{delegates: []io.Writer{os.Stdout}}
	buf := common.NewRingLineBuffer(2000, 4096)
	buf.TruncateTooLongLines = true
	consumer.Default = consumer.NewWriter(wf)

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
	a.OtherSignals = []ps.Signal{&ps.Systray{
		IconOn:  micOnIcon,
		IconOff: micOffIcon,
	}}

	cmd := kingpin.New(os.Args[0], "").
		Action(func(*kingpin.ParseContext) error {
			if err := a.Initialize(); err != nil {
				return err
			}
			systray.Run(func() {
				defer func() { _ = a.Dispose() }()

				systray.SetIcon(micOffIcon)
				systray.SetTitle("Talk indicator")
				showConsoleMi := systray.AddMenuItem("Show Console", "Shows the console with more information.")
				quitMi := systray.AddMenuItem("Exit", "Exit the talk indicator")

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var consoleCloser atomic.Pointer[context.CancelFunc]

				go func() {
					c := make(chan os.Signal, 1)
					signal.Notify(c, os.Interrupt, syscall.SIGTERM)
					for {
						select {
						case <-showConsoleMi.ClickedCh:
							for {
								if cl := consoleCloser.Load(); cl != nil {
									(*cl)()
									showConsoleMi.SetTitle("Show Console")
									showConsoleMi.SetTooltip("Shows the console with more information.")
									break
								} else {
									shCtx, shCancel := context.WithCancel(ctx)
									if !consoleCloser.CompareAndSwap(nil, &shCancel) {
										shCancel()
										continue
									}
									showConsoleMi.SetTitle("Hide Console")
									showConsoleMi.SetTooltip("Hide the currently opened console.")
									go showConsole(shCtx, buf, wf, func() {
										shCancel()
										consoleCloser.Store(nil)
									})
									break
								}
							}
						case <-c:
							log.Info("Terminated. Going down...")
							cancel()
						case <-quitMi.ClickedCh:
							log.Info("Exit clicked. Going down...")
							cancel()
						}
					}
				}()

				wf.set([]io.Writer{buf})
				a.Run(ctx)
				os.Exit(0)
			}, nil)
			return nil
		})
	a.SetupConfiguration(cmd)

	cmd.Flag("log.level", "").
		SetValue(lv.Level)
	cmd.Flag("log.format", "").
		Default("text").
		SetValue(lv.Consumer.Formatter)
	cmd.Flag("log.color", "").
		Default("always").
		SetValue(lv.Consumer.Formatter.ColorMode)

	kingpin.MustParse(cmd.Parse(os.Args[1:]))
}

func showConsole(bCtx context.Context, buf *common.RingLineBuffer, wf *writerFacade, finalizer func()) {
	defer finalizer()
	fail := func(err error) {
		log.WithError(err).
			Warn("Cannot create console.")
	}

	dc, err := console.NewDedicatedConsole("Talk Indicator")
	if err != nil {
		fail(err)
		return
	}
	defer func() { _ = dc.Close() }()

	wf.set([]io.Writer{buf, dc.Stdout}, func(current, next []io.Writer) {
		_, _ = buf.WriteTo(dc.Stdout)
	})
	defer wf.set([]io.Writer{buf})

	ctx, cancelFunc := context.WithCancel(bCtx)
	defer cancelFunc()

	dc.OnCtrlC = func(event any) bool {
		cancelFunc()
		return false
	}

	<-ctx.Done()
}

type writerFacade struct {
	delegates []io.Writer
	mutex     sync.RWMutex
}

func (this *writerFacade) Write(p []byte) (n int, err error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	for i, w := range this.delegates {
		var nn int
		if nn, err = w.Write(p); err != nil {
			return n, err
		}
		if i == 0 {
			n = nn
		} else if n != nn {
			return n, fmt.Errorf("the previous writer wrote %d, but the current one wrote %d bytes", nn, n)
		}
	}

	return
}

func (this *writerFacade) set(next []io.Writer, whileChange ...func(current, next []io.Writer)) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	current := this.delegates
	for _, fn := range whileChange {
		fn(current, next)
	}
	this.delegates = next
}

var (
	//go:embed assets/mic-off.ico
	micOffIcon []byte
	//go:embed assets/mic-on.ico
	micOnIcon []byte
)
