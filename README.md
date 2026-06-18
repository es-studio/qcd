# qcd - Quick Directory Changer

[한국어 README](README.ko.md)

`qcd` is a Go-based CLI tool designed to help you navigate directories quickly and conveniently in your terminal (Shell). It provides a beautiful TUI powered by Bubble Tea, and automatically displays your most frequently visited directories at the top of the list.

## Key Features
- **Register Directory**: `qcd add [path]` (registers the current directory if path is omitted)
- **Delete Directory**: `qcd del [path]` (deletes the current directory from the list if path is omitted)
- **Interactive Navigation**: `qcd`
  - Select a directory using arrow keys (`↑`/`↓` or `k`/`j`) and press `Enter` to change to that directory instantly.
  - Cancel navigation with `Esc` or `q`.
  - The directory list is sorted by **visit frequency (Score)**, and directories with the same score are sorted by **last accessed time (most recent first)**.

---

## How to Build
Ensure Go (1.25 or higher recommended) is installed, then run the following command to build the binary.

```bash
go build -o qcd
```

---

## Installation & Shell Integration

To execute the `cd` command directly in the shell, `qcd` must be registered as a shell function. `qcd` provides an `install` command that automatically sets up this integration for your active shell.

### 1. Automatic Installation (Recommended)
Place the built binary in your preferred directory and run:

```bash
./qcd install
```

This will detect your active shell environment (`.zshrc` or `.bashrc`) and append the shell function integration code to the shell config file.
After installation, **open a new terminal session or run `source ~/.zshrc` (or `source ~/.bashrc`)** to apply the changes.

### 2. Manual Installation (For moving binary to a custom system path)
If you prefer to copy the binary to a system PATH like `/usr/local/bin`, follow these steps:

1. Copy the binary:
   ```bash
   mv qcd /usr/local/bin/qcd
   ```
2. Manually append the following code to the bottom of your shell configuration file (e.g., `~/.zshrc` or `~/.bashrc`):
   ```bash
   # Register qcd shell function
   eval "$(/usr/local/bin/qcd init)"
   ```
3. Restart your terminal or source the configuration file.

---

## Usage Guide

### 1. Add Current Directory
Navigate to your frequently visited directory and run:
```bash
qcd add .
```
You can also specify a target path to register:
```bash
qcd add ~/Projects/my-app
```

### 2. Remove Current Directory
To remove the current directory from the list, run:
```bash
qcd del .
```

### 3. Change Directory
Run `qcd` without any arguments to open the registered directory list interface:
```bash
qcd
```
Select the directory you want to go to and press `Enter`. The active directory in your shell will instantly change to the selected path.

---

## Storage Location
The registered directory information and accessibility scores are saved in JSON format at:
- Mac / Linux: `~/.config/qcd/dirs.json`
