package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	registry := NewHandlerRegistry()
	scanner := bufio.NewScanner(os.Stdin)

	// Read all input first
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	// Process the text 3 times to allow recursive template expansion
	text := strings.Join(lines, "\n")
	for i := 0; i < 3; i++ {
		text = processText(text, registry)
	}

	// Output the final result
	fmt.Print(text)
}

// processText processes all lines in a text body and returns the processed text
func processText(text string, registry *HandlerRegistry) string {
	lines := strings.Split(text, "\n")
	var result strings.Builder

	for _, line := range lines {
		processed := processLine(line, registry)
		result.WriteString(processed)
	}

	return result.String()
}

// processLine processes a single line and returns the output
func processLine(line string, registry *HandlerRegistry) string {
	if len(line) == 0 {
		return line + "\n"
	}

	// Handle escaping: \! and \/ at line start
	if len(line) >= 2 && line[0] == '\\' {
		nextChar := line[1]
		if nextChar == '!' || nextChar == '/' {
			// Output the escaped character and the rest of the line
			return line[1:] + "\n"
		}
	}

	// Check for command execution: ! at column 0
	if line[0] == '!' {
		handler := &ShellCommandHandler{}
		result, err := handler.Handle(line)
		if err != nil {
			result = fmt.Sprintf("error: %v", err)
		}
		// Ensure result ends with newline if it doesn't already
		if result != "" && !strings.HasSuffix(result, "\n") {
			result += "\n"
		}
		return result
	}

	// Check for command pattern: /commandname at column 0
	if line[0] == '/' {
		// Parse the command name (everything after / until first space)
		spaceIdx := strings.Index(line, " ")
		if spaceIdx == -1 {
			spaceIdx = len(line)
		}
		commandName := line[1:spaceIdx]

		handler, exists := registry.GetHandler(commandName)
		if exists {
			result, err := handler.Handle(line)
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
			}
			// Ensure result ends with newline if it doesn't already
			if result != "" && !strings.HasSuffix(result, "\n") {
				result += "\n"
			}
			return result
		}
		// If command not found, treat as normal text
	}

	// Normal text: output as-is with newline
	return line + "\n"
}
