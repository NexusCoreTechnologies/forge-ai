package optimizer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	contextmodels "forgeai/backend/internal/context/models"
)

const (
	MaxContextTokens = 12000
	MaxFileLines     = 500
)

var ignoredDirNames = map[string]bool{
	"generated":    true,
	"logs":         true,
	"reports":      true,
	"cache":        true,
	"node_modules": true,
	".git":         true,
}

type OptimizedContext struct {
	Prompt        string          `json:"prompt"`
	ExecutionPlan string          `json:"executionPlan"`
	CurrentSprint string          `json:"currentSprint,omitempty"`
	Architecture  string          `json:"architectureSummary,omitempty"`
	Files         []OptimizedFile `json:"files,omitempty"`
	Readme        string          `json:"readme,omitempty"`
	ADRs          []string        `json:"adrs,omitempty"`
}

type OptimizedFile struct {
	Path      string `json:"path"`
	Content   string `json:"content,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Truncated bool   `json:"truncated"`
	Lines     int    `json:"lines,omitempty"`
}

type Metrics struct {
	OriginalTokens   int     `json:"originalTokens"`
	OptimizedTokens  int     `json:"optimizedTokens"`
	CompressionRatio float64 `json:"compressionRatio"`
}

func Optimize(workspace, promptContent, planContent string, ctx *contextmodels.ExecutionContext) (*OptimizedContext, *Metrics, error) {
	opt := &OptimizedContext{
		Prompt:        promptContent,
		ExecutionPlan: planContent,
	}
	if ctx != nil {
		opt.CurrentSprint = ctx.CurrentSprint
		opt.Architecture = ctx.Architecture
	}

	var sections []struct {
		label   string
		content string
		file    *OptimizedFile
	}
	if opt.CurrentSprint != "" {
		sections = append(sections, struct {
			label   string
			content string
			file    *OptimizedFile
		}{label: "currentSprint", content: opt.CurrentSprint})
	}
	if opt.Architecture != "" {
		sections = append(sections, struct {
			label   string
			content string
			file    *OptimizedFile
		}{label: "architecture", content: opt.Architecture})
	}

	selectedFiles := []OptimizedFile{}
	if ctx != nil && len(ctx.RelevantFiles) > 0 {
		keywords := collectKeywords(promptContent, planContent, opt.CurrentSprint, opt.Architecture)
		for _, rel := range ctx.RelevantFiles {
			absPath := rel
			if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(workspace, rel)
			}
			if shouldSkip(absPath) || shouldExcludeFile(absPath) {
				continue
			}
			content := ""
			if ctx.Documentation != nil {
				if doc, ok := ctx.Documentation[absPath]; ok {
					content = doc
				} else if doc, ok := ctx.Documentation[rel]; ok {
					content = doc
				}
			}
			if content == "" {
				fileContent, err := os.ReadFile(absPath)
				if err != nil {
					continue
				}
				content = string(fileContent)
			}
			if !matchesKeywords(content, keywords) && !matchesKeywords(absPath, keywords) {
				continue
			}
			lines := countLines(content)
			if lines > MaxFileLines {
				selectedFiles = append(selectedFiles, OptimizedFile{Path: absPath, Summary: summarizeText(content, 200), Truncated: true, Lines: lines})
				continue
			}
			selectedFiles = append(selectedFiles, OptimizedFile{Path: absPath, Content: content, Truncated: false, Lines: lines})
		}
	}
	opt.Files = selectedFiles

	readmePath, err := findReadme(workspace)
	if err == nil && readmePath != "" {
		content, err := os.ReadFile(readmePath)
		if err == nil {
			opt.Readme = string(content)
		}
	}
	adrs, err := findADRs(workspace)
	if err == nil {
		opt.ADRs = adrs
	}

	originalContent := strings.Join([]string{promptContent, planContent, opt.CurrentSprint, opt.Architecture, opt.Readme, strings.Join(opt.ADRs, "\n")}, "\n")
	for _, file := range opt.Files {
		if file.Content != "" {
			originalContent += "\n" + file.Content
		} else if file.Summary != "" {
			originalContent += "\n" + file.Summary
		}
	}

	originalTokens := estimateTokens(originalContent)
	optimizedTokens := 0
	optimizedText := []string{}

	addSection := func(label, content string) {
		if content == "" {
			return
		}
		if estimateTokens(content) > MaxContextTokens {
			content = summarizeText(content, 600)
		}
		if estimateTokens(content) > MaxContextTokens {
			content = summarizeText(content, 200)
		}
		if estimateTokens(content) > MaxContextTokens {
			return
		}
		optimizedText = append(optimizedText, fmt.Sprintf("%s:\n%s", label, content))
	}

	addSection("prompt", promptContent)
	addSection("executionPlan", planContent)
	addSection("currentSprint", opt.CurrentSprint)
	addSection("architecture", opt.Architecture)
	for _, file := range opt.Files {
		if file.Content != "" {
			addSection(filepath.Base(file.Path), file.Content)
		} else if file.Summary != "" {
			addSection(filepath.Base(file.Path), file.Summary)
		}
	}
	if opt.Readme != "" {
		addSection("readme", opt.Readme)
	}
	if len(opt.ADRs) > 0 {
		addSection("adrs", strings.Join(opt.ADRs, "\n"))
	}

	optimizedTextJoined := strings.Join(optimizedText, "\n\n")
	optimizedTokens = estimateTokens(optimizedTextJoined)
	if optimizedTokens > MaxContextTokens {
		optimizedTextJoined = summarizeText(optimizedTextJoined, 8000)
		optimizedTokens = estimateTokens(optimizedTextJoined)
	}

	metric := &Metrics{OriginalTokens: originalTokens, OptimizedTokens: optimizedTokens}
	if originalTokens > 0 {
		metric.CompressionRatio = float64(originalTokens) / float64(optimizedTokens)
		if metric.CompressionRatio == 0 {
			metric.CompressionRatio = 1
		}
	} else {
		metric.CompressionRatio = 1
	}
	if metric.OptimizedTokens == 0 {
		metric.OptimizedTokens = estimateTokens(optimizedTextJoined)
	}
	if metric.OptimizedTokens == 0 {
		metric.OptimizedTokens = 1
	}

	opt.Prompt = optimizedTextJoined
	return opt, metric, nil
}

func summarizeText(text string, maxTokens int) string {
	if maxTokens <= 0 {
		return ""
	}
	words := strings.Fields(text)
	if len(words) <= maxTokens {
		return text
	}
	sample := words[:maxTokens]
	return strings.Join(sample, " ") + "..."
}

func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	return len(strings.Fields(text)) + len(text)/4
}

func countLines(text string) int {
	if text == "" {
		return 0
	}
	return strings.Count(text, "\n") + 1
}

func shouldSkip(path string) bool {
	absPath := path
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Clean(absPath)
	}
	segments := strings.Split(absPath, string(filepath.Separator))
	for _, segment := range segments {
		if ignoredDirNames[segment] {
			return true
		}
	}
	return false
}

func shouldExcludeFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if base == "execution_context.json" || base == "execution_plan.json" || base == "optimized_context.json" || base == "response.md" || strings.HasPrefix(base, "execution_result_") || strings.HasSuffix(base, ".execution.md") {
		return true
	}
	return false
}

func collectKeywords(parts ...string) []string {
	seen := map[string]struct{}{}
	keywords := []string{}
	for _, part := range parts {
		for _, word := range strings.Fields(strings.ToLower(part)) {
			cleaned := strings.Trim(word, `"'(),.-:/`)
			if len(cleaned) < 3 {
				continue
			}
			if _, exists := seen[cleaned]; exists {
				continue
			}
			seen[cleaned] = struct{}{}
			keywords = append(keywords, cleaned)
		}
	}
	return keywords
}

func matchesKeywords(text string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	lowerText := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(lowerText, keyword) {
			return true
		}
	}
	return false
}

func findReadme(root string) (string, error) {
	candidates := []string{"README.md", "readme.md", "README.MD", "readme.MD"}
	for _, name := range candidates {
		path := filepath.Join(root, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("readme not found")
}

func findADRs(root string) ([]string, error) {
	var adrs []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if shouldSkip(path) {
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToLower(filepath.Base(path))
		if strings.HasPrefix(name, "adr") || strings.Contains(name, "adr") {
			data, readErr := os.ReadFile(path)
			if readErr == nil {
				adrs = append(adrs, fmt.Sprintf("%s:\n%s", path, summarizeText(string(data), 120)))
			}
		}
		return nil
	})
	sort.Strings(adrs)
	return adrs, err
}
