package agent

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ExtractFileMentions(prompt string) []string {
	re := regexp.MustCompile(`@([A-Za-z0-9._/\-]+)`)
	matches := re.FindAllStringSubmatch(prompt, -1)
	seen := map[string]struct{}{}
	var out []string
	for _, m := range matches {
		p := m[1]
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func LoadMentionedFiles(root string, mentions []string, allowRead bool, maxFileBytes, maxTotalBytes int) []FileRef {
	var refs []FileRef
	total := 0
	for _, m := range mentions {
		resolved, ok := resolveMention(root, m)
		if !ok {
			refs = append(refs, FileRef{Mention: m, Path: m, Err: errors.New("not found")})
			continue
		}
		if !allowRead {
			refs = append(refs, FileRef{Mention: m, Path: resolved, Err: errors.New("permission denied: reading file content requires approval")})
			continue
		}
		clean, err := safeRelPath(resolved)
		if err != nil {
			refs = append(refs, FileRef{Mention: m, Path: resolved, Err: err})
			continue
		}
		p := filepath.Join(root, clean)
		info, err := os.Stat(p)
		if err == nil && maxFileBytes > 0 && info.Size() > int64(maxFileBytes) {
			refs = append(refs, FileRef{Mention: m, Path: clean, Err: errors.New("file too large")})
			continue
		}
		if maxTotalBytes > 0 && total >= maxTotalBytes {
			refs = append(refs, FileRef{Mention: m, Path: clean, Err: errors.New("total read limit exceeded")})
			continue
		}
		b, err := os.ReadFile(p)
		if err != nil {
			refs = append(refs, FileRef{Mention: m, Path: clean, Err: err})
			continue
		}
		if isBinary(b) {
			refs = append(refs, FileRef{Mention: m, Path: clean, Err: errors.New("binary file skipped")})
			continue
		}
		if maxTotalBytes > 0 && total+len(b) > maxTotalBytes {
			refs = append(refs, FileRef{Mention: m, Path: clean, Err: errors.New("total read limit exceeded")})
			continue
		}
		total += len(b)
		refs = append(refs, FileRef{Mention: m, Path: clean, Content: string(b)})
	}
	return refs
}

func MergeFileRefs(a, b []FileRef) []FileRef {
	seen := map[string]struct{}{}
	var out []FileRef
	for _, r := range a {
		key := r.Path
		if key == "" {
			key = r.Mention
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	for _, r := range b {
		key := r.Path
		if key == "" {
			key = r.Mention
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}

func safeRelPath(p string) (string, error) {
	p = filepath.Clean(p)
	if filepath.IsAbs(p) {
		return "", errors.New("absolute paths not allowed")
	}
	if strings.HasPrefix(p, "..") || strings.Contains(p, ".."+string(filepath.Separator)) {
		return "", errors.New("path traversal not allowed")
	}
	return p, nil
}

func resolveMention(root, mention string) (string, bool) {
	mention = strings.TrimSpace(mention)
	if mention == "" {
		return "", false
	}

	clean := filepath.Clean(mention)
	if filepath.IsAbs(clean) {
		return "", false
	}

	exactPath := filepath.Join(root, clean)
	if _, err := os.Stat(exactPath); err == nil {
		return clean, true
	}

	var bestPath string
	bestScore := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "dist" || name == "build" || name == "bin" || name == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		score := fuzzyScore(rel, mention)
		if score > bestScore {
			bestScore = score
			bestPath = rel
		}
		return nil
	})

	if bestScore < 300 {
		return "", false
	}
	return bestPath, true
}

func fuzzyScore(path, mention string) int {
	lp := strings.ToLower(path)
	lm := strings.ToLower(mention)
	base := strings.ToLower(filepath.Base(path))

	if lp == lm {
		return 1000
	}
	if base == lm {
		return 900
	}

	score := 0
	if strings.Contains(base, lm) {
		score = 700 - (len(base) - len(lm))
	}
	if strings.Contains(lp, lm) {
		c := 600 - (len(lp) - len(lm))
		if c > score {
			score = c
		}
	}

	dist := levenshtein(base, lm)
	c := 500 - dist
	if c > score {
		score = c
	}
	return score
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			cur[j] = min3(
				cur[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}
		prev = cur
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func isBinary(b []byte) bool {
	n := len(b)
	if n > 8000 {
		n = 8000
	}
	for i := 0; i < n; i++ {
		if b[i] == 0 {
			return true
		}
	}
	return false
}
