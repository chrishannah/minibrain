package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"minibrain/internal/agent"
)

func (m *tuiModel) appendAction(text string) {
	m.history = append(m.history, historyEntry{text: text, kind: "action"})
	m.refreshViewport()
}

func (m *tuiModel) appendPermission(text string) {
	m.history = append(m.history, historyEntry{text: text, kind: "assistant", bold: true})
	m.refreshViewport()
}

func (m *tuiModel) appendChoice(kind, question string, options []string) {
	m.choiceActive = true
	m.choiceKind = kind
	m.choiceIndex = 0
	m.history = append(m.history, historyEntry{text: question, kind: "choice", options: options, bold: true})
	m.refreshViewport()
}

func (m *tuiModel) appendText(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	m.history = append(m.history, historyEntry{text: text, kind: "assistant"})
	m.refreshViewport()
}

func (m *tuiModel) appendStream(delta string) {
	if delta == "" {
		return
	}
	if len(m.history) > 0 && m.history[len(m.history)-1].kind == "assistant_stream" {
		m.history[len(m.history)-1].text += delta
	} else {
		m.history = append(m.history, historyEntry{text: delta, kind: "assistant_stream"})
	}
	m.refreshViewport()
}

func (m *tuiModel) clearStream() {
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].kind == "assistant_stream" {
			m.history = append(m.history[:i], m.history[i+1:]...)
			break
		}
	}
	m.refreshViewport()
}

func (m *tuiModel) appendSecondary(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	m.history = append(m.history, historyEntry{text: text, kind: "assistant_secondary"})
	m.refreshViewport()
}

func (m *tuiModel) appendUser(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	m.history = append(m.history, historyEntry{text: text, kind: "user"})
	m.refreshViewport()
}

func (m *tuiModel) appendRunResult(res agent.Result) {
	if res.LLMOutput != "" {
		thinking, final := splitThinkingFinal(res.LLMOutput)
		if thinking != "" {
			m.appendSecondary(thinking)
		}
		if final != "" {
			m.appendText(final)
		}
	}
	for _, w := range res.AppliedWrites {
		m.appendAction("WRITE: " + w.Path)
	}
	for _, d := range res.AppliedDeletes {
		m.appendAction("DELETE: " + d.Path)
	}
	for _, p := range res.AppliedPatches {
		m.appendAction("PATCH: " + p.Path)
	}
	if res.Condensed {
		m.appendAction("MEMORY CONDENSED")
	}
}

func (m *tuiModel) refreshViewport() {
	var b strings.Builder
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary))
	secondaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary))
	first := true
	for _, h := range m.history {
		if h.kind == "action" && !m.showActions {
			continue
		}
		if !first {
			b.WriteString("\n\n")
		}
		first = false
		rendered := renderHistoryLine(h, m.viewport.Width, actionStyle, secondaryStyle, m.choiceActive, m.choiceIndex, m.mdRenderer)
		b.WriteString(rendered)
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func initialStats() (agent.MemoryStats, error) {
	brainDir, err := agent.ResolveBrainDir()
	if err != nil {
		return agent.MemoryStats{}, err
	}
	root, err := os.Getwd()
	if err != nil {
		return agent.MemoryStats{}, err
	}
	if err := agent.EnsureBrainLayout(brainDir, root); err != nil {
		return agent.MemoryStats{}, err
	}
	return agent.GetMemoryStats(brainDir, "", "")
}

func (m *tuiModel) updateMarkdownRenderer() {
	contentWidth := int(float64(m.viewport.Width) * 0.8)
	if contentWidth < 20 {
		contentWidth = m.viewport.Width
	}
	if contentWidth <= 0 {
		contentWidth = 80
	}
	if m.mdRenderer != nil && m.mdWidth == contentWidth {
		return
	}
	m.mdWidth = contentWidth
	m.mdRenderer = newMarkdownRenderer(contentWidth)
}

func initialUsage() (agent.UsageStats, error) {
	cfg, err := baseConfig()
	if err != nil {
		return agent.UsageStats{}, err
	}
	return agent.GetUsageStats(cfg)
}

func usageFromConfig() agent.UsageStats {
	cfg, err := baseConfig()
	if err != nil {
		return agent.UsageStats{}
	}
	usage, err := agent.GetUsageStats(cfg)
	if err != nil {
		return agent.UsageStats{}
	}
	return usage
}
