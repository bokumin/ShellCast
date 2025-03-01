package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ShellCast is the main application structure
type ShellCast struct {
	config       Config
	outputBuffer string
	mutex        sync.Mutex
	streaming    bool
	streamProc   *os.Process
	recording    bool
	recordPath   string
	startTime    time.Time
}

// NewShellCast creates a new ShellCast instance
func NewShellCast(config Config) *ShellCast {
	return &ShellCast{
		config:     config,
		streaming:  false,
		recording:  false,
		streamProc: nil,
		startTime:  time.Now(),
	}
}

// ExecuteCommand runs a shell command and captures its output
func (s *ShellCast) ExecuteCommand(command string) error {
	// Split the command string into parts
	parts := strings.Split(command, " ")
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Create the command
	cmd := exec.Command(parts[0], parts[1:]...)

	// Get pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}

	// Handle output in goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Process stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			formattedLine := s.formatOutput(line)
			fmt.Println(formattedLine)

			// Store in buffer
			s.mutex.Lock()
			s.outputBuffer += formattedLine + "\n"
			s.mutex.Unlock()

			// If streaming, append to output file
			if s.streaming && s.config.OutputFile != "" {
				appendToFile(s.config.OutputFile, formattedLine+"\n")
			}

			// If recording, save to record file
			if s.recording && s.recordPath != "" {
				appendToFile(s.recordPath, formattedLine+"\n")
			}
		}
	}()

	// Process stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			formattedLine := s.formatOutput(line)
			fmt.Fprintln(os.Stderr, formattedLine)

			// Store in buffer
			s.mutex.Lock()
			s.outputBuffer += formattedLine + "\n"
			s.mutex.Unlock()

			// If streaming, append to output file
			if s.streaming && s.config.OutputFile != "" {
				appendToFile(s.config.OutputFile, formattedLine+"\n")
			}

			// If recording, save to record file
			if s.recording && s.recordPath != "" {
				appendToFile(s.recordPath, formattedLine+"\n")
			}
		}
	}()

	// Wait for command to finish
	wg.Wait()
	return cmd.Wait()
}

// formatOutput adds timestamp and other formatting to the output
func (s *ShellCast) formatOutput(line string) string {
	if s.config.ShowTimestamp {
		timestamp := time.Now().Format(s.config.TimestampFormat)
		return fmt.Sprintf("[%s] %s", timestamp, line)
	}
	return line
}

// StartStreaming starts the FFmpeg process to stream terminal output
func (s *ShellCast) StartStreaming() error {
	if s.streaming {
		return fmt.Errorf("already streaming")
	}

	// Create output file if it doesn't exist
	if s.config.OutputFile == "" {
		tmpFile, err := os.CreateTemp("", "shellcast_*.txt")
		if err != nil {
			return fmt.Errorf("error creating temp file: %v", err)
		}
		s.config.OutputFile = tmpFile.Name()
		tmpFile.Close()
	}

	// Write current buffer to file
	s.mutex.Lock()
	err := os.WriteFile(s.config.OutputFile, []byte(s.outputBuffer), 0644)
	s.mutex.Unlock()
	if err != nil {
		return fmt.Errorf("error writing to output file: %v", err)
	}

	// Prepare FFmpeg command
	ffmpegPath := s.config.FFmpegPath
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg" // Use from PATH
	}

	// Create complex filter for custom formatting
	vfFilter := s.createVideoFilter()

	args := []string{
		"-re",
		"-f", "concat",
		"-safe", "0",
		"-i", s.config.OutputFile,
		"-vf", vfFilter,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-s", fmt.Sprintf("%dx%d", s.config.ScreenWidth, s.config.ScreenHeight),
		"-f", "flv",
		s.config.RTMPUrl,
	}

	cmd := exec.Command(ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting FFmpeg: %v", err)
	}

	s.streamProc = cmd.Process
	s.streaming = true

	fmt.Printf("Streaming started to %s\n", s.config.RTMPUrl)
	return nil
}

// createVideoFilter creates the FFmpeg video filter string
func (s *ShellCast) createVideoFilter() string {
	// Basic text display
	filter := fmt.Sprintf("drawtext=fontfile=/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf:fontcolor=%s:fontsize=%d:box=1:boxcolor=%s:x=20:y=20:text='%s'",
		s.config.FontColor,
		s.config.FontSize,
		s.config.BackgroundColor,
		"%{eif\\:n\\:d}") // Line number will be added by FFmpeg

	// Add timestamp if requested
	if s.config.ShowTimestamp {
		filter += ",drawtext=fontfile=/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf:" +
			fmt.Sprintf("fontcolor=%s:fontsize=%d:box=1:boxcolor=%s:x=w-200:y=20:text='%%{localtime}'",
				s.config.FontColor, s.config.FontSize, s.config.BackgroundColor)
	}

	return filter
}

