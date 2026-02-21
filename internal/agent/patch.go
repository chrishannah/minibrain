package agent

import (
	"bufio"
	"regexp"
	"strings"
)

type hunk struct {
	oldStart int
	oldCount int
	newStart int
	newCount int
	lines    []string
}

var hunkHeader = regexp.MustCompile(`@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

func applyUnifiedPatch(original, patch string) (string, bool) {
	hadTrailingNewline := strings.HasSuffix(original, "\n")
	lines := splitLines(original)
	hunks := parseHunks(patch)
	if len(hunks) == 0 {
		return "", false
	}
	out := make([]string, 0, len(lines))
	idx := 0
	for _, h := range hunks {
		// Copy unchanged lines before hunk.
		start := h.oldStart - 1
		if start < 0 {
			start = 0
		}
		if start > len(lines) {
			return "", false
		}
		if start < idx {
			start = idx
		}
		out = append(out, lines[idx:start]...)
		idx = start

		for _, line := range h.lines {
			if line == "" {
				continue
			}
			switch line[0] {
			case ' ':
				content := line[1:]
				if idx >= len(lines) || lines[idx] != content {
					return "", false
				}
				out = append(out, content)
				idx++
			case '-':
				content := line[1:]
				if idx >= len(lines) || lines[idx] != content {
					return "", false
				}
				idx++
			case '+':
				out = append(out, line[1:])
			case '\\':
				// ignore no-newline marker
			default:
				return "", false
			}
		}
	}

	out = append(out, lines[idx:]...)
	result := strings.Join(out, "\n")
	if hadTrailingNewline {
		result += "\n"
	}
	return result, true
}

func parseHunks(patch string) []hunk {
	scanner := bufio.NewScanner(strings.NewReader(patch))
	var hunks []hunk
	var current *hunk
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "@@ ") {
			m := hunkHeader.FindStringSubmatch(line)
			if len(m) == 0 {
				continue
			}
			current = &hunk{
				oldStart: parseInt(m[1]),
				oldCount: parseIntDefault(m[2], 1),
				newStart: parseInt(m[3]),
				newCount: parseIntDefault(m[4], 1),
			}
			hunks = append(hunks, *current)
			continue
		}
		if current != nil {
			current.lines = append(current.lines, line)
			// Update the last hunk in slice.
			hunks[len(hunks)-1] = *current
		}
	}
	return hunks
}

func parseInt(s string) int {
	v := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		v = v*10 + int(r-'0')
	}
	return v
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	return parseInt(s)
}

func splitLines(s string) []string {
	trim := strings.TrimSuffix(s, "\n")
	if trim == "" {
		return []string{}
	}
	return strings.Split(trim, "\n")
}
