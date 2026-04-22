package files

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

const (
	fileAIHardMaxScanFiles       = 800
	fileAIDefaultMaxScanFiles    = 150
	fileAIUserMaxScanFiles       = 500
	fileAIHardMaxFileBytes       = 4 * 1024 * 1024
	fileAIDefaultMaxFileBytes    = 512 * 1024
	fileAIUserMaxFileBytes       = 4 * 1024 * 1024
	fileAIHardMaxHitsPerFile     = 200
	fileAIDefaultMaxHitsPerFile  = 40
	fileAIHardMaxTotalHits       = 2000
	fileAIDefaultMaxTotalHits    = 500
	fileAIHardMaxHitTextRunes    = 2000
	fileAIDefaultMaxHitTextRunes = 500
	fileAIHardMaxPromptHitBytes  = 50000
	fileAIDefaultPromptHitBytes  = 12000
	fileAIMaxRegexPatternRunes   = 512
)

type ContentSearchOptions struct {
	MatchCase                 bool
	WholeWord                 bool
	UseRegex                  bool
	ExtensionAllow            map[string]struct{}
	MinSize                   int64
	MaxSize                   int64
	ModAfter                  *time.Time
	ModBefore                 *time.Time
	MaxScanFiles              int
	MaxFileBytes              int64
	MaxHitsPerFile            int
	MaxTotalHits              int
	MaxHitTextRunes           int
	ContentHitsPromptMaxBytes int
	LlmMaxOutputTokens        int
}

func defaultContentSearchOptions() ContentSearchOptions {
	return ContentSearchOptions{
		MaxScanFiles:              fileAIDefaultMaxScanFiles,
		MaxFileBytes:              int64(fileAIDefaultMaxFileBytes),
		MaxHitsPerFile:            fileAIDefaultMaxHitsPerFile,
		MaxTotalHits:              fileAIDefaultMaxTotalHits,
		MaxHitTextRunes:           fileAIDefaultMaxHitTextRunes,
		ContentHitsPromptMaxBytes: fileAIDefaultPromptHitBytes,
	}
}

func MergeContentSearchOptions(
	matchCase, wholeWord, useRegex bool,
	extensions []string,
	minSize, maxSize int64,
	modAfter, modBefore string,
	maxScanFiles int,
	maxFileBytes int64,
	maxHitsPerFile, maxTotalHits, promptHitBytes, llmMaxOut int,
) (ContentSearchOptions, error) {
	o := defaultContentSearchOptions()
	o.MatchCase = matchCase
	o.WholeWord = wholeWord
	o.UseRegex = useRegex
	if len(extensions) > 0 {
		o.ExtensionAllow = make(map[string]struct{}, len(extensions))
		for _, e := range extensions {
			e = strings.TrimSpace(strings.ToLower(e))
			if e == "" {
				continue
			}
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			o.ExtensionAllow[e] = struct{}{}
		}
		if len(o.ExtensionAllow) == 0 {
			o.ExtensionAllow = nil
		}
	}
	o.MinSize = minSize
	o.MaxSize = maxSize
	if minSize > 0 && maxSize > 0 && minSize > maxSize {
		return ContentSearchOptions{}, fmt.Errorf("minSize must be <= maxSize")
	}
	if strings.TrimSpace(modAfter) != "" {
		t, e := ParseFileSearchTime(modAfter)
		if e != nil {
			return ContentSearchOptions{}, fmt.Errorf("modifiedAfter: %w", e)
		}
		o.ModAfter = &t
	}
	if strings.TrimSpace(modBefore) != "" {
		t, e := ParseFileSearchTime(modBefore)
		if e != nil {
			return ContentSearchOptions{}, fmt.Errorf("modifiedBefore: %w", e)
		}
		o.ModBefore = &t
	}
	if o.ModAfter != nil && o.ModBefore != nil && o.ModAfter.After(*o.ModBefore) {
		return ContentSearchOptions{}, fmt.Errorf("modifiedAfter must be before modifiedBefore")
	}
	if maxScanFiles > 0 {
		o.MaxScanFiles = clampInt(maxScanFiles, 1, fileAIUserMaxScanFiles)
	}
	if o.MaxScanFiles > fileAIHardMaxScanFiles {
		o.MaxScanFiles = fileAIHardMaxScanFiles
	}
	if maxFileBytes > 0 {
		o.MaxFileBytes = clampInt64(maxFileBytes, 1024, fileAIUserMaxFileBytes)
	}
	if o.MaxFileBytes > int64(fileAIHardMaxFileBytes) {
		o.MaxFileBytes = int64(fileAIHardMaxFileBytes)
	}
	if maxHitsPerFile > 0 {
		o.MaxHitsPerFile = clampInt(maxHitsPerFile, 1, fileAIHardMaxHitsPerFile)
	}
	if maxTotalHits > 0 {
		o.MaxTotalHits = clampInt(maxTotalHits, 1, fileAIHardMaxTotalHits)
	}
	if promptHitBytes > 0 {
		o.ContentHitsPromptMaxBytes = clampInt(promptHitBytes, 2048, fileAIHardMaxPromptHitBytes)
	}
	if llmMaxOut > 0 {
		o.LlmMaxOutputTokens = clampInt(llmMaxOut, 256, 4096)
	}
	return o, nil
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampInt64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func ParseFileSearchTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	loc := time.Local
	if t, err := time.ParseInLocation(constant.DateTimeLayout, s, loc); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid time format, use RFC3339 or %s", constant.DateTimeLayout)
}

