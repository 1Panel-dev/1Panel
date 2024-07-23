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
		if d, err := strconv.Atoi(v); err == nil {
			lines = append(lines, fmt.Sprintf(`%s=%d`, k, d))
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
