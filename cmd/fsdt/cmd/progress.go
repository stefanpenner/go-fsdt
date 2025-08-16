package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

type progressModel struct {
	sp       spinner.Model
	enabled  bool

	// state (accessed only on Bubble Tea goroutine)
	task    string
	scanDir string
	leftDone, leftTotal   uint64
	rightDone, rightTotal uint64
}

type setTaskMsg string

type setScanDirMsg string

type setLeftTotalMsg uint64

type setRightTotalMsg uint64

type incLeftDoneMsg struct{}

type incRightDoneMsg struct{}

func newProgressModel(enabled bool, initial string) progressModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return progressModel{sp: sp, enabled: enabled, task: initial}
}

func (m progressModel) Init() tea.Cmd {
	if !m.enabled {
		return nil
	}
	return m.sp.Tick
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.enabled {
		return m, nil
	}
	switch v := msg.(type) {
	case tea.KeyMsg:
		return m, nil
	case tea.WindowSizeMsg:
		return m, nil
	case setTaskMsg:
		m.task = string(v)
		return m, nil
	case setScanDirMsg:
		m.scanDir = string(v)
		return m, nil
	case setLeftTotalMsg:
		m.leftTotal = uint64(v)
		return m, nil
	case setRightTotalMsg:
		m.rightTotal = uint64(v)
		return m, nil
	case incLeftDoneMsg:
		m.leftDone++
		return m, nil
	case incRightDoneMsg:
		m.rightDone++
		return m, nil
	default:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	}
}

func (m progressModel) View() string {
	if !m.enabled {
		return ""
	}
	var parts []string
	if m.task != "" {
		parts = append(parts, m.task)
	}
	if m.scanDir != "" {
		parts = append(parts, fmt.Sprintf("dir: %s", m.scanDir))
	}
	if m.leftTotal > 0 || m.rightTotal > 0 {
		parts = append(parts, fmt.Sprintf("files L %d/%d R %d/%d", m.leftDone, m.leftTotal, m.rightDone, m.rightTotal))
	}
	line := strings.Join(parts, "  ·  ")
	return fmt.Sprintf("\r\x1b[2K%s %s", m.sp.View(), line)
}

type progressUI struct {
	enabled  bool
	program  *tea.Program
	started  atomic.Bool
	stopped  atomic.Bool
}

func newProgressUI(enabled bool) *progressUI {
	return &progressUI{enabled: enabled}
}

func (p *progressUI) Start(initial string) {
	if !p.enabled || p.started.Load() {
		return
	}
	m := newProgressModel(p.enabled, initial)
	prog := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	p.program = prog
	p.started.Store(true)
	go func() { _, _ = prog.Run() }()
}

func (p *progressUI) Stop() {
	if !p.enabled || !p.started.Load() || p.stopped.Load() {
		return
	}
	p.stopped.Store(true)
	if p.program != nil {
		p.program.Quit()
	}
	// Clear the line and add a newline to separate from stdout output
	fmt.Fprint(os.Stderr, "\r\x1b[2K\n")
}

// Async setters via Bubble Tea messages
func (p *progressUI) SetTask(s string) {
	if p.enabled && p.program != nil {
		p.program.Send(setTaskMsg(s))
	}
}

func (p *progressUI) SetScanDir(s string) {
	if p.enabled && p.program != nil {
		p.program.Send(setScanDirMsg(s))
	}
}

func (p *progressUI) SetLeftTotal(n int) {
	if p.enabled && p.program != nil {
		p.program.Send(setLeftTotalMsg(n))
	}
}

func (p *progressUI) SetRightTotal(n int) {
	if p.enabled && p.program != nil {
		p.program.Send(setRightTotalMsg(n))
	}
}

func (p *progressUI) IncLeftDone() {
	if p.enabled && p.program != nil {
		p.program.Send(incLeftDoneMsg{})
	}
}

func (p *progressUI) IncRightDone() {
	if p.enabled && p.program != nil {
		p.program.Send(incRightDoneMsg{})
	}
}