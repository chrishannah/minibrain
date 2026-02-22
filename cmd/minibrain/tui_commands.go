package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/chrishannah/minibrain/internal/agent"
	"github.com/chrishannah/minibrain/internal/userconfig"
)

type commandItem struct {
	cmd  string
	desc string
}

type choiceBlock struct {
	question string
	options  []string
}

func listenStream(ch <-chan streamMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return streamMsg{done: true}
		}
		return msg
	}
}

func startAgentStream(m *tuiModel, prompt string, allowRead, allowWrite bool, readPaths []string) tea.Cmd {
	ch := make(chan streamMsg)
	m.streamCh = ch
	go func() {
		var res agent.Result
		var err error
		if len(readPaths) > 0 {
			res, err = runAgentStreamWithAllowAndReads(prompt, allowRead, allowWrite, readPaths, func(delta string) {
				ch <- streamMsg{delta: delta}
			})
		} else {
			res, err = runAgentStreamWithAllow(prompt, allowRead, allowWrite, func(delta string) {
				ch <- streamMsg{delta: delta}
			})
		}
		ch <- streamMsg{done: true, res: res, err: err}
		close(ch)
	}()
	return listenStream(ch)
}

func runMemoryCmd(prompt string) tea.Cmd {
	cmd := strings.ToLower(strings.TrimSpace(prompt))
	switch cmd {
	case "/clear":
		return func() tea.Msg {
			cfg, err := baseConfig()
			if err != nil {
				return memMsg{err: err}
			}
			if err := agent.ClearShortTerm(cfg); err != nil {
				return memMsg{err: err}
			}
			stats, _ := agent.GetMemoryStats(cfg.BrainDir, cfg.NeoPath, cfg.PrefrontalPath)
			return memMsg{action: "MEMORY CLEARED", stats: stats}
		}
	case "/condense":
		return func() tea.Msg {
			cfg, err := baseConfig()
			if err != nil {
				return memMsg{err: err}
			}
			_, err = agent.CondenseShortTerm(cfg)
			if err != nil {
				return memMsg{err: err}
			}
			stats, _ := agent.GetMemoryStats(cfg.BrainDir, cfg.NeoPath, cfg.PrefrontalPath)
			return memMsg{action: "MEMORY CONDENSED", stats: stats, condensed: true}
		}
	default:
		return func() tea.Msg {
			return memMsg{action: "UNKNOWN COMMAND: " + cmd}
		}
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func splitThinkingFinal(s string) (string, string) {
	trim := strings.TrimSpace(s)
	if trim == "" {
		return "", ""
	}
	lower := strings.ToLower(trim)
	if strings.HasPrefix(lower, "plan:") ||
		strings.HasPrefix(lower, "plan\n") ||
		strings.HasPrefix(lower, "# plan") ||
		strings.HasPrefix(lower, "## plan") ||
		strings.HasPrefix(lower, "### plan") ||
		strings.HasPrefix(lower, "thinking:") ||
		strings.HasPrefix(lower, "thoughts:") ||
		strings.HasPrefix(lower, "analysis:") {
		parts := strings.SplitN(trim, "\n\n", 2)
		if len(parts) == 1 {
			return parts[0], ""
		}
		return parts[0], parts[1]
	}
	return "", trim
}

func normalizePermissionResponse(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(v, "/") {
		return v
	}
	if v == "yes" || v == "no" || v == "always" {
		return "/" + v
	}
	return v
}

func readOnlyPrompt(original string) string {
	trim := strings.TrimSpace(original)
	if trim == "" {
		return "Respond only with READ <path> lines. No other text."
	}
	return "You requested file reads in prose. Respond only with READ <path> lines, no other text.\n\nOriginal request:\n" + trim
}

func mentionsReadInProse(s string) bool {
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "read") {
		return false
	}
	trim := strings.TrimSpace(lower)
	if strings.HasPrefix(trim, "read ") {
		return false
	}
	if strings.Contains(lower, "\nread ") {
		return false
	}
	return true
}

func handleRetry(m *tuiModel) tea.Cmd {
	if m.lastPrompt == "" {
		m.appendAction("No previous prompt to retry.")
		return nil
	}
	m.running = true
	m.err = nil
	m.appendAction("Retrying...")
	m.appendUser(m.lastPrompt)
	return startAgentStream(m, m.lastPrompt, m.lastAllowRead, m.allowWriteAll && !m.denyWriteAll, m.lastReadPaths)
}

func handleApplyCommand(m *tuiModel, cmd string) tea.Cmd {
	switch cmd {
	case "/apply":
		m.allowWriteAll = true
		m.denyWriteAll = false
		return applyPending(m, false)
	case "/apply-always":
		m.allowWriteAll = true
		m.denyWriteAll = false
		m.projectCfg.AllowWriteAlways = true
		m.projectCfg.DenyWriteAlways = false
		if err := saveProjectConfig(m); err != nil {
			m.appendAction("ERROR: " + err.Error())
		}
		return applyPending(m, true)
	case "/deny":
		m.pendingWrites = nil
		m.pendingDeletes = nil
		m.pendingPatches = nil
		m.pendingPrefrontal = ""
		m.allowWriteAll = false
		m.denyWriteAll = true
		m.appendAction("CHANGES DENIED (session)")
		return nil
	case "/deny-always":
		m.allowWriteAll = false
		m.denyWriteAll = true
		m.pendingWrites = nil
		m.pendingDeletes = nil
		m.pendingPatches = nil
		m.pendingPrefrontal = ""
		m.projectCfg.AllowWriteAlways = false
		m.projectCfg.DenyWriteAlways = true
		if err := saveProjectConfig(m); err != nil {
			m.appendAction("ERROR: " + err.Error())
		}
		m.appendAction("CHANGES DENIED (always)")
		return nil
	default:
		return nil
	}
}

func applyPending(m *tuiModel, always bool) tea.Cmd {
	if len(m.pendingWrites) == 0 && len(m.pendingDeletes) == 0 && len(m.pendingPatches) == 0 {
		m.appendAction("No pending changes.")
		return nil
	}
	root, err := os.Getwd()
	if err != nil {
		m.appendAction("ERROR: " + err.Error())
		return nil
	}
	appliedWrites := agent.ApplyWrites(root, m.pendingWrites)
	appliedDeletes := agent.ApplyDeletes(root, m.pendingDeletes)
	appliedPatches, failedPatches := agent.ApplyPatches(root, m.pendingPatches)
	if m.pendingPrefrontal != "" {
		agent.AppendPrefrontal(m.pendingPrefrontal, agent.FormatWritesSummary(appliedWrites))
		agent.AppendPrefrontal(m.pendingPrefrontal, agent.FormatDeletesSummary(appliedDeletes))
		agent.AppendPrefrontal(m.pendingPrefrontal, agent.FormatPatchesSummary(appliedPatches))
	}
	for _, w := range appliedWrites {
		m.appendAction("WRITE: " + w.Path)
	}
	for _, d := range appliedDeletes {
		m.appendAction("DELETE: " + d.Path)
	}
	for _, p := range appliedPatches {
		m.appendAction("PATCH: " + p.Path)
	}
	for _, p := range failedPatches {
		m.appendAction("PATCH FAILED: " + p.Path + " (" + p.Reason + ")")
	}
	if always {
		m.appendAction("CHANGES AUTO-APPLY ENABLED")
	}
	m.pendingWrites = nil
	m.pendingDeletes = nil
	m.pendingPatches = nil
	m.pendingPrefrontal = ""
	return nil
}

func helpLines() []string {
	return []string{
		"/help  Show commands",
		"/clear  Clear short-term memory",
		"/condense  Condense short-term memory to free space",
		"/retry  Retry last prompt",
		"/model  Show or set model",
		"/usage  Show memory and token usage",
		"/actions  Toggle action log",
		"/yes  Allow reading for session",
		"/no  Deny reading for session",
		"/always  Always allow reading",
		"/apply  Apply and allow writes for session",
		"/apply-always  Always apply writes/deletes",
		"/deny  Deny writes for session",
		"/deny-always  Always deny writes/deletes",
	}
}

func helpItems() []commandItem {
	return []commandItem{
		{cmd: "/help", desc: "Show commands"},
		{cmd: "/clear", desc: "Clear short-term memory"},
		{cmd: "/condense", desc: "Condense short-term memory to free space"},
		{cmd: "/retry", desc: "Retry last prompt"},
		{cmd: "/model", desc: "Show or set model"},
		{cmd: "/usage", desc: "Show memory and token usage"},
		{cmd: "/actions", desc: "Toggle action log"},
		{cmd: "/yes", desc: "Allow reading for session"},
		{cmd: "/no", desc: "Deny reading for session"},
		{cmd: "/always", desc: "Always allow reading"},
		{cmd: "/apply", desc: "Apply and allow writes for session"},
		{cmd: "/apply-always", desc: "Always apply writes/deletes"},
		{cmd: "/deny", desc: "Deny writes for session"},
		{cmd: "/deny-always", desc: "Always deny writes/deletes"},
	}
}

func commandSuggestions(prefix string) []commandItem {
	all := helpItems()
	if prefix == "/" {
		return all
	}
	if strings.HasPrefix(prefix, "/model") {
		return modelSuggestions()
	}
	var out []commandItem
	for _, item := range all {
		if strings.HasPrefix(item.cmd, prefix) {
			out = append(out, item)
		}
	}
	return out
}

func currentSuggestions(m tuiModel) []commandItem {
	val := strings.TrimSpace(m.input.Value())
	if strings.HasPrefix(val, "/") {
		return commandSuggestions(val)
	}
	return nil
}

func parseChoiceBlock(s string) *choiceBlock {
	lines := strings.Split(s, "\n")
	var question string
	var options []string
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(strings.ToUpper(line), "CHOICE:") {
			question = strings.TrimSpace(strings.TrimPrefix(line, "CHOICE:"))
			for j := i + 1; j < len(lines); j++ {
				l := strings.TrimSpace(lines[j])
				if l == "" {
					break
				}
				if strings.HasPrefix(l, "-") {
					options = append(options, strings.TrimSpace(strings.TrimPrefix(l, "-")))
				} else if len(l) >= 3 && l[1] == '.' {
					options = append(options, strings.TrimSpace(l[2:]))
				}
			}
			break
		}
	}
	if question == "" || len(options) == 0 {
		return nil
	}
	return &choiceBlock{question: question, options: options}
}

