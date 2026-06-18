package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DirectoryEntry represents a directory saved in qcd
type DirectoryEntry struct {
	Path         string    `json:"path"`
	Score        int       `json:"score"`
	LastAccessed time.Time `json:"last_accessed"`
}

// Config wraps the list of saved directories
type Config struct {
	Directories []DirectoryEntry `json:"directories"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "qcd")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "dirs.json"), nil
}

func loadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{Directories: []DirectoryEntry{}}, nil
		}
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		// If file is empty or corrupted, return empty config
		return &Config{Directories: []DirectoryEntry{}}, nil
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

func addDirectory(dirPath string) (string, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return absPath, fmt.Errorf("directory does not exist: %w", err)
	}
	if !info.IsDir() {
		return absPath, fmt.Errorf("path is not a directory")
	}

	cfg, err := loadConfig()
	if err != nil {
		return absPath, err
	}

	for _, entry := range cfg.Directories {
		if entry.Path == absPath {
			return absPath, fmt.Errorf("already registered")
		}
	}

	cfg.Directories = append(cfg.Directories, DirectoryEntry{
		Path:         absPath,
		Score:        0,
		LastAccessed: time.Now(),
	})

	return absPath, saveConfig(cfg)
}

func removeDirectory(dirPath string) (string, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", err
	}

	cfg, err := loadConfig()
	if err != nil {
		return absPath, err
	}

	foundIndex := -1
	for i, entry := range cfg.Directories {
		if entry.Path == absPath {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		// Also try comparing clean paths, in case of trailing slashes etc.
		cleanAbs := filepath.Clean(absPath)
		for i, entry := range cfg.Directories {
			if filepath.Clean(entry.Path) == cleanAbs {
				foundIndex = i
				break
			}
		}
	}

	if foundIndex == -1 {
		return absPath, fmt.Errorf("not registered in qcd")
	}

	cfg.Directories = append(cfg.Directories[:foundIndex], cfg.Directories[foundIndex+1:]...)
	return absPath, saveConfig(cfg)
}

func incrementScore(absPath string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	for i, entry := range cfg.Directories {
		if entry.Path == absPath {
			cfg.Directories[i].Score++
			cfg.Directories[i].LastAccessed = time.Now()
			break
		}
	}

	return saveConfig(cfg)
}

func getSortedDirectories() ([]DirectoryEntry, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	sort.Slice(cfg.Directories, func(i, j int) bool {
		if cfg.Directories[i].Score != cfg.Directories[j].Score {
			return cfg.Directories[i].Score > cfg.Directories[j].Score
		}
		return cfg.Directories[i].LastAccessed.After(cfg.Directories[j].LastAccessed)
	})

	return cfg.Directories, nil
}
