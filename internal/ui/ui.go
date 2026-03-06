package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/farhapartex/loadforge/internal/loader"
)

type tickMsg time.Time
type metricsMsg loader.MetricsSnapshot
type workerCountMsg int

type doneMsg struct{}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		waitForMetrics(m.metricsCh),
		waitForDone(m.doneCh),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, tea.Batch(tickCmd(), waitForMetrics(m.metricsCh))

	case metricsMsg:
		snap := loader.MetricsSnapshot(msg)
		m.snapshot = snap
		m.pushRPS(snap.RPS)
		m.started = true
		return m, waitForMetrics(m.metricsCh)

	case workerCountMsg:
		m.activeWorkers = int(msg)
		return m, nil

	case doneMsg:
		m.done = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.done = true
			return m, tea.Quit
		case "e":
			m.showErrors = !m.showErrors
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing...\\n"
	}

	var sections []string

	sections = append(sections, m.renderHeader())
	statsPanel := m.renderStats()
	latencyPanel := m.renderLatency()

	middleRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		statsPanel,
		"  ",
		latencyPanel,
	)
	sections = append(sections, middleRow)
	sections = append(sections, m.renderSparkline())

	if m.snapshot.TotalFailures > 0 {
		sections = append(sections, m.renderErrors())
	}

	sections = append(sections, m.renderFooter())

	return strings.Join(sections, "\\n")
}

func (m *Model) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	hints := "  q quit    e toggle errors"
	if m.done {
		hints = "  test complete — press q to exit"
	}
	return footerStyle.Render(hints)
}

func (m *Model) renderErrors() string {
	if len(m.snapshot.Errors) == 0 {
		return ""
	}

	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	var lines []string
	count := 0
	for msg, n := range m.snapshot.Errors {
		if count >= 3 {
			lines = append(lines, metaStyle.Render(fmt.Sprintf("  ... and %d more error types", len(m.snapshot.Errors)-3)))
			break
		}
		lines = append(lines, errStyle.Render(fmt.Sprintf("  [%d] %s", n, truncate(msg, 80))))
		count++
	}

	content := headerStyle.Render("⚠  Errors") + "\\n" + strings.Join(lines, "\\n")
	return boxStyle.Render(content)
}

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func waitForMetrics(ch <-chan loader.MetricsSnapshot) tea.Cmd {
	return func() tea.Msg {
		select {
		case snap, ok := <-ch:
			if !ok {
				return doneMsg{}
			}
			return metricsMsg(snap)
		default:
			return nil
		}
	}
}

func waitForDone(ch <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		<-ch
		return doneMsg{}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func Run(model Model) error {
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}

func UpdateWorkerCount(p *tea.Program, count int) {
	p.Send(workerCountMsg(count))
}
