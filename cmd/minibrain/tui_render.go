package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

func renderHistoryLine(h historyEntry, viewportWidth int, actionStyle, secondaryStyle lipgloss.Style, choiceActive bool, choiceIndex int, md *glamour.TermRenderer) string {
	w := viewportWidth
	if w <= 0 {
		w = 80
	}
	contentWidth := int(float64(w) * 0.8)
	if contentWidth < 20 {
		contentWidth = w
	}

	switch h.kind {
	case "user":
		prefixed := "> " + h.text
		bubbleStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		bubble := bubbleStyle.Render(prefixed)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(secondaryStyle.Render(bubble))
	case "assistant_secondary":
		bubbleStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		rendered := renderMarkdown(h.text, md)
		bubble := bubbleStyle.Render(rendered)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(secondaryStyle.Render(bubble))
	case "assistant_stream":
		bubbleStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		rendered := renderMarkdown(h.text, md)
		bubble := bubbleStyle.Render(rendered)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(bubble)
	case "action":
		bubbleStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		label, body := actionLabel(h.text)
		prefix := lipgloss.NewStyle().Bold(true).Render(label)
		bubble := bubbleStyle.Render(prefix + ": " + body)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(actionStyle.Render(bubble))
	case "preview":
		bubbleStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Left).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorSecondary)).
			Padding(0, 1)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		bubble := bubbleStyle.Render(h.text)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(secondaryStyle.Render(bubble))
	case "raw":
		bubbleStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		bubble := bubbleStyle.Render(h.text)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(secondaryStyle.Render(bubble))
	case "choice":
		bubbleStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		question := bubbleStyle.Render(h.text)
		buttons := renderChoiceButtons(h.options, contentWidth, choiceActive, choiceIndex)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(question + "\n" + buttons)
	default:
		bubbleStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Left)
		if h.bold {
			bubbleStyle = bubbleStyle.Bold(true)
		}
		rendered := renderMarkdown(h.text, md)
		bubble := bubbleStyle.Render(rendered)
		return lipgloss.NewStyle().Width(w).Align(lipgloss.Left).Render(bubble)
	}
}

func actionLabel(text string) (string, string) {
	trim := strings.TrimSpace(text)
	label := ""
	body := trim
	if idx := strings.Index(trim, ":"); idx >= 0 {
		label = strings.ToUpper(strings.TrimSpace(trim[:idx]))
		body = strings.TrimSpace(trim[idx+1:])
	}
	upper := label
	if upper == "" {
		upper = strings.ToUpper(trim)
	}
	switch {
	case strings.HasPrefix(upper, "READ REQUEST"):
		return "Read request", body
	case strings.HasPrefix(upper, "READ "):
		return "Read", body
	case strings.HasPrefix(upper, "READ"):
		return "Read", body
	case strings.HasPrefix(upper, "WRITE "):
		return "Write", body
	case strings.HasPrefix(upper, "WRITE"):
		return "Write", body
	case strings.HasPrefix(upper, "DELETE "):
		return "Delete", body
	case strings.HasPrefix(upper, "DELETE"):
		return "Delete", body
	case strings.HasPrefix(upper, "PATCH "):
		return "Patch", body
	case strings.HasPrefix(upper, "PATCH"):
		return "Patch", body
	case strings.HasPrefix(upper, "MODEL "):
		return "Model", body
	case strings.HasPrefix(upper, "MODEL"):
		return "Model", body
	case strings.HasPrefix(upper, "ERROR"):
		return "Error", body
	case strings.HasPrefix(upper, "CHANGES "):
		return "Changes", body
	case strings.HasPrefix(upper, "CHANGES"):
		return "Changes", body
	case strings.HasPrefix(upper, "MEMORY "):
		return "Memory", body
	case strings.HasPrefix(upper, "MEMORY"):
		return "Memory", body
	case strings.HasPrefix(upper, "RAW OUTPUT"):
		return "Raw output", body
	default:
		return "Info", body
	}
}

func renderChoiceButtons(options []string, width int, active bool, selected int) string {
	if len(options) == 0 {
		return ""
	}
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	var lines []string
	for i, opt := range options {
		style := border.Foreground(lipgloss.Color(colorSecondary)).Padding(0, 1)
		if active && i == selected {
			style = border.Foreground(lipgloss.Color(colorPrimary)).Bold(true).Padding(0, 1)
		}
		lines = append(lines, style.Render(opt))
	}
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Left).Render(strings.Join(lines, "\n"))
}

func renderSuggestions(items []commandItem, index int) []string {
	if len(items) == 0 {
		return nil
	}
	index = clamp(index, 0, len(items)-1)
	var out []string
	for i, it := range items {
		line := fmt.Sprintf("%s  %s", it.cmd, it.desc)
		if i == index {
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}
		out = append(out, line)
	}
	return out
}

func renderDivider(width int) string {
	if width <= 0 {
		width = 80
	}
	return strings.Repeat("─", width)
}

func renderMarkdown(text string, r *glamour.TermRenderer) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return text
	}
	if r == nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return strings.TrimRight(out, "\n")
}

func formatBytes(n int) string {
	if n < 1024 {
		return fmt.Sprintf("(%dB)", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("(%.1fKB)", float64(n)/1024.0)
	}
	return fmt.Sprintf("(%.1fMB)", float64(n)/1024.0/1024.0)
}
