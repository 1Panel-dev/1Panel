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
		if d, err := strconv.Atoi(v); err == nil && !isStartWithZero(v) {
			lines = append(lines, fmt.Sprintf(`%s=%d`, k, d))
		} else if hasEvenDoubleQuotes(v) {
			lines = append(lines, fmt.Sprintf(`%s='%s'`, k, v))
		} else {
			lines = append(lines, fmt.Sprintf(`%s="%s"`, k, v))
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
	if d, err := strconv.Atoi(v); err == nil && !isStartWithZero(v) {
		return fmt.Sprintf(`%s=%d`, k, d)
	} else if hasEvenDoubleQuotes(v) {
		return fmt.Sprintf(`%s='%s'`, k, v)
	} else {
		return fmt.Sprintf(`%s="%s"`, k, v)
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
