package env

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Write(envMap map[string]string, filename string) error {
	content, err := Marshal(envMap)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content + "\n")
	if err != nil {
		return err
	}
	return file.Sync()
}

func Marshal(envMap map[string]string) (string, error) {
	lines := make([]string, 0, len(envMap))
	for k, v := range envMap {
		if d, ok := canonicalInt(v); ok {
			lines = append(lines, fmt.Sprintf(`%s=%d`, k, d))
		} else if shouldUseSingleQuotes(v) {
			lines = append(lines, fmt.Sprintf(`%s='%s'`, k, v))
		} else {
			lines = append(lines, fmt.Sprintf(`%s="%s"`, k, escapeDoubleQuotedEnvValue(v)))
		}
	}
	return strings.Join(lines, "\n"), nil
}

func WriteWithOrder(envMap map[string]string, filename string, orders []string) error {
	content, err := MarshalWithOrder(envMap, orders)
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content + "\n")
	if err != nil {
		return err
	}
	return file.Sync()
}

func MarshalWithOrder(envMap map[string]string, orders []string) (string, error) {
	lines := make([]string, 0, len(envMap))
	for _, k := range orders {
		if v, ok := envMap[k]; ok {
			lines = append(lines, formatEnvLine(k, v))
		}
	}

	extraKeys := make([]string, 0)
	for k := range envMap {
		found := false
		for _, okk := range orders {
			if k == okk {
				found = true
				break
			}
		}
		if !found {
			extraKeys = append(extraKeys, k)
		}
	}
	sort.Strings(extraKeys)
	for _, k := range extraKeys {
		lines = append(lines, formatEnvLine(k, envMap[k]))
	}
	return strings.Join(lines, "\n"), nil
}

func formatEnvLine(k, v string) string {
	if d, ok := canonicalInt(v); ok {
		return fmt.Sprintf(`%s=%d`, k, d)
	} else if shouldUseSingleQuotes(v) {
		return fmt.Sprintf(`%s='%s'`, k, v)
	} else {
		return fmt.Sprintf(`%s="%s"`, k, escapeDoubleQuotedEnvValue(v))
	}
}

func GetEnvValueByKey(envPath, key string) (string, error) {
	envMap, err := godotenv.Read(envPath)
	if err != nil {
		return "", err
	}
	value, ok := envMap[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in %s", key, envPath)
	}
	return value, nil
}

// canonicalInt reports whether value is an integer that formats back to the
// same string, so values such as "007", "+1" or "-01" keep their exact text.
func canonicalInt(value string) (int, bool) {
	d, err := strconv.Atoi(value)
	if err != nil || strconv.Itoa(d) != value {
		return 0, false
	}
	return d, true
}

func hasEvenDoubleQuotes(s string) bool {
	count := 0
	for _, ch := range s {
		if ch == '"' {
			count++
		}
	}
	return count%2 == 0
}

func shouldUseSingleQuotes(s string) bool {
	return hasEvenDoubleQuotes(s) && !strings.Contains(s, `\`) && !strings.Contains(s, `'`)
}

func escapeDoubleQuotedEnvValue(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"\n", `\n`,
	)
	return replacer.Replace(s)
}
