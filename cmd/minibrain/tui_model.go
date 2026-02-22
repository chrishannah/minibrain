package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/chrishannah/minibrain/internal/agent"
)

const (
	colorPrimary   = "252"
	colorSecondary = "244"
)

type runMsg struct {
	res agent.Result
	err error
}

type memMsg struct {
	action    string
	stats     agent.MemoryStats
	err       error
	condensed bool
}

type streamMsg struct {
	delta string
	done  bool
	res   agent.Result
	err   error
}

type historyEntry struct {
	text    string
	kind    string
	bold    bool
	options []string
}

type tuiModel struct {
	input             textinput.Model
	viewport          viewport.Model
	history           []historyEntry
	running           bool
	res               *agent.Result
	err               error
	width             int
	height            int
	defaultPrompt     bool
	stats             agent.MemoryStats
	usage             agent.UsageStats
	allowReadAll      bool
	allowWriteAll     bool
	denyWriteAll      bool
	pendingPrompt     string
	model             string
	status            string
	lastPrompt        string
	lastAllowRead     bool
	lastReadPaths     []string
	pendingWrites     []agent.WriteOp
	pendingDeletes    []agent.DeleteOp
	pendingPatches    []agent.PatchOp
	pendingPrefrontal string
	pendingReadPaths  []string
	readRequestDepth  int
	patchReadRerun    bool
	patchFormatRetry  bool
	patchWriteRetry   bool
	pendingPreviewed  bool
	thinkingActive    bool
	suggestIndex      int
	choiceActive      bool
	choiceKind        string
	choiceIndex       int
	denyReadAll       bool
	projectCfg        agent.ProjectConfig
	mdRenderer        *glamour.TermRenderer
	mdWidth           int
	streamCh          chan streamMsg
	showActions       bool
	showRaw           bool
}

func runTUI() {
	m := newTUIModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("TUI error:", err)
	}
}

func newTUIModel() tuiModel {
	ti := textinput.New()
	ti.Placeholder = "Hello"
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 80
	vp := viewport.New(80, 10)
	stats, _ := initialStats()
	usage, _ := initialUsage()
	root, _ := os.Getwd()
	perms := agent.ResolvePermissionState(root, readAllowedFromEnv(), writeAllowedFromEnv())
	m := tuiModel{
		input:         ti,
		viewport:      vp,
		history:       []historyEntry{},
		defaultPrompt: false,
		stats:         stats,
		usage:         usage,
		allowReadAll:  perms.AllowRead,
		allowWriteAll: perms.AllowWrite,
		denyWriteAll:  perms.DenyWrite,
		model:         currentModel(),
		choiceIndex:   0,
		projectCfg:    perms.Project,
		showActions:   true,
		showRaw:       false,
		status:        "Ready",
	}
	m.updateMarkdownRenderer()
	m.refreshViewport()
	return m
}

