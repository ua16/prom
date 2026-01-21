package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func getHTMLString(url string) (string, error) {

	// Fetch the website
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Error fetching URL %v\n", err)

	}

	defer resp.Body.Close()

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response %v\n", err)
	}

	return string(htmlBytes), nil
}

func sanitizeString(raw string) (string) {
		lines := strings.Split(raw, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "!") {
				lines[i] = `\` + line
			} else if strings.HasPrefix(line, "/") {
				lines[i] = `\` + line
			}
		}
		return strings.Join(lines, "\n")

}
