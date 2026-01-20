package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CommandHandler defines the interface for command handlers
type CommandHandler interface {
	Handle(line string) (string, error)
}

// ShellCommandHandler handles lines starting with '!'
type ShellCommandHandler struct{}

func (h *ShellCommandHandler) Handle(line string) (string, error) {
	// Remove the '!' prefix
	cmdStr := strings.TrimSpace(line[1:])
	if cmdStr == "" {
		return "", fmt.Errorf("empty command")
	}

	// Execute the command using the shell
	// On Windows, use cmd.exe /c; on Unix, use sh -c
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	output, err := cmd.CombinedOutput()
	result := string(output)
	if err != nil {
		// Return error message as string (for stdout output)
		return fmt.Sprintf("%s (error: %v)", result, err), nil
	}
	return result, nil
}

// FileTemplateHandler handles lines starting with '/file'
type FileTemplateHandler struct{}

func (h *FileTemplateHandler) Handle(line string) (string, error) {
	// Parse: /file filename
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", fmt.Errorf("usage: /file filename")
	}

	filename := parts[1]
	templatePath, err := getTemplatePath(filename)
	if err != nil {
		return "", fmt.Errorf("failed to get template path: %v", err)
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		// Return error message as string (for stdout output)
		return fmt.Sprintf("error: failed to read template file '%s': %v", filename, err), nil
	}

	return string(content), nil
}

// HandlerRegistry manages command handlers
type HandlerRegistry struct {
	handlers map[string]CommandHandler
}

func NewHandlerRegistry() *HandlerRegistry {
	registry := &HandlerRegistry{
		handlers: make(map[string]CommandHandler),
	}

	// Register built-in handlers
	registry.Register("file", &FileTemplateHandler{})

	return registry
}

func (r *HandlerRegistry) Register(commandName string, handler CommandHandler) {
	r.handlers[commandName] = handler
}

func (r *HandlerRegistry) GetHandler(commandName string) (CommandHandler, bool) {
	handler, exists := r.handlers[commandName]
	return handler, exists
}