func (o ContentSearchOptions) ContentMatchDescription() string {
	switch {
	case o.UseRegex:
		if o.MatchCase {
			return "regex (case-sensitive)"
		}
		return "regex (case-insensitive)"
	case o.WholeWord:
		if o.MatchCase {
			return "whole-word literal (case-sensitive)"
		}
		return "whole-word literal (case-insensitive)"
	default:
		if o.MatchCase {
			return "substring (case-sensitive)"
		}
		return "substring (case-insensitive)"
	}
}

func NewContentLineMatcher(query string, o ContentSearchOptions) (func(string) bool, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	if o.UseRegex {
		if utf8.RuneCountInString(q) > fileAIMaxRegexPatternRunes {
			return nil, fmt.Errorf("regex pattern too long (max %d runes)", fileAIMaxRegexPatternRunes)
		}
		pat := q
		if !o.MatchCase {
			pat = "(?i)" + pat
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		return func(line string) bool { return re.FindStringIndex(line) != nil }, nil
	}
	if o.WholeWord {
		escaped := regexp.QuoteMeta(q)
		pat := `\b` + escaped + `\b`
		if !o.MatchCase {
			pat = "(?i)" + pat
		}
		re, err := regexp.Compile(pat)
		if err != nil {
			return nil, err
		}
		return func(line string) bool { return re.FindStringIndex(line) != nil }, nil
	}
	qCmp := q
	if !o.MatchCase {
		qCmp = strings.ToLower(q)
		return func(line string) bool {
			return strings.Contains(strings.ToLower(line), qCmp)
		}, nil
	}
	return func(line string) bool { return strings.Contains(line, q) }, nil
}

func (o ContentSearchOptions) passesInventoryMeta(it AISearchInventoryItem, rel string) bool {
	if it.IsDir {
		return false
	}
	ext := strings.ToLower(filepath.Ext(rel))
	if len(o.ExtensionAllow) > 0 {
		if _, ok := o.ExtensionAllow[ext]; !ok {
			return false
		}
	}
	if o.MinSize > 0 && it.Size < o.MinSize {
		return false
	}
	if o.MaxSize > 0 && it.Size > o.MaxSize {
		return false
	}
	if o.ModAfter != nil || o.ModBefore != nil {
		mt, err := time.ParseInLocation(constant.DateTimeLayout, it.ModTime, time.Local)
		if err != nil {
			return true
		}
		if o.ModAfter != nil && mt.Before(*o.ModAfter) {
			return false
		}
		if o.ModBefore != nil && mt.After(*o.ModBefore) {
			return false
		}
	}
	return true
}
