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

	b.WriteString("When modifying existing files, prefer PATCH with a unified diff for the smallest possible change.\n")
	b.WriteString("Only use EDIT (full-file rewrite) if a patch is impractical.\n")
	b.WriteString("Use WRITE only for new files (or full replacement if explicitly requested).\n\n")
	b.WriteString("WRITE <relative/path>\n")
	b.WriteString("```\n<content>\n```\n\n")
	b.WriteString("EDIT <relative/path>\n")
	b.WriteString("```\n<full new content>\n```\n\n")
	b.WriteString("To delete a file, use a single line:\n")
	b.WriteString("DELETE <relative/path>\n\n")
	b.WriteString("To apply a unified diff, use:\n")
	b.WriteString("PATCH <relative/path>\n")
	b.WriteString("```patch\n<unified diff>\n```\n\n")
	b.WriteString("If file contents were not provided due to permissions, ask the user to approve reading them.\n")
	b.WriteString("If you need more context, respond with one or more lines in this form (no extra text):\n")
	b.WriteString("READ <relative/path>\n\n")
	b.WriteString("Do not ask for file reads in prose; only use READ lines. Requests in prose will be ignored.\n")
	b.WriteString("Never assume file contents from filenames alone. If a change depends on file contents, request READ lines and stop.\n\n")

	b.WriteString("If you state that you will modify files, you must output the WRITE/EDIT/DELETE blocks in the same response.\n")
	b.WriteString("Otherwise, respond with your plan and any questions.\n")
	return b.String()
}
