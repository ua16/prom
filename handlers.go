package main

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
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
		return "", fmt.Errorf("%s (error: %v)", result, err)
	}
	return result, nil
}

// TemplateHandler handles lines starting with '/templ'
type TemplateHandler struct{}

func (h *TemplateHandler) Handle(line string) (string, error) {
	// Parse: /templ templatename
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", fmt.Errorf("usage: /templ templatename")
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

// FileInsertHandler handles lines starting with '/file'. It inserts
type FileInsertHandler struct{}

func (h *FileInsertHandler) Handle(line string) (string, error) {
	// Parse: /file filename
	// Looks in cwd for file and inserts it. When called from vim, this'll be project root
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to get file path")
	}

	filename := parts[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error: failed to read file '%s' : %v", filename, err)
	}

	return string(content), nil
}

type WebInsertHandler struct{}

func (h *WebInsertHandler) Handle(line string) (string, error) {
	// Parse: /webraw link
	// Copies the raw file via http and inserts it.
	// Note that the html will get stored in memory

	// Checks if html should be sanitized ie remove anything that could trigger a command
	sanitize := true

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to get file path")
	} else if len(parts) == 3 && parts[2] == "unsafe" {
		sanitize = false
	}

	url := parts[1]

	htmlString, err := getHTMLString(url)
	if err != nil {
		return "", err
	}

	if sanitize {
		return sanitizeString(htmlString), nil
	} else {
		return htmlString, nil
	}
}

type WebInsertMdHandler struct{}

func (h *WebInsertMdHandler) Handle(line string) (string, error) {
	// Parse: /webmd link sanitize
	// Copies the raw file via http and inserts it.
	// Note that the html will get stored in memory

	// Checks if html should be sanitized ie remove anything that could trigger a command
	sanitize := true
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to get file path")
	} else if len(parts) == 3 && parts[2] == "unsafe" {
		sanitize = false
	}

	url := parts[1]
	htmlString, err := getHTMLString(url)
	if err != nil {
		return "", err
	}

	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(htmlString)
	if err != nil {
		return "", fmt.Errorf("Error converting %s to markdown %v ", url, err)
	}

	if sanitize {
		return sanitizeString(markdown), nil
	}

	return markdown, nil

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
	registry.Register("templ", &TemplateHandler{})
	registry.Register("file", &FileInsertHandler{})
	registry.Register("webraw", &WebInsertHandler{})
	registry.Register("webmd", &WebInsertMdHandler{})

	return registry
}

func (r *HandlerRegistry) Register(commandName string, handler CommandHandler) {
	r.handlers[commandName] = handler
}

func (r *HandlerRegistry) GetHandler(commandName string) (CommandHandler, bool) {
	handler, exists := r.handlers[commandName]
	return handler, exists
}
