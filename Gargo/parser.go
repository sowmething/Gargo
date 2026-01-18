package main

import (
	"bufio"
	"os"
	"strings"
)

type Config map[string]map[string]string

func Parse(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(Config)
	currentsection := "default"
	config[currentsection] = make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentsection = line[1 : len(line)-1]
			if _, ok := config[currentsection]; !ok {
				config[currentsection] = make(map[string]string)
			}
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[currentsection][key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}
