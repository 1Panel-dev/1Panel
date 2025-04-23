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
	content, err := marshal(envMap)
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

func marshal(envMap map[string]string) (string, error) {
	lines := make([]string, 0, len(envMap))
	for k, v := range envMap {
		if d, err := strconv.Atoi(v); err == nil && !isStartWithZero(v) {
			lines = append(lines, fmt.Sprintf(`%s=%d`, k, d))
		} else if hasEvenDoubleQuotes(v) {
			lines = append(lines, fmt.Sprintf(`%s='%s'`, k, v))
		} else {
			lines = append(lines, fmt.Sprintf(`%s="%s"`, k, v))
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n"), nil
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

func isStartWithZero(value string) bool {
	if strings.HasPrefix(value, "0") && len(value) > 1 {
		return true
	}
	return false
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
