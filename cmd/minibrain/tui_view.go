package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m tuiModel) View() string {
	header := lipgloss.NewStyle().Bold(true).Render(asciiHeader())
	statusBar := lipgloss.NewStyle().Width(m.width).Render(renderStatusBar(m))

	inputLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary)).Render("Prompt (Enter to run)")
	inputBox := lipgloss.NewStyle().Width(m.width).Render(m.input.View())

	suggestions := ""
	items := currentSuggestions(m)
	if len(items) > 0 {
		lines := renderSuggestions(items, m.suggestIndex)
		suggestions = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary)).Render(strings.Join(lines, "\n"))
	}

	var b strings.Builder
	b.WriteString(header + "\n\n")
	b.WriteString(m.viewport.View() + "\n\n")
	b.WriteString(renderDivider(m.width) + "\n")
	b.WriteString(inputLabel + "\n")
	b.WriteString(inputBox + "\n\n")
	if suggestions != "" {
		b.WriteString(suggestions + "\n\n")
	}
	b.WriteString(statusBar)

	return b.String()
}

func asciiHeader() string {
	return "\n" +
		" _____ ______   ___  ________   ___  ________  ________  ________  ___  ________      \n" +
		"|\\   _ \\  _   \\|\\  \\|\\   ___  \\|\\  \\|\\   __  \\|\\   __  \\|\\   __  \\|\\  \\|\\   ___  \\    \n" +
		"\\ \\  \\\\\\__\\ \\  \\ \\  \\ \\  \\\\ \\  \\ \\  \\ \\  \\|\\ /\\ \\  \\|\\  \\ \\  \\|\\  \\ \\  \\ \\  \\\\ \\  \\   \n" +
		" \\ \\  \\\\|__| \\  \\ \\  \\ \\  \\\\ \\  \\ \\  \\ \\   __  \\ \\   _  _\\ \\   __  \\ \\  \\ \\  \\\\ \\  \\  \n" +
		"  \\ \\  \\    \\ \\  \\ \\  \\ \\  \\\\ \\  \\ \\  \\ \\  \\|\\  \\ \\  \\\\  \\\\ \\  \\ \\  \\ \\  \\ \\  \\\\ \\  \\ \n" +
		"   \\ \\__\\    \\ \\__\\ \\__\\ \\__\\\\ \\__\\ \\__\\ \\_______\\ \\__\\\\ _\\\\ \\__\\ \\__\\ \\__\\ \\__\\\\ \\__\\\n" +
		"    \\|__|     \\|__|\\|__|\\|__| \\|__|\\|__|\\|_______|\\|__|\\|__|\\|__|\\|__|\\|__|\\|__| \\|__|\n"
}

func renderStatusBar(m tuiModel) string {
	activity := m.status
	if strings.TrimSpace(activity) == "" {
		activity = "Ready"
	}
	if m.err != nil {
		activity = "Error"
	}

	model := m.model
	if model == "" {
		model = "gpt-4.1"
	}

	actions := "Actions on"
	if !m.showActions {
		actions = "Actions off"
	}
	ctxUsage := fmt.Sprintf("Ctx ~%d/%d tok", m.usage.ApproxTokens, m.usage.BudgetTokens)
	stats := "Long-term Memory " + formatBytes(m.stats.LtmBytes) + " | Short-term Memory " + formatBytes(m.stats.StmBytes) + " | " + ctxUsage + " | " + actions

	full := activity + " | " + model + " | " + stats
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary))
	if activity != "Ready" {
		style = style.Bold(true)
	}
	return style.Render(full)
}
