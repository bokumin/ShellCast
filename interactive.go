package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// InteractiveOptions
type InteractiveOptions struct {
	ConfigPath string
}

// Interactive shell
func RunInteractiveMode(sc *ShellCast, options InteractiveOptions) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("ShellCast Interactive Mode")
	fmt.Println("==========================")
	fmt.Println("Type 'help' for available commands")
	fmt.Println("Type 'exit' or 'quit' to exit")

	for {
		fmt.Print("\nshellcast> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Split input into command and arguments
		parts := strings.SplitN(input, " ", 2)
		cmd := strings.ToLower(parts[0])
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		// Process commands
		switch cmd {
		case "exit", "quit":
			return

		case "help":
			showHelp()

		case "stream":
			if sc.config.RTMPUrl == "" {
				fmt.Print("Enter RTMP URL: ")
				rtmpUrl, _ := reader.ReadString('\n')
				rtmpUrl = strings.TrimSpace(rtmpUrl)
				if rtmpUrl == "" {
					fmt.Println("No RTMP URL provided")
					continue
				}
				sc.config.RTMPUrl = rtmpUrl
			}

			if err := sc.StartStreaming(); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting stream: %v\n", err)
			}

		case "stop":
			if err := sc.StopStreaming(); err != nil {
				fmt.Fprintf(os.Stderr, "Error stopping stream: %v\n", err)
			}

		case "record":
			if err := sc.StartRecording(); err != nil {
				fmt.Fprintf(os.Stderr, "Error starting recording: %v\n", err)
			}

		case "stoprecord":
			if err := sc.StopRecording(); err != nil {
				fmt.Fprintf(os.Stderr, "Error stopping recording: %v\n", err)
			}

		case "theme":
			if args == "" {
				ListThemes()
				continue
			}

			if err := sc.config.ApplyTheme(args); err != nil {
				fmt.Fprintf(os.Stderr, "Error applying theme: %v\n", err)
			} else {
				fmt.Printf("Applied theme: %s\n", args)
			}

		case "timestamp":
			switch args {
			case "on":
				sc.config.ShowTimestamp = true
				fmt.Println("Timestamps enabled")
			case "off":
				sc.config.ShowTimestamp = false
				fmt.Println("Timestamps disabled")
			default:
				fmt.Println("Usage: timestamp [on|off]")
			}

		case "size":
			if args == "" {
				fmt.Printf("Current screen size: %dx%d\n",
					sc.config.ScreenWidth, sc.config.ScreenHeight)
				continue
			}

			var width, height int
			if _, err := fmt.Sscanf(args, "%dx%d", &width, &height); err != nil {
				fmt.Println("Usage: size WIDTHxHEIGHT (e.g., 1280x720)")
				continue
			}

			sc.config.ScreenWidth = width
			sc.config.ScreenHeight = height
			fmt.Printf("Screen size set to %dx%d\n", width, height)

		case "split":
			// Parse command list
			if args == "" {
				fmt.Println("Usage: split \"command1\" \"command2\" ...")
				continue
			}

			// Very simple parsing for demonstration
			commands := strings.Split(args, "\" \"")
			commands[0] = strings.TrimPrefix(commands[0], "\"")
			commands[len(commands)-1] = strings.TrimSuffix(commands[len(commands)-1], "\"")

			fmt.Printf("Running %d commands in split mode\n", len(commands))
			if err := sc.ExecuteSplitCommands(commands); err != nil {
				fmt.Fprintf(os.Stderr, "Error executing split commands: %v\n", err)
			}

		case "fontsize":
			if args == "" {
				fmt.Printf("Current font size: %d\n", sc.config.FontSize)
				continue
			}

			var size int
			if _, err := fmt.Sscanf(args, "%d", &size); err != nil {
				fmt.Println("Usage: fontsize SIZE (e.g., 24)")
				continue
			}

			sc.config.FontSize = size
			fmt.Printf("Font size set to %d\n", size)

		case "save":
			if args == "" {
				args = "shellcast_config.json"
			}

			if err := sc.config.SaveConfig(args); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			} else {
				fmt.Printf("Config saved to %s\n", args)
			}

		case "load":
			if args == "" {
				args = "shellcast_config.json"
			}

			config, err := LoadConfig(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			} else {
				sc.config = config
				fmt.Printf("Config loaded from %s\n", args)
			}

		default:
			if err := sc.ExecuteCommand(input); err != nil {
				fmt.Fprintf(os.Stderr, "Command error: %v\n", err)
			}
		}
	}
}

// showHelp displays available commands
func showHelp() {
	help := `
Available Commands:
------------------
help              Show this help message
exit, quit        Exit ShellCast
stream            Start streaming (prompts for RTMP URL if not set)
stop              Stop streaming
record            Start recording the session
stoprecord        Stop recording the session
theme [NAME]      List themes or apply a theme by name
timestamp [on|off] Enable or disable timestamps
size [WxH]        Show or set screen size (e.g., 1280x720)
split "cmd1" "cmd2" Run multiple commands in split screen mode
fontsize [SIZE]   Show or set font size
save [FILE]       Save configuration to a file
load [FILE]       Load configuration from a file

Any other input will be executed as a shell command.
`
	fmt.Println(help)
}
