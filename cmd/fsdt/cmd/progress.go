package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type progressUI struct {
	enabled   bool
	started   atomic.Bool
	stopped   atomic.Bool

	// status
	task    atomic.Value // string
	scanDir atomic.Value // string

	leftDone  atomic.Uint64
	leftTotal atomic.Uint64
	rightDone  atomic.Uint64
	rightTotal atomic.Uint64

	// rendering
	mu      sync.Mutex
	ticker  *time.Ticker
	stopCh  chan struct{}
	frames  []string
	frameIx int
}

func newProgressUI(enabled bool) *progressUI {
	p := &progressUI{enabled: enabled}
	p.frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	p.task.Store("")
	p.scanDir.Store("")
	return p
}

func (p *progressUI) Start(initial string) {
	if !p.enabled || p.started.Load() {
		return
	}
	p.task.Store(initial)
	p.stopCh = make(chan struct{})
	p.ticker = time.NewTicker(100 * time.Millisecond)
	p.started.Store(true)
	go p.loop()
}

func (p *progressUI) loop() {
	for {
		select {
		case <-p.ticker.C:
			p.renderOnce()
		case <-p.stopCh:
			p.clearLine()
			return
		}
	}
}

func (p *progressUI) renderOnce() {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	p.frameIx = (p.frameIx + 1) % len(p.frames)
	frame := p.frames[p.frameIx]
	p.mu.Unlock()

	task := p.getString(&p.task)
	dir := p.getString(&p.scanDir)
	ld, lt := p.leftDone.Load(), p.leftTotal.Load()
	rd, rt := p.rightDone.Load(), p.rightTotal.Load()

	var parts []string
	if task != "" {
		parts = append(parts, task)
	}
	if dir != "" {
		parts = append(parts, fmt.Sprintf("dir: %s", dir))
	}
	if lt > 0 || rt > 0 {
		parts = append(parts, fmt.Sprintf("files L %d/%d R %d/%d", ld, lt, rd, rt))
	}
	line := strings.Join(parts, "  ·  ")
	fmt.Fprintf(os.Stderr, "\r\x1b[2K%s %s", frame, line)
}

func (p *progressUI) clearLine() {
	if !p.enabled {
		return
	}
	fmt.Fprint(os.Stderr, "\r\x1b[2K")
}

func (p *progressUI) Stop() {
	if !p.enabled || !p.started.Load() || p.stopped.Load() {
		return
	}
	p.stopped.Store(true)
	close(p.stopCh)
	p.ticker.Stop()
	// Move to next line so following stdout prints are clean
	fmt.Fprintln(os.Stderr)
}

func (p *progressUI) getString(v *atomic.Value) string {
	if s, ok := v.Load().(string); ok {
		return s
	}
	return ""
}

// Setters used by the CLI flow
func (p *progressUI) SetTask(s string) { p.task.Store(s) }
func (p *progressUI) SetScanDir(s string) { p.scanDir.Store(s) }

func (p *progressUI) SetLeftTotal(n int)  { p.leftTotal.Store(uint64(n)) }
func (p *progressUI) SetRightTotal(n int) { p.rightTotal.Store(uint64(n)) }
func (p *progressUI) IncLeftDone()        { p.leftDone.Add(1) }
func (p *progressUI) IncRightDone()       { p.rightDone.Add(1) }