package main

import (
        "flag"
        "fmt"
        "log"
        "os"
        "os/signal"
        "strings"
        "syscall"
        "time"
)

func main() {
        // Define command line flags
        rtmpUrl := flag.String("rtmp", "", "RTMP URL to stream to")
        ffmpegPath := flag.String("ffmpeg", "", "Path to FFmpeg executable")
        fontSize := flag.Int("font-size", 24, "Font size for streaming")
        fontColor := flag.String("font-color", "white", "Font color for streaming")
        bgColor := flag.String("bg-color", "black", "Background color for streaming")
        interactive := flag.Bool("interactive", false, "Run in interactive mode")
        configFile := flag.String("config", "", "Path to configuration file")
        showTimestamp := flag.Bool("timestamp", false, "Show timestamps in output")
        timestampFormat := flag.String("timestamp-format", "2006-01-02 15:04:05", "Format for timestamps")
        screenSize := flag.String("screen-size", "1280x720", "Screen size for streaming (WIDTHxHEIGHT)")
        record := flag.Bool("record", false, "Record session to file")
        recordPath := flag.String("record-path", "./recordings", "Directory to save recordings")
        themeName := flag.String("theme", "default", "Theme preset to use")
        splitMode := flag.Bool("split", false, "Run commands in split screen mode")
        listThemes := flag.Bool("list-themes", false, "List available theme presets")

        // 変数がどのフラグの状態を追跡するか保持するためのマップを作成
        flagsSet := make(map[string]bool)

        // フラグが設定されたかどうかを追跡するためのカスタムUsage関数
        oldUsage := flag.CommandLine.Usage
        flag.CommandLine.Usage = func() {
                oldUsage()
        }

        // Visitを使用して明示的に設定されたフラグを追跡
        visitor := func(f *flag.Flag) {
                flagsSet[f.Name] = true
        }

        flag.Parse()
        // 解析後、明示的に設定されたフラグを記録
        flag.Visit(visitor)

        // If list-themes flag is set, just list themes and exit
        if *listThemes {
                ListThemes()
                return
        }

        // Parse screen size
        var width, height int
        fmt.Sscanf(*screenSize, "%dx%d", &width, &height)
        if width <= 0 || height <= 0 {
                width, height = 1280, 720
        }

        // Create or load config
        var config Config
        var err error

        if *configFile != "" {
                config, err = LoadConfig(*configFile)
                if err != nil {
                        log.Printf("Error loading config, using defaults: %v", err)
                        config = GetDefaultConfig()
                }
        } else {
                config = GetDefaultConfig()
        }

        // Override config with command-line flags if provided
        if *rtmpUrl != "" {
                config.RTMPUrl = *rtmpUrl
        }
        if *ffmpegPath != "" {
                config.FFmpegPath = *ffmpegPath
        }
        if flagsSet["font-size"] {
                config.FontSize = *fontSize
        }
        if flagsSet["font-color"] {
                config.FontColor = *fontColor
        }
        if flagsSet["bg-color"] {
                config.BackgroundColor = *bgColor
        }
        if flagsSet["timestamp"] {
                config.ShowTimestamp = *showTimestamp
        }
        if flagsSet["timestamp-format"] {
                config.TimestampFormat = *timestampFormat
        }
        config.ScreenWidth = width
        config.ScreenHeight = height
        if flagsSet["record-path"] {
                config.RecordPath = *recordPath
        }
        if flagsSet["theme"] {
                config.ThemeName = *themeName
                config.ApplyTheme(*themeName)
        }

        // Create ShellCast instance
        shellcast := NewShellCast(config)

        // Set up signal handling for cleanup
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        go func() {
                <-sigChan
                fmt.Println("\nReceived termination signal. Cleaning up...")
                shellcast.Cleanup()
                os.Exit(0)
        }()

        // Check if a command was provided (non-flag arguments)
        args := flag.Args()
        hasCommand := len(args) > 0

        // Start recording if requested
        if *record {
                if err := shellcast.StartRecording(); err != nil {
                        log.Printf("Warning: Failed to start recording: %v", err)
                }
        }

        // Run in appropriate mode
        if *interactive {
                options := InteractiveOptions{
                        ConfigPath: *configFile,
                }
                RunInteractiveMode(shellcast, options)
        } else if *splitMode && hasCommand {
                // Split mode with multiple commands
                if err := shellcast.ExecuteSplitCommands(args); err != nil {
                        log.Fatalf("Error executing split commands: %v", err)
                }
        } else if hasCommand {
                command := strings.Join(args, " ")

                // Start streaming if RTMP URL is provided
                if config.RTMPUrl != "" {
                        if err := shellcast.StartStreaming(); err != nil {
                                log.Fatalf("Error starting stream: %v", err)
                        }
                        // Add delay to ensure streaming starts
                        time.Sleep(2 * time.Second)
                }

                // Execute the command
                if err := shellcast.ExecuteCommand(command); err != nil {
                        log.Printf("Command error: %v", err)
                }

                // If streaming, keep it running for a few seconds after command completes
                if shellcast.streaming {
                        fmt.Println("Command completed. Streaming for 5 more seconds...")
                        time.Sleep(5 * time.Second)
                        shellcast.StopStreaming()
                }
        } else {
                flag.Usage()
                fmt.Println("\nExamples:")
                fmt.Println("  shellcast -interactive")
                fmt.Println("  shellcast -rtmp rtmp://server/app ls -la")
                fmt.Println("  shellcast -theme hacker -timestamp on -record top")
                fmt.Println("  shellcast -split \"ls -la\" \"top -n 1\"")
        }

        // Clean up before exit
        shellcast.Cleanup()
}
