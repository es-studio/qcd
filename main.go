package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "add":
			path := "."
			if len(os.Args) > 2 {
				path = os.Args[2]
			}
			absPath, err := addDirectory(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Successfully added: %s\n", absPath)
			return

		case "del":
			path := "."
			if len(os.Args) > 2 {
				path = os.Args[2]
			}
			absPath, err := removeDirectory(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Successfully deleted: %s\n", absPath)
			return

		case "init":
			// Generate the shell integration function using the current executable path
			execPath, err := os.Executable()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
				os.Exit(1)
			}
			execPath = filepath.ToSlash(execPath)
			
			// Output zsh/bash compatible function
			// Uses a temporary file mechanism to avoid hijacking TUI stdin/stdout via subshells
			fmt.Printf(`qcd() {
    local cmd="$1"
    if [ "$cmd" = "add" ] || [ "$cmd" = "del" ]; then
        "%s" "$@"
    elif [ "$cmd" = "init" ]; then
        "%s" init
    elif [ "$cmd" = "install" ]; then
        "%s" install
    elif [ "$cmd" = "--help" ] || [ "$cmd" = "-h" ]; then
        "%s" --help
    else
        local tmpfile
        tmpfile=$(mktemp -t qcd.XXXXXX)
        
        QCD_TMP_FILE="$tmpfile" "%s" "$@"
        
        local target_dir
        if [ -f "$tmpfile" ]; then
            target_dir=$(cat "$tmpfile")
            rm -f "$tmpfile"
        fi
        
        if [ -n "$target_dir" ] && [ -d "$target_dir" ]; then
            cd "$target_dir" || return
        fi
    fi
}
`, execPath, execPath, execPath, execPath, execPath)
			return

		case "install":
			err := installShellIntegration()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error installing: %v\n", err)
				os.Exit(1)
			}
			return

		case "--help", "-h":
			printHelp()
			return

		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
			printHelp()
			os.Exit(1)
		}
	}

	// No arguments -> run interactive TUI
	selected, err := runTUI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	if selected != "" {
		tmpFile := os.Getenv("QCD_TMP_FILE")
		if tmpFile != "" {
			// Write the selected path to the temporary file for the shell wrapper
			err := os.WriteFile(tmpFile, []byte(selected), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing target path: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Fallback: output to stdout if run directly outside shell wrapper
			fmt.Print(selected)
		}
	}
}

func installShellIntegration() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath = filepath.ToSlash(execPath)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	shell := os.Getenv("SHELL")
	var rcFile string
	if strings.Contains(shell, "zsh") {
		rcFile = filepath.Join(home, ".zshrc")
	} else if strings.Contains(shell, "bash") {
		rcFile = filepath.Join(home, ".bashrc")
	} else {
		return fmt.Errorf("unsupported shell: %s. Please manually add: eval \"$(%s init)\" to your shell config", shell, execPath)
	}

	// Read existing rc file
	content, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", rcFile, err)
	}

	integrationLine := fmt.Sprintf("\n# qcd shell integration\neval \"$(%s init)\"\n", execPath)

	// Check if already integrated
	if strings.Contains(string(content), "qcd shell integration") || strings.Contains(string(content), execPath+" init") {
		return fmt.Errorf("qcd is already installed in %s", rcFile)
	}

	// Open for appending
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", rcFile, err)
	}
	defer f.Close()

	if _, err := f.WriteString(integrationLine); err != nil {
		return fmt.Errorf("failed to write to %s: %w", rcFile, err)
	}

	fmt.Printf("Successfully added shell integration to %s\n", rcFile)
	fmt.Printf("Please run: source %s  or restart your terminal to activate 'qcd'.\n", rcFile)
	return nil
}

func printHelp() {
	fmt.Fprintf(os.Stderr, `qcd - Quick Directory Changer

Usage:
  qcd add [path]  - Register a directory (defaults to current directory ".")
  qcd del [path]  - Deregister a directory (defaults to current directory ".")
  qcd             - Show interactive list of registered directories (frequent paths first)
  qcd install     - Automatically add qcd to your shell config (.zshrc or .bashrc)
  qcd init        - Output shell integration script (run: eval "$(qcd init)")
  qcd --help, -h  - Show this help message
`)
}
