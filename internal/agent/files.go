package agent

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ListFiles(root string, maxFiles int) ([]string, bool) {
	var files []string
	truncated := false
	stopErr := errors.New("stop walk")

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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
		files = append(files, rel)
		if maxFiles > 0 && len(files) >= maxFiles {
			truncated = true
			return stopErr
		}
		return nil
	})

	if err != nil && !errors.Is(err, stopErr) {
		return files, truncated
	}
	return files, truncated
}

type scoredPath struct {
	path  string
	score int
}

func ListRelevantFiles(root, prompt string, maxFiles int) ([]string, bool) {
	tokens := promptTokens(prompt)
	if len(tokens) == 0 {
		return ListFiles(root, maxFiles)
	}

	var scored []scoredPath
	truncated := false
	stopErr := errors.New("stop walk")

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
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
		score := scorePath(rel, tokens)
		if score > 0 {
			scored = append(scored, scoredPath{path: rel, score: score})
		}
		if maxFiles > 0 && len(scored) >= maxFiles*10 {
			truncated = true
			return stopErr
		}
		return nil
	})

	if err != nil && !errors.Is(err, stopErr) {
		return nil, truncated
	}

	if len(scored) == 0 {
		return ListFiles(root, maxFiles)
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].path < scored[j].path
		}
		return scored[i].score > scored[j].score
	})

	limit := len(scored)
	if maxFiles > 0 && limit > maxFiles {
		limit = maxFiles
		truncated = true
	}

	files := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		files = append(files, scored[i].path)
	}
	return files, truncated
}

func promptTokens(prompt string) []string {
	parts := strings.FieldsFunc(strings.ToLower(prompt), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.')
	})
	seen := map[string]struct{}{}
	var out []string
	for _, p := range parts {
		if len(p) < 3 {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func scorePath(path string, tokens []string) int {
	best := 0
	for _, t := range tokens {
		if t == "" {
			continue
		}
		if s := fuzzyScore(path, t); s > best {
			best = s
		}
	}
	return best
}
