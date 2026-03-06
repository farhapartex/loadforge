package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
var sparkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
var sparkLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

func (m *Model) renderSparkline() string {
	history := m.rpsHistory

	if len(history) == 0 {
		empty := strings.Repeat("▁", sparklineSize)
		content := headerStyle.Render("RPS Over Time") + "\\n" +
			sparkStyle.Render(empty) + "\\n" +
			sparkLabelStyle.Render("  0                                              now")
		return boxStyle.Render(content)
	}

	maxRPS := 0.0
	for _, v := range history {
		if v > maxRPS {
			maxRPS = v
		}
	}

	padded := make([]float64, sparklineSize)
	start := sparklineSize - len(history)
	if start < 0 {
		start = 0
	}
	copy(padded[start:], history)

	var sb strings.Builder
	for _, v := range padded {
		ch := sparkChar(v, maxRPS)
		sb.WriteRune(ch)
	}

	currentRPS := 0.0
	if len(history) > 0 {
		currentRPS = history[len(history)-1]
	}

	label := fmt.Sprintf("  current: %.1f req/s   max: %.1f req/s", currentRPS, maxRPS)

	content := headerStyle.Render("📈 RPS Over Time") + "\\n" +
		sparkStyle.Render(sb.String()) + "\\n" +
		sparkLabelStyle.Render(label)

	return boxStyle.Render(content)
}

func sparkChar(value, max float64) rune {
	if max == 0 || value == 0 {
		return sparkChars[0]
	}
	ratio := value / max
	idx := int(ratio * float64(len(sparkChars)-1))
	if idx >= len(sparkChars) {
		idx = len(sparkChars) - 1
	}
	return sparkChars[idx]
}
