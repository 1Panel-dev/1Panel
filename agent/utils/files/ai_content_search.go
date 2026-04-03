package files

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

type FileAIContentHit struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

var fileAIBinaryExt = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {}, ".ico": {}, ".bmp": {},
	".zip": {}, ".gz": {}, ".tgz": {}, ".bz2": {}, ".xz": {}, ".7z": {}, ".rar": {},
	".exe": {}, ".dll": {}, ".so": {}, ".dylib": {}, ".bin": {},
	".pdf": {}, ".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {},
	".mp3": {}, ".mp4": {}, ".avi": {}, ".mkv": {}, ".mov": {},
	".woff": {}, ".woff2": {}, ".ttf": {}, ".eot": {},
}

func SearchFileAIContentHits(root string, inventory []AISearchInventoryItem, opts ContentSearchOptions, match func(string) bool) (hits []FileAIContentHit, scannedFiles int, scanTruncated bool) {
	if len(inventory) == 0 || match == nil {
		return nil, 0, false
	}
	root = filepath.Clean(root)
	maxFile := opts.MaxFileBytes
	if maxFile <= 0 {
		maxFile = int64(fileAIDefaultMaxFileBytes)
	}
	if maxFile > int64(fileAIHardMaxFileBytes) {
		maxFile = int64(fileAIHardMaxFileBytes)
	}
	maxScan := opts.MaxScanFiles
	if maxScan <= 0 {
		maxScan = fileAIDefaultMaxScanFiles
	}
	if maxScan > fileAIHardMaxScanFiles {
		maxScan = fileAIHardMaxScanFiles
	}
	maxPerFile := opts.MaxHitsPerFile
	if maxPerFile <= 0 {
		maxPerFile = fileAIDefaultMaxHitsPerFile
	}
	if maxPerFile > fileAIHardMaxHitsPerFile {
		maxPerFile = fileAIHardMaxHitsPerFile
	}
	maxTotal := opts.MaxTotalHits
	if maxTotal <= 0 {
		maxTotal = fileAIDefaultMaxTotalHits
	}
	if maxTotal > fileAIHardMaxTotalHits {
		maxTotal = fileAIHardMaxTotalHits
	}
	maxTextRunes := opts.MaxHitTextRunes
	if maxTextRunes <= 0 {
		maxTextRunes = fileAIDefaultMaxHitTextRunes
	}
	if maxTextRunes > fileAIHardMaxHitTextRunes {
		maxTextRunes = fileAIHardMaxHitTextRunes
	}

	hits = make([]FileAIContentHit, 0, 32)
	totalHits := 0

	for _, it := range inventory {
		rel := strings.TrimSpace(it.RelPath)
		if rel == "" {
			continue
		}
		if !opts.passesInventoryMeta(it, rel) {
			continue
		}
		if it.Size > maxFile {
			continue
		}
		if it.Size < 0 {
			continue
		}
		ext := strings.ToLower(filepath.Ext(rel))
		if _, bad := fileAIBinaryExt[ext]; bad {
			continue
		}

		fullPath := filepath.Join(root, filepath.FromSlash(rel))
		fullPath = filepath.Clean(fullPath)
		underRoot, err := filepath.Rel(root, fullPath)
		if err != nil || strings.HasPrefix(underRoot, "..") {
			continue
		}
		if ShouldFilterSensitivePath(fullPath) {
			continue
		}

		if scannedFiles >= maxScan {
			scanTruncated = true
			break
		}

		st, err := os.Stat(fullPath)
		if err != nil || st.IsDir() {
			continue
		}
		if st.Size() > maxFile {
			continue
		}

		fileHits, err := scanFileForMatch(fullPath, match, maxPerFile, maxTextRunes)
		scannedFiles++
		if err != nil {
			continue
		}
		for _, h := range fileHits {
			if totalHits >= maxTotal {
				scanTruncated = true
				return hits, scannedFiles, scanTruncated
			}
			hits = append(hits, h)
			totalHits++
		}
	}

	return hits, scannedFiles, scanTruncated
}

func scanFileForMatch(absPath string, match func(string) bool, maxPerFile, maxTextRunes int) ([]FileAIContentHit, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var out []FileAIContentHit
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if len(out) >= maxPerFile {
			break
		}
		line := scanner.Text()
		if !match(line) {
			continue
		}
		out = append(out, FileAIContentHit{
			Path: absPath,
			Line: lineNo,
			Text: truncateHitLine(line, maxTextRunes),
		})
	}
	return out, scanner.Err()
}

func truncateHitLine(s string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = fileAIDefaultMaxHitTextRunes
	}
	if utf8.RuneCountInString(s) <= maxRunes {
		return strings.TrimRight(s, "\r\n")
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= maxRunes {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		n++
	}
	return strings.TrimRight(b.String(), "\r\n")
}

func FormatContentHitsForPrompt(hits []FileAIContentHit, maxBytes int, matchDesc string) string {
	if len(hits) == 0 || maxBytes <= 0 {
		return ""
	}
	var b strings.Builder
	if matchDesc == "" {
		matchDesc = "search"
	}
	b.WriteString("Content line matches (" + matchDesc + "). Do not invent lines; only reference these.\n")
	for _, h := range hits {
		var line string
		if h.Line <= 0 {
			line = h.Path + " (filename): " + h.Text + "\n"
		} else {
			line = h.Path + ":" + strconv.Itoa(h.Line) + ": " + h.Text + "\n"
		}
		if b.Len()+len(line) > maxBytes {
			b.WriteString("... (further hits omitted for prompt size)\n")
			break
		}
		b.WriteString(line)
	}
	return b.String()
}
