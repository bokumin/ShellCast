# ShellCast

A tool for streaming and recording terminal sessions with customizable themes, screen layouts, and timestamps.

## Features

- Stream terminal output to RTMP servers (e.g., Twitch, YouTube)
- Record terminal sessions to text files with timestamps
- Split screen mode to run and display multiple commands simultaneously
- Customizable themes with presets (hacker, solarized, light, monokai)
- Adjustable screen size, font size, and colors
- Interactive mode for easy management and configuration
- Timestamp display with configurable formats
- Configuration saving and loading

## Files

- `config.go` - Configuration handling, theme presets
- `shellcast.go` - Core functionality for command execution, streaming, and recording
- `interactive.go` - Interactive CLI mode
- `main.go` - Command-line interface and application entry point

## Usage

### Basic Usage

```bash
# Run in interactive mode
go run *.go -interactive

# Execute a command and stream output to RTMP server
go run *.go -rtmp rtmp://server/path command args

# Execute a command with custom theme and timestamps
go run *.go -theme hacker -timestamp on ls -la

# Run multiple commands in split screen mode
go run *.go -split "ls -la" "top -n 1" "df -h"
```

### Command-line Options

```
  -bg-color string
        Background color for streaming (default "black")
  -config string
        Path to configuration file
  -ffmpeg string
        Path to FFmpeg executable
  -font-color string
        Font color for streaming (default "white")
  -font-size int
        Font size for streaming (default 24)
  -interactive
        Run in interactive mode
  -list-themes
        List available theme presets
  -record
        Record session to file
  -record-path string
        Directory to save recordings (default "./recordings")
  -rtmp string
        RTMP URL to stream to
  -screen-size string
        Screen size for streaming (WIDTHxHEIGHT) (default "1280x720")
  -split
        Run commands in split screen mode
  -theme string
        Theme preset to use (default "default")
  -timestamp
        Show timestamps in output
  -timestamp-format string
        Format for timestamps (default "2006-01-02 15:04:05")
```

### Interactive Mode Commands

- `help` - Show available commands
- `exit`, `quit` - Exit ShellCast
- `stream` - Start streaming (prompts for RTMP URL if not set)
- `stop` - Stop streaming
- `record` - Start recording the session
- `stoprecord` - Stop recording the session
- `theme [NAME]` - List themes or apply a theme by name
- `timestamp [on|off]` - Enable or disable timestamps
- `size [WxH]` - Show or set screen size (e.g., 1280x720)
- `split "cmd1" "cmd2"` - Run multiple commands in split screen mode
- `fontsize [SIZE]` - Show or set font size
- `save [FILE]` - Save configuration to a file
- `load [FILE]` - Load configuration from a file

## Available Themes

- `default` - White text on black background
- `hacker` - Lime text on black background
- `solarized` - Solarized color scheme
- `light` - Dark text on light background
- `monokai` - Monokai-inspired color scheme

## Requirements

- Go 1.16 or higher
- FFmpeg (for streaming functionality)
- RTMP server (for streaming destination)

## Building

```bash
go build -o shellcast *.go
```

## Examples

```bash
# Simple recording with timestamps
./shellcast -timestamp on -record ps aux

# Streaming to Twitch with hacker theme
./shellcast -rtmp rtmp://live.twitch.tv/app/YOUR_STREAM_KEY -theme hacker top

# Split screen with different commands
./shellcast -split -theme solarized -timestamp on "ls -la" "top -n 1" "netstat -an"
```
