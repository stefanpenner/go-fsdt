package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

type statusMsg string

type progressModel struct {
	sp      spinner.Model
	status  string
	quitting bool
}

func (m progressModel) Init() tea.Cmd {
	return m.sp.Tick
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ignore key input; non-interactive
		return m, nil
	case statusMsg:
		m.status = string(msg)
		return m, nil
	default:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	}
}

func (m progressModel) View() string {
	if m.quitting {
		return ""
	}
	// Render to stderr; actual output is handled by Bubble Tea program options
	return fmt.Sprintf("\r%s %s", m.sp.View(), m.status)
}

type progressUI struct {
	mu       sync.Mutex
	program  *tea.Program
	enabled  bool
	closed   bool
}

func newProgressUI(enabled bool) *progressUI {
	return &progressUI{enabled: enabled}
}

func (p *progressUI) Start(initial string) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.program != nil {
		return
	}
	sp := spinner.New()
	m := progressModel{sp: sp, status: initial}
	prog := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	p.program = prog
	go func() { _, _ = prog.Run() }()
	// Give the renderer a moment to start so first status is visible
	time.Sleep(30 * time.Millisecond)
}

func (p *progressUI) UpdateStatus(format string, a ...any) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	prog := p.program
	p.mu.Unlock()
	if prog == nil {
		return
	}
	prog.Send(statusMsg(fmt.Sprintf(format, a...)))
}

func (p *progressUI) Stop() {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.program == nil {
		return
	}
	p.program.Quit()
	p.closed = true
	// Ensure the spinner line is cleared
	fmt.Fprint(os.Stderr, "\r\x1b[2K")
}