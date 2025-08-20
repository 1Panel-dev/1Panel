package webdav

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
)

func PathEscape(path string) string {
	s := strings.Split(path, "/")
	for i, e := range s {
		s[i] = url.PathEscape(e)
	}
	return strings.Join(s, "/")
}

func FixSlash(s string) string {
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}

func SplitPathToHierarchy(fullPath string) []string {
	cleanPath := path.Clean(fullPath)
	parts := strings.Split(cleanPath, "/")

	var result []string
	currentPath := ""

	for _, part := range parts {
		if part == "" {
			currentPath = "/"
			result = append(result, currentPath)
			continue
		}

		if currentPath == "/" {
			currentPath = path.Join(currentPath, part)
		} else {
			currentPath = path.Join(currentPath, part)
		}

		result = append(result, currentPath)
	}

	return result
}

func FixSlashes(s string) string {
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}

	return FixSlash(s)
}

func Join(path0 string, path1 string) string {
	return strings.TrimSuffix(path0, "/") + "/" + strings.TrimPrefix(path1, "/")
}

func String(r io.Reader) string {
	buf := new(bytes.Buffer)
	// TODO - make String return an error as well
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

func parseInt64(s *string) int64 {
	if n, e := strconv.ParseInt(*s, 10, 64); e == nil {
		return n
	}
	return 0
}

func parseXML(data io.Reader, resp interface{}, parse func(resp interface{}) error) error {
	decoder := xml.NewDecoder(data)
	for t, _ := decoder.Token(); t != nil; t, _ = decoder.Token() {
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "response" {
				if e := decoder.DecodeElement(resp, &se); e == nil {
					if err := parse(resp); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