func (m tuiModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = clamp(m.width-4, 40, 120)
		m.viewport.Width = m.width
		m.viewport.Height = clamp(m.height-6, 6, 40)
		m.updateMarkdownRenderer()
		m.refreshViewport()
		return m, nil
	case tea.KeyMsg:
		if m.defaultPrompt {
			switch msg.Type {
			case tea.KeyRunes, tea.KeyBackspace, tea.KeyDelete:
				m.input.SetValue("")
				m.defaultPrompt = false
			}
		}

		if m.choiceActive {
			switch msg.Type {
			case tea.KeyUp:
				m.choiceIndex = clamp(m.choiceIndex-1, 0, choiceCount(m)-1)
				m.refreshViewport()
				return m, nil
			case tea.KeyDown:
				m.choiceIndex = clamp(m.choiceIndex+1, 0, choiceCount(m)-1)
				m.refreshViewport()
				return m, nil
			case tea.KeyEnter:
				return m, applyChoice(&m)
			}
			return m, nil
		}

		suggestions := currentSuggestions(m)
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			if m.running {
				return m, nil
			}
			if len(suggestions) > 0 && strings.HasPrefix(strings.TrimSpace(m.input.Value()), "/") {
				m.suggestIndex = clamp(m.suggestIndex-1, 0, len(suggestions)-1)
				return m, nil
			}
			if m.lastPrompt != "" {
				m.input.SetValue(m.lastPrompt)
				m.input.CursorEnd()
			}
			return m, nil
		case tea.KeyDown:
			if m.running {
				return m, nil
			}
			if len(suggestions) > 0 && strings.HasPrefix(strings.TrimSpace(m.input.Value()), "/") {
				m.suggestIndex = clamp(m.suggestIndex+1, 0, len(suggestions)-1)
				return m, nil
			}
		case tea.KeyEnter:
			if m.running {
				return m, nil
			}
			prompt := strings.TrimSpace(m.input.Value())
			if prompt == "" {
				return m, nil
			}
			if len(suggestions) > 0 && strings.HasPrefix(prompt, "/") {
				selected := suggestions[clamp(m.suggestIndex, 0, len(suggestions)-1)].cmd
				if prompt != selected {
					m.input.SetValue(selected)
					m.input.CursorEnd()
					return m, nil
				}
			}
			m.input.SetValue("")
			return m, submitPrompt(&m, prompt)
		}
	case runMsg:
		m.running = false
		m.clearThinking()
		m.status = "Ready"
		if msg.err != nil {
			m.err = msg.err
			m.appendAction(formatAction(ActionError, msg.err.Error()))
			m.appendAction("Type /retry to try again.")
			return m, nil
		}
		m.res = &msg.res
		m.appendRaw(msg.res.LLMOutput)
		readReq := msg.res.ReadRequests
		if len(readReq) > 0 && m.denyReadAll {
			m.appendAction(formatAction(ActionReadDenied, "session"))
			m.appendRunResult(msg.res)
			m.stats = msg.res.Memory
			return m, nil
		}
		if len(readReq) > 0 && !m.allowReadAll {
			m.pendingPrompt = m.lastPrompt
			m.pendingReadPaths = readReq
			m.status = "Ready"
			m.appendPermission("READ REQUEST: can I read files in this directory?")
			m.appendChoice("read", "Choose:", []string{"/yes allow for session", "/no deny for session", "/always always allow"})
			return m, nil
		}
		if len(readReq) > 0 && m.allowReadAll {
			if m.readRequestDepth >= 1 {
				// avoid repeated read loops
			} else {
				m.readRequestDepth++
				m.lastReadPaths = readReq
				m.running = true
				return m, startAgentStream(&m, m.lastPrompt, true, m.allowWriteAll && !m.denyWriteAll, readReq)
			}
		}

		if len(readReq) == 0 && len(msg.res.ProposedPatches) > 0 {
			var patchPaths []string
			invalidPatch := false
			for _, p := range msg.res.ProposedPatches {
				if strings.TrimSpace(p.Path) != "" {
					patchPaths = append(patchPaths, p.Path)
				}
				if !agent.HasValidHunks(p.Patch) {
					invalidPatch = true
				}
			}
			if invalidPatch {
				if !m.patchFormatRetry {
					m.patchFormatRetry = true
					m.running = true
					return m, startAgentStream(&m, patchFormatPrompt(m.lastPrompt), m.allowReadAll, m.allowWriteAll && !m.denyWriteAll, m.lastReadPaths)
				}
				m.appendAction(formatAction(ActionPatchFailed, "invalid diff format (missing @@ -a,b +c,d @@ hunks)"))
				m.appendAction(formatAction(ActionInfo, "Use /retry to try again"))
				m.stats = msg.res.Memory
				m.usage = usageFromConfig()
				return m, nil
			}
			if len(patchPaths) > 0 {
				if m.allowReadAll && !m.patchReadRerun {
					m.patchReadRerun = true
					m.appendAction("READ: " + strings.Join(patchPaths, ", "))
					m.lastReadPaths = patchPaths
					m.running = true
					return m, startAgentStream(&m, m.lastPrompt, true, m.allowWriteAll && !m.denyWriteAll, patchPaths)
				}
				if !m.allowReadAll && !m.denyReadAll {
					m.pendingPrompt = m.lastPrompt
					m.pendingReadPaths = patchPaths
					m.status = "Ready"
					m.appendAction(formatAction(ActionReadRequest, "files needed for patches"))
					m.appendPermission("READ FILES FOR PATCHES? Choose an option:")
					m.appendChoice("read", "Choose:", []string{"/yes allow for session", "/no deny for session", "/always always allow"})
					return m, nil
				}
			}
		}

		if choice := parseChoiceBlock(msg.res.LLMOutput); choice != nil {
			m.appendChoice("model", choice.question, choice.options)
			return m, nil
		}

		m.appendRunResult(msg.res)
		m.stats = msg.res.Memory
		m.usage = usageFromConfig()
		if len(msg.res.FailedPatches) > 0 && !m.patchWriteRetry {
			var retryPaths []string
			for _, p := range msg.res.FailedPatches {
				if strings.TrimSpace(p.Path) != "" {
					retryPaths = append(retryPaths, p.Path)
				}
			}
			if len(retryPaths) > 0 {
				if !m.allowReadAll && !m.denyReadAll {
					m.pendingPrompt = m.lastPrompt
					m.pendingReadPaths = retryPaths
					m.status = "Ready"
					m.appendPermission("READ FILES FOR FULL REWRITE? Choose an option:")
					m.appendChoice("read", "Choose:", []string{"/yes allow for session", "/no deny for session", "/always always allow"})
					return m, nil
				}
				m.patchWriteRetry = true
				m.running = true
				return m, startAgentStream(&m, patchRewritePrompt(m.lastPrompt, retryPaths), m.allowReadAll, m.allowWriteAll && !m.denyWriteAll, retryPaths)
			}
		}
		if !msg.res.Applied && (len(msg.res.ProposedWrites) > 0 || len(msg.res.ProposedDeletes) > 0 || len(msg.res.ProposedPatches) > 0) {
			m.pendingWrites = msg.res.ProposedWrites
			m.pendingDeletes = msg.res.ProposedDeletes
			m.pendingPatches = msg.res.ProposedPatches
			m.pendingPrefrontal = msg.res.PrefrontalPath
			if !m.pendingPreviewed {
				appendChangePreview(&m)
				m.pendingPreviewed = true
			}
			if m.denyWriteAll {
				m.appendAction("CHANGES DENIED (always)")
			} else if m.allowWriteAll {
				return m, applyPending(&m, true)
			} else {
				m.status = "Ready"
				m.appendPermission("APPLY CHANGES? Choose an option:")
				m.appendChoice("apply", "Choose:", []string{"/apply allow for session", "/apply-always always apply", "/deny deny for session", "/deny-always always deny"})
			}
		}
		return m, nil
	case streamMsg:
		if msg.err != nil {
			m.running = false
			m.err = msg.err
			m.clearThinking()
			m.appendAction(formatAction(ActionError, msg.err.Error()))
			m.appendAction(formatAction(ActionInfo, "Type /retry to try again"))
			m.status = "Error"
			return m, nil
		}
		if msg.done {
			m.running = false
			msg2 := runMsg{res: msg.res, err: nil}
			return m, func() tea.Msg { return msg2 }
		}
		return m, listenStream(m.streamCh)
	case memMsg:
		m.running = false
		if msg.err != nil {
			m.err = msg.err
			m.appendAction(formatAction(ActionError, msg.err.Error()))
			return m, nil
		}
		m.stats = msg.stats
		m.usage = usageFromConfig()
		if msg.action != "" {
			m.appendAction(msg.action)
		} else if msg.condensed {
			m.appendAction(formatAction(ActionMemory, "CONDENSED"))
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.viewport, _ = m.viewport.Update(msg)
	return m, cmd
}
