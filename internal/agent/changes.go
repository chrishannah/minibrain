package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ParseWriteBlocks(s string) []WriteOp {
	var writes []WriteOp
	lines := strings.Split(s, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "WRITE ") && !strings.HasPrefix(line, "EDIT ") {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "WRITE "), "EDIT "))
		if path == "" {
			continue
		}
		if i+1 >= len(lines) || !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "```") {
			continue
		}
		i += 2
		var content []string
		for ; i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				break
			}
			content = append(content, lines[i])
		}
		writes = append(writes, WriteOp{Path: path, Content: strings.Join(content, "\n")})
	}
	return writes
}

func ParseDeleteLines(s string) []DeleteOp {
	var deletes []DeleteOp
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "DELETE ") {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(line, "DELETE "))
		if path == "" {
			continue
		}
		deletes = append(deletes, DeleteOp{Path: path})
	}
	return deletes
}

func ParseReadLines(s string) []string {
	var reads []string
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if !strings.HasPrefix(line, "READ ") {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(line, "READ "))
		if path == "" {
			continue
		}
		reads = append(reads, path)
	}
	return reads
}

func ParsePatchBlocks(s string) []PatchOp {
	var patches []PatchOp
	lines := strings.Split(s, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "PATCH ") {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(line, "PATCH "))
		if path == "" {
			continue
		}
		if i+1 >= len(lines) {
			continue
		}
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		if j >= len(lines) {
			continue
		}
		next := strings.TrimSpace(lines[j])
		// Prefer fenced patch blocks.
		if strings.HasPrefix(next, "```") {
			i = j + 1
			var content []string
			for ; i < len(lines); i++ {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
					break
				}
				content = append(content, lines[i])
			}
			patches = append(patches, PatchOp{Path: path, Patch: strings.Join(content, "\n")})
			continue
		}
		// Fallback: accept raw unified diff without fences.
		if strings.HasPrefix(next, "@@") || strings.HasPrefix(next, "--- ") || strings.HasPrefix(next, "+++ ") {
			i = j
			var content []string
			for ; i < len(lines); i++ {
				trimmed := strings.TrimSpace(lines[i])
				if strings.HasPrefix(trimmed, "PATCH ") || strings.HasPrefix(trimmed, "WRITE ") || strings.HasPrefix(trimmed, "EDIT ") || strings.HasPrefix(trimmed, "DELETE ") || strings.HasPrefix(trimmed, "READ ") {
					i--
					break
				}
				content = append(content, lines[i])
			}
			patches = append(patches, PatchOp{Path: path, Patch: strings.Join(content, "\n")})
		}
	}
	return patches
}

func ApplyWrites(root string, writes []WriteOp) []WriteOp {
	var applied []WriteOp
	for _, w := range writes {
		clean, err := safeRelPath(w.Path)
		if err != nil {
			continue
		}
		p := filepath.Join(root, clean)
		if err := ensureDir(filepath.Dir(p)); err != nil {
			continue
		}
		if err := os.WriteFile(p, []byte(w.Content), 0644); err != nil {
			continue
		}
		applied = append(applied, WriteOp{Path: clean, Content: w.Content})
	}
	return applied
}

func ApplyDeletes(root string, deletes []DeleteOp) []DeleteOp {
	var applied []DeleteOp
	for _, d := range deletes {
		clean, err := safeRelPath(d.Path)
		if err != nil {
			continue
		}
		p := filepath.Join(root, clean)
		if err := os.Remove(p); err != nil {
			continue
		}
		applied = append(applied, DeleteOp{Path: clean})
	}
	return applied
}

type PatchFailure struct {
	Path   string
	Reason string
}

func ApplyPatches(root string, patches []PatchOp) ([]PatchOp, []PatchFailure) {
	var applied []PatchOp
	var failed []PatchFailure
	for _, p := range patches {
		clean, err := safeRelPath(p.Path)
		if err != nil {
			failed = append(failed, PatchFailure{Path: p.Path, Reason: "invalid path"})
			continue
		}
		abs := filepath.Join(root, clean)
		b, err := os.ReadFile(abs)
		if err != nil {
			failed = append(failed, PatchFailure{Path: clean, Reason: "read failed: " + err.Error()})
			continue
		}
		updated, ok := applyUnifiedPatch(string(b), p.Patch)
		if !ok {
			failed = append(failed, PatchFailure{Path: clean, Reason: "patch failed to apply"})
			continue
		}
		if err := os.WriteFile(abs, []byte(updated), 0644); err != nil {
			failed = append(failed, PatchFailure{Path: clean, Reason: "write failed: " + err.Error()})
			continue
		}
		applied = append(applied, PatchOp{Path: clean, Patch: p.Patch})
	}
	return applied, failed
}

func FormatWritesSummary(writes []WriteOp) string {
	return FormatWritesSummaryWithTitle("Writes", writes)
}

func FormatDeletesSummary(deletes []DeleteOp) string {
	return FormatDeletesSummaryWithTitle("Deletes", deletes)
}

func FormatPatchesSummary(patches []PatchOp) string {
	return FormatPatchesSummaryWithTitle("Patches", patches)
}

func FormatWritesSummaryWithTitle(title string, writes []WriteOp) string {
	if len(writes) == 0 {
		return "\n## " + title + "\n(none)\n"
	}
	var b strings.Builder
	b.WriteString("\n## " + title + "\n")
	for _, w := range writes {
		b.WriteString("- " + w.Path + " (" + fmt.Sprintf("%d", len(w.Content)) + " bytes)\n")
	}
	return b.String()
}

func FormatDeletesSummaryWithTitle(title string, deletes []DeleteOp) string {
	if len(deletes) == 0 {
		return "\n## " + title + "\n(none)\n"
	}
	var b strings.Builder
	b.WriteString("\n## " + title + "\n")
	for _, d := range deletes {
		b.WriteString("- " + d.Path + "\n")
	}
	return b.String()
}

func FormatPatchesSummaryWithTitle(title string, patches []PatchOp) string {
	if len(patches) == 0 {
		return "\n## " + title + "\n(none)\n"
	}
	var b strings.Builder
	b.WriteString("\n## " + title + "\n")
	for _, p := range patches {
		b.WriteString("- " + p.Path + "\n")
	}
	return b.String()
}