func choiceCount(m tuiModel) int {
	if len(m.history) == 0 {
		return 0
	}
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].kind == "choice" {
			return len(m.history[i].options)
		}
	}
	return 0
}

func applyChoice(m *tuiModel) tea.Cmd {
	var options []string
	for i := len(m.history) - 1; i >= 0; i-- {
		if m.history[i].kind == "choice" {
			options = m.history[i].options
			break
		}
	}
	if len(options) == 0 {
		m.choiceActive = false
		return nil
	}
	idx := clamp(m.choiceIndex, 0, len(options)-1)
	selected := options[idx]
	m.choiceActive = false

	switch m.choiceKind {
	case "read":
		cmd := strings.Fields(selected)[0]
		return submitPrompt(m, cmd)
	case "apply":
		cmd := strings.Fields(selected)[0]
		return submitPrompt(m, cmd)
	case "model":
		m.running = true
		m.appendUser(selected)
		m.lastPrompt = selected
		m.lastAllowRead = m.allowReadAll
		m.lastReadPaths = nil
		return startAgentStream(m, selected, m.allowReadAll, m.allowWriteAll && !m.denyWriteAll, nil)
	default:
		return nil
	}
}

func submitPrompt(m *tuiModel, prompt string) tea.Cmd {
	if m.pendingPrompt != "" {
		switch normalizePermissionResponse(prompt) {
		case "/yes":
			m.allowReadAll = true
			m.denyReadAll = false
			m.running = true
			p := m.pendingPrompt
			m.pendingPrompt = ""
			m.appendAction("READ APPROVED (session)")
			m.appendUser(p)
			if len(m.pendingReadPaths) > 0 {
				paths := m.pendingReadPaths
				m.pendingReadPaths = nil
				m.readRequestDepth = 0
				m.lastReadPaths = paths
				return startAgentStream(m, p, true, m.allowWriteAll && !m.denyWriteAll, paths)
			}
			return startAgentStream(m, p, true, m.allowWriteAll && !m.denyWriteAll, nil)
		case "/always":
			m.allowReadAll = true
			m.denyReadAll = false
			m.projectCfg.AllowReadAlways = true
			if err := saveProjectConfig(m); err != nil {
				m.appendAction("ERROR: " + err.Error())
			}
			m.running = true
			p := m.pendingPrompt
			m.pendingPrompt = ""
			m.appendAction("READ ALWAYS APPROVED")
			m.appendUser(p)
			if len(m.pendingReadPaths) > 0 {
				paths := m.pendingReadPaths
				m.pendingReadPaths = nil
				m.readRequestDepth = 0
				m.lastReadPaths = paths
				return startAgentStream(m, p, true, m.allowWriteAll && !m.denyWriteAll, paths)
			}
			return startAgentStream(m, p, true, m.allowWriteAll && !m.denyWriteAll, nil)
		case "/no":
			m.allowReadAll = false
			m.denyReadAll = true
			m.appendAction("READ DENIED (session)")
			m.pendingPrompt = ""
			m.pendingReadPaths = nil
			m.readRequestDepth = 0
			return nil
		default:
			m.appendPermission("READ REQUIRED. Choose an option:")
			m.appendChoice("read", "Choose:", []string{"/yes allow for session", "/no deny for session", "/always always allow"})
			return nil
		}
	}

	if strings.HasPrefix(prompt, "/") {
		cmd := strings.ToLower(strings.TrimSpace(prompt))
		if cmd == "/yes" || cmd == "/no" || cmd == "/always" {
			m.appendAction("No pending permission request.")
			return nil
		}
		if strings.HasPrefix(cmd, "/model") {
			fields := strings.Fields(prompt)
			if len(fields) < 2 {
				m.appendAction("MODEL: " + m.model)
				return nil
			}
			newModel := strings.TrimSpace(fields[1])
			if newModel == "" {
				m.appendAction("MODEL: " + m.model)
				return nil
			}
			cfg, err := userconfig.Load()
			if err != nil {
				cfg = userconfig.Config{}
			}
			cfg.Model = newModel
			if err := userconfig.Save(cfg); err != nil {
				m.appendAction("ERROR: " + err.Error())
				return nil
			}
			m.model = newModel
			m.appendAction("MODEL SET: " + newModel)
			return nil
		}
		if cmd == "/usage" {
			usage := usageFromConfig()
			m.appendAction("Usage:")
			m.appendAction("LTM bytes: " + formatBytes(usage.LtmBytes))
			m.appendAction("STM bytes: " + formatBytes(usage.StmBytes))
			m.appendAction("STM context bytes: " + formatBytes(usage.StmContextBytes))
			m.appendAction("Conversation bytes: " + formatBytes(usage.ConvBytes))
			m.appendAction("Conversation context bytes: " + formatBytes(usage.ConvContextBytes))
			m.appendAction("Approx tokens: " + fmt.Sprintf("%d/%d", usage.ApproxTokens, usage.BudgetTokens))
			return nil
		}
		if cmd == "/actions" {
			if m.showActions {
				m.appendAction("ACTIONS HIDDEN")
				m.showActions = false
			} else {
				m.showActions = true
				m.appendAction("ACTIONS SHOWN")
			}
			return nil
		}
		if cmd == "/clear" {
			m.input.SetValue("")
			m.history = nil
			m.viewport.SetContent("")
			m.lastPrompt = ""
			m.pendingPrompt = ""
			m.pendingReadPaths = nil
			m.pendingWrites = nil
			m.pendingDeletes = nil
			m.pendingPatches = nil
			m.pendingPrefrontal = ""
			m.readRequestDepth = 0
			m.choiceActive = false
			m.choiceKind = ""
			m.choiceIndex = 0
			m.running = true
			return runMemoryCmd(prompt)
		}
		if cmd == "/condense" {
			m.running = true
		}
		if cmd == "/retry" {
			return handleRetry(m)
		}
		if cmd == "/help" {
			m.appendAction("Commands:")
			for _, line := range helpLines() {
				m.appendAction(line)
			}
			return nil
		}
		if cmd == "/apply" || cmd == "/apply-always" || cmd == "/deny" || cmd == "/deny-always" {
			return handleApplyCommand(m, cmd)
		}
		return runMemoryCmd(prompt)
	}

	mentions := agent.ExtractFileMentions(prompt)
	if len(mentions) > 0 && !m.allowReadAll && !m.denyReadAll {
		m.pendingPrompt = prompt
		m.appendPermission("READ FILES? Choose an option:")
		m.appendChoice("read", "Choose:", []string{"/yes allow for session", "/no deny for session", "/always always allow"})
		return nil
	}
	m.running = true
	m.res = nil
	m.err = nil
	m.appendUser(prompt)
	m.lastPrompt = prompt
	m.lastAllowRead = m.allowReadAll
	m.readRequestDepth = 0
	m.lastReadPaths = nil
	return startAgentStream(m, prompt, m.allowReadAll, m.allowWriteAll && !m.denyWriteAll, nil)
}

func saveProjectConfig(m *tuiModel) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	return agent.SaveProjectConfig(root, m.projectCfg)
}

func newMarkdownRenderer(width int) *glamour.TermRenderer {
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("ascii"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	return r
}

func modelSuggestions() []commandItem {
	seen := map[string]struct{}{}
	add := func(model string, out *[]commandItem) {
		model = strings.TrimSpace(model)
		if model == "" {
			return
		}
		if _, ok := seen[model]; ok {
			return
		}
		seen[model] = struct{}{}
		*out = append(*out, commandItem{cmd: "/model " + model, desc: "Set model"})
	}

	var out []commandItem
	add(currentModel(), &out)
	if cfg, err := userconfig.Load(); err == nil {
		add(cfg.Model, &out)
	}
	add("gpt-4.1", &out)
	return out
}
