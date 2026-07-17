package files

import (
	"context"
	"fmt"
	"strings"
	"time"

	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
)

const fileAISearchMaxPathRunes = 240

func buildFileAISearchSystemPrompt(responseLanguage string) string {
	lines := []string{
		"You are a file browser assistant for a server panel.",
		"Answer using Markdown (headings, bullet lists).",
		"Only mention files and directories that appear in the provided inventory. Do not invent paths.",
		"If a \"Content line matches\" section is present, you may reference those lines and numbers only; never invent line numbers or snippets that are not listed there.",
		"If the inventory was truncated or incomplete, say so and suggest narrowing the directory or increasing limits.",
		"Group or rank results by relevance to the user's question when helpful.",
	}
	if lang := normalizeFileAISearchLanguage(responseLanguage); lang != "" {
		lines = append(lines, "Respond in "+lang+".")
	} else {
		lines = append(lines, "Prefer responding in the same language as the user's query (e.g. Chinese if the query is in Chinese).")
	}
	return strings.Join(lines, "\n")
}

func buildFileAISearchUserPrompt(root, query, responseLanguage string, items []AISearchInventoryItem, truncated, preFiltered bool, contentHits []FileAIContentHit, contentScannedFiles int, contentHitsTruncated bool, matchDesc string, promptHitMaxBytes int) string {
	var b strings.Builder
	b.WriteString("User question:\n")
	b.WriteString(strings.TrimSpace(query))
	b.WriteString("\n\nRoot directory:\n")
	b.WriteString(root)
	if lang := normalizeFileAISearchLanguage(responseLanguage); lang != "" {
		b.WriteString("\n\nPanel reply language:\n")
		b.WriteString(lang)
	}
	b.WriteString("\n\nInventory notes:\n")
	if truncated {
		b.WriteString("- Listing was truncated; not all files under the root were included.\n")
	} else {
		b.WriteString("- Full listing within limits.\n")
	}
	if preFiltered {
		b.WriteString("- Inventory was pre-filtered by keyword match on relative paths.\n")
	}
	b.WriteString("\nInventory (relative_path | dir | size_bytes | mtime):\n")
	for _, it := range items {
		rel := strings.TrimSpace(it.RelPath)
		if rel == "" {
			continue
		}
		runes := []rune(rel)
		if len(runes) > fileAISearchMaxPathRunes {
			rel = string(runes[:fileAISearchMaxPathRunes]) + "…"
		}
		dirMark := "f"
		if it.IsDir {
			dirMark = "d"
		}
		b.WriteString(fmt.Sprintf("%s | %s | %d | %s\n", rel, dirMark, it.Size, it.ModTime))
	}

	if promptHitMaxBytes <= 0 {
		promptHitMaxBytes = fileAIDefaultPromptHitBytes
	}
	if promptHitMaxBytes > fileAIHardMaxPromptHitBytes {
		promptHitMaxBytes = fileAIHardMaxPromptHitBytes
	}
	b.WriteString("\nContent line search (" + matchDesc + ", scanned plain-text-sized files only):\n")
	if contentScannedFiles > 0 {
		b.WriteString(fmt.Sprintf("- Files scanned for line matches: %d\n", contentScannedFiles))
	} else {
		b.WriteString("- No eligible files were content-scanned (limits, binary extensions, filters, or only directories in inventory).\n")
	}
	if contentHitsTruncated {
		b.WriteString("- Scan or hit list was truncated by safety limits; not all files or matches are shown.\n")
	}
	if len(contentHits) > 0 {
		b.WriteString(FormatContentHitsForPrompt(contentHits, promptHitMaxBytes, matchDesc))
	} else if contentScannedFiles > 0 {
		b.WriteString("- No lines matched the search in scanned files. The summary may still use path/metadata matches.\n")
	}

	b.WriteString("\nOutput requirement:\nSummarize what matches the user's intent and list the most relevant paths. Use Markdown.\n")
	return b.String()
}

func RunFileAISearchLLM(ctx context.Context, cfg terminalai.GeneratorConfig, clientTimeout time.Duration, root, query, responseLanguage string, items []AISearchInventoryItem, truncated, preFiltered bool, contentHits []FileAIContentHit, contentScannedFiles int, contentHitsTruncated bool, matchDesc string, promptHitMaxBytes, llmMaxOutputTokens int) (string, terminalai.ResponseUsage, error) {
	timeout := clientTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	client, err := terminalai.NewClient(terminalai.ClientConfig{
		Provider:  cfg.Provider,
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		APIType:   cfg.APIType,
		AuthMode:  cfg.AuthMode,
		MaxTokens: cfg.MaxTokens,
		Timeout:   timeout,
	})
	if err != nil {
		return "", terminalai.ResponseUsage{}, err
	}
	outTokens := cfg.MaxTokens
	if llmMaxOutputTokens > 0 {
		outTokens = llmMaxOutputTokens
	}
	if outTokens < 512 {
		outTokens = 2048
	}
	if outTokens > 4096 {
		outTokens = 4096
	}
	resp, err := client.ChatCompletion(ctx, terminalai.ChatCompletionRequest{
		Messages: []terminalai.ChatMessage{
			{Role: "system", Content: buildFileAISearchSystemPrompt(responseLanguage)},
			{Role: "user", Content: buildFileAISearchUserPrompt(root, query, responseLanguage, items, truncated, preFiltered, contentHits, contentScannedFiles, contentHitsTruncated, matchDesc, promptHitMaxBytes)},
		},
		MaxTokens: outTokens,
	})
	if err != nil {
		return "", terminalai.ResponseUsage{}, err
	}
	summary := strings.TrimSpace(resp.Content)
	if summary == "" {
		return "", resp.Usage, fmt.Errorf("model returned empty summary")
	}
	return summary, resp.Usage, nil
}

func normalizeFileAISearchLanguage(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	switch {
	case lang == "", lang == "*":
		return "English"
	case strings.HasPrefix(lang, "zh-hant"), strings.HasPrefix(lang, "zh-tw"), strings.HasPrefix(lang, "zh-hk"):
		return "Traditional Chinese"
	case strings.HasPrefix(lang, "zh"):
		return "Simplified Chinese"
	case strings.HasPrefix(lang, "en"):
		return "English"
	case strings.HasPrefix(lang, "ja"):
		return "Japanese"
	case strings.HasPrefix(lang, "ko"):
		return "Korean"
	case strings.HasPrefix(lang, "ru"):
		return "Russian"
	case strings.HasPrefix(lang, "ms"):
		return "Malay"
	case strings.HasPrefix(lang, "tr"):
		return "Turkish"
	case strings.HasPrefix(lang, "pt-br"):
		return "Brazilian Portuguese"
	case strings.HasPrefix(lang, "es"):
		return "Spanish"
	default:
		return lang
	}
}
