package agent

import "strings"

func BuildDeveloperMessage(agentConfig, soul, neo, stmContext, convContext, prompt string, refs []FileRef, fileList []string, listTruncated bool) string {
	var b strings.Builder
	b.WriteString("You are minibrain, a minimal agentic loop runner.\n")
	b.WriteString("Stay concise and explicit.\n\n")

	b.WriteString("Core config (MINIBRAIN.md):\n")
	if strings.TrimSpace(agentConfig) == "" {
		b.WriteString("(empty)\n\n")
	} else {
		b.WriteString(agentConfig + "\n\n")
	}

	b.WriteString("Personality (SOUL.md):\n")
	if strings.TrimSpace(soul) == "" {
		b.WriteString("(empty)\n\n")
	} else {
		b.WriteString(soul + "\n\n")
	}

	b.WriteString("Long-term memory (cortex/NEO.md):\n")
	if strings.TrimSpace(neo) == "" {
		b.WriteString("(empty)\n\n")
	} else {
		b.WriteString(neo + "\n\n")
	}

	b.WriteString("Short-term memory context (recent PREFRONTAL.md):\n")
	if strings.TrimSpace(stmContext) == "" {
		b.WriteString("(empty)\n\n")
	} else {
		b.WriteString(stmContext + "\n\n")
	}

	b.WriteString("Recent conversation summary (cortex/CONTEXT.md):\n")
	if strings.TrimSpace(convContext) == "" {
		b.WriteString("(empty)\n\n")
	} else {
		b.WriteString(convContext + "\n\n")
	}

	b.WriteString("User prompt:\n" + prompt + "\n\n")

	b.WriteString("Relevant repository files (shortlist, relative paths):\n")
	if len(fileList) == 0 {
		b.WriteString("(none)\n\n")
	} else {
		for _, f := range fileList {
			b.WriteString("- " + f + "\n")
		}
		if listTruncated {
			b.WriteString("- ... (truncated)\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("Mentioned files (contents provided below):\n")
	if len(refs) == 0 {
		b.WriteString("(none)\n\n")
	} else {
		for _, r := range refs {
			if r.Err != nil {
				b.WriteString("- " + formatMentionPath(r) + ": " + r.Err.Error() + "\n")
				continue
			}
			b.WriteString("- " + formatMentionPath(r) + "\n")
		}
		b.WriteString("\n")
		for _, r := range refs {
			if r.Err != nil {
				continue
			}
			b.WriteString("### " + r.Path + "\n")
			b.WriteString(r.Content + "\n\n")
		}
	}

	b.WriteString("You must respond ONLY with JSON matching the provided schema. No extra text.\n")
	b.WriteString("Use these fields:\n")
	b.WriteString("- read: list of file paths you need to read\n")
	b.WriteString("- patches: list of {path, diff} with unified diffs including @@ -a,b +c,d @@ hunks\n")
	b.WriteString("- writes: list of {path, content} for full-file rewrites or new files\n")
	b.WriteString("- deletes: list of paths to delete\n")
	b.WriteString("- message: short user-facing summary\n\n")
	b.WriteString("If you need file contents, populate read[] and leave patches/writes/deletes empty.\n")
	b.WriteString("Never assume file contents from filenames alone.\n")
	b.WriteString("Prefer patches for edits, writes for full replacements.\n")
	return b.String()
}
