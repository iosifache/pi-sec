package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LookupEnvOrDotEnv(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value, nil
	}

	file, err := os.Open(".env")
	if err != nil {
		return "", fmt.Errorf("open .env: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		keyPart, valuePart, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		if strings.TrimSpace(keyPart) != key {
			continue
		}

		value := strings.TrimSpace(valuePart)
		value = strings.Trim(value, `"'`)
		if value == "" {
			break
		}
		return value, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan .env: %w", err)
	}

	return "", fmt.Errorf("%s is not set", key)
}
