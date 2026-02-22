package main

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/chrishannah/minibrain/internal/agent"
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
	m.input.SetValue("")
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

func (m *tuiModel) clearThinking() {
	for i := len(m.history) - 1; i >= 0; i-- {
		h := m.history[i]
		if h.kind == "assistant_secondary" && strings.TrimSpace(h.text) == "Thinking..." {
			m.history = append(m.history[:i], m.history[i+1:]...)
			break
		}
	}
	m.thinkingActive = false
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
	raw := res.RawOutput
	if raw == "" {
		raw = res.LLMOutput
	}
	if raw != "" {
		m.appendRaw(raw)
	}
	text := res.Message
	if strings.TrimSpace(text) == "" {
		text = res.LLMOutput
	}
	if text != "" {
		thinking, final := splitThinkingFinal(text)
		if thinking != "" {
			m.appendSecondary(thinking)
		}
		if final != "" {
			m.appendText(final)
		}
	}
	for _, w := range res.AppliedWrites {
		m.appendAction(formatAction(ActionWrite, w.Path))
	}
	for _, d := range res.AppliedDeletes {
		m.appendAction(formatAction(ActionDelete, d.Path))
	}
	for _, p := range res.AppliedPatches {
		m.appendAction(formatAction(ActionPatch, p.Path))
	}
	for _, p := range res.FailedPatches {
		m.appendAction(formatAction(ActionPatchFailed, p.Path+" ("+p.Reason+")"))
	}
	if res.Condensed {
		m.appendAction(formatAction(ActionMemory, "CONDENSED"))
	}
}

func (m *tuiModel) appendRaw(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	m.history = append(m.history, historyEntry{text: "RAW RESPONSE:\n" + text, kind: "raw"})
	m.refreshViewport()
}

func (m *tuiModel) appendPreview(text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	m.history = append(m.history, historyEntry{text: text, kind: "preview"})
	m.refreshViewport()
}

func formatPreviewBlock(kind, path string, lines []string) string {
	var b strings.Builder
	title := strings.TrimSpace(kind + " " + path)
	if title == "" {
		title = "Preview"
	}
	b.WriteString("Preview: " + title)
	if len(lines) > 0 {
		b.WriteString("\n")
		for i, line := range lines {
			if i == 0 {
				b.WriteString(line)
			} else {
				b.WriteString("\n")
				b.WriteString(line)
			}
		}
	}
	return b.String()
}

func appendChangePreview(m *tuiModel) {
	const maxLines = 12
	for _, p := range m.pendingPatches {
		lines := strings.Split(p.Patch, "\n")
		if len(lines) > maxLines {
			lines = append(lines[:maxLines], "...")
		}
		m.appendPreview(formatPreviewBlock("PATCH", p.Path, lines))
	}
	for _, w := range m.pendingWrites {
		lines := strings.Split(w.Content, "\n")
		if len(lines) > maxLines {
			lines = append(lines[:maxLines], "...")
		}
		m.appendPreview(formatPreviewBlock("WRITE", w.Path, lines))
	}
	for _, d := range m.pendingDeletes {
		m.appendPreview(formatPreviewBlock("DELETE", d.Path, nil))
	}
}

func (m *tuiModel) refreshViewport() {
	var b strings.Builder
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorPrimary))
	secondaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary))
	first := true
	for _, h := range m.history {
		if h.kind == "action" && !m.showActions {
			continue
		}
		if h.kind == "raw" && !m.showRaw {
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