// StopStreaming stops the streaming process
func (s *ShellCast) StopStreaming() error {
	if !s.streaming || s.streamProc == nil {
		return fmt.Errorf("not streaming")
	}

	// Kill FFmpeg process
	if err := s.streamProc.Kill(); err != nil {
		return fmt.Errorf("error killing FFmpeg process: %v", err)
	}

	s.streaming = false
	s.streamProc = nil

	// Clean up output file
	if s.config.OutputFile != "" {
		os.Remove(s.config.OutputFile)
		s.config.OutputFile = ""
	}

	fmt.Println("Streaming stopped")
	return nil
}

// StartRecording starts recording the session to a file
func (s *ShellCast) StartRecording() error {
	if s.recording {
		return fmt.Errorf("already recording")
	}

	// Create recordings directory if it doesn't exist
	if _, err := os.Stat(s.config.RecordPath); os.IsNotExist(err) {
		if err := os.MkdirAll(s.config.RecordPath, 0755); err != nil {
			return fmt.Errorf("error creating recordings directory: %v", err)
		}
	}

	// Generate record filename based on timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("shellcast_%s.txt", timestamp)
	s.recordPath = filepath.Join(s.config.RecordPath, filename)

	// Write header to recording file
	header := fmt.Sprintf("ShellCast Recording - Started at %s\n", 
		time.Now().Format(s.config.TimestampFormat))
	header += fmt.Sprintf("Command: %s\n", strings.Join(os.Args, " "))
	header += strings.Repeat("-", 80) + "\n\n"

	if err := os.WriteFile(s.recordPath, []byte(header), 0644); err != nil {
		return fmt.Errorf("error writing to record file: %v", err)
	}

	s.recording = true
	fmt.Printf("Recording started: %s\n", s.recordPath)
	return nil
}

// StopRecording stops the recording process
func (s *ShellCast) StopRecording() error {
	if !s.recording {
		return fmt.Errorf("not recording")
	}

	// Write footer to recording file
	footer := fmt.Sprintf("\n\n%s\n", strings.Repeat("-", 80))
	footer += fmt.Sprintf("Recording ended at %s\n", 
		time.Now().Format(s.config.TimestampFormat))
	footer += fmt.Sprintf("Duration: %s\n", time.Since(s.startTime).Round(time.Second))

	if err := appendToFile(s.recordPath, footer); err != nil {
		return fmt.Errorf("error writing to record file: %v", err)
	}

	s.recording = false
	fmt.Printf("Recording stopped: %s\n", s.recordPath)
	return nil
}

// ExecuteSplitCommands executes multiple commands in a split screen view
func (s *ShellCast) ExecuteSplitCommands(commands []string) error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands provided for split screen")
	}

	// Create a wait group for all commands
	var wg sync.WaitGroup
	wg.Add(len(commands))

	// Execute each command in a separate goroutine
	for i, cmd := range commands {
		go func(idx int, command string) {
			defer wg.Done()
			
			// Create a prefix for this command output
			prefix := fmt.Sprintf("[CMD%d] ", idx+1)
			
			parts := strings.Split(command, " ")
			if len(parts) == 0 {
				fmt.Printf("%sEmpty command\n", prefix)
				return
			}
			
			// Create and execute the command
			cmd := exec.Command(parts[0], parts[1:]...)
			
			// Get pipes for stdout and stderr
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%sError creating stdout pipe: %v\n", prefix, err)
				return
			}
			
			stderr, err := cmd.StderrPipe()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%sError creating stderr pipe: %v\n", prefix, err)
				return
			}
			
			// Start the command
			if err := cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "%sError starting command: %v\n", prefix, err)
				return
			}
			
			// Process stdout
			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					line := scanner.Text()
					formattedLine := s.formatOutput(prefix + line)
					fmt.Println(formattedLine)
					
					// Add to buffer and recording if active
					s.mutex.Lock()
					s.outputBuffer += formattedLine + "\n"
					s.mutex.Unlock()
					
					if s.streaming && s.config.OutputFile != "" {
						appendToFile(s.config.OutputFile, formattedLine+"\n")
					}
					
					if s.recording && s.recordPath != "" {
						appendToFile(s.recordPath, formattedLine+"\n")
					}
				}
			}()
			
			// Process stderr
			go func() {
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					line := scanner.Text()
					formattedLine := s.formatOutput(prefix + line)
					fmt.Fprintln(os.Stderr, formattedLine)
					
					// Add to buffer and recording if active
					s.mutex.Lock()
					s.outputBuffer += formattedLine + "\n"
					s.mutex.Unlock()
					
					if s.streaming && s.config.OutputFile != "" {
						appendToFile(s.config.OutputFile, formattedLine+"\n")
					}
					
					if s.recording && s.recordPath != "" {
						appendToFile(s.recordPath, formattedLine+"\n")
					}
				}
			}()
			
			// Wait for command to finish
			cmd.Wait()
			fmt.Printf("%sCommand completed\n", prefix)
		}(i, cmd)
	}

	// Wait for all commands to complete
	wg.Wait()
	return nil
}

// Cleanup performs cleanup operations
func (s *ShellCast) Cleanup() {
	if s.streaming {
		s.StopStreaming()
	}
	
	if s.recording {
		s.StopRecording()
	}
}

// Helper function to append text to a file
func appendToFile(filename, text string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}
