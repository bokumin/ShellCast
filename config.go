package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	RTMPUrl         string `json:"rtmp_url"`
	FFmpegPath      string `json:"ffmpeg_path"`
	FontSize        int    `json:"font_size"`
	FontColor       string `json:"font_color"`
	BackgroundColor string `json:"background_color"`
	OutputFile      string `json:"output_file"`
	
	// New customization options
	ShowTimestamp   bool   `json:"show_timestamp"`
	TimestampFormat string `json:"timestamp_format"`
	ScreenWidth     int    `json:"screen_width"`
	ScreenHeight    int    `json:"screen_height"`
	RecordSession   bool   `json:"record_session"`
	RecordPath      string `json:"record_path"`
	SplitScreen     bool   `json:"split_screen"`
	SplitCommands   []string `json:"split_commands"`
	ThemeName       string `json:"theme_name"`
}

// ThemePreset represents a predefined color scheme
type ThemePreset struct {
	Name           string `json:"name"`
	FontColor      string `json:"font_color"`
	BackgroundColor string `json:"background_color"`
	BorderColor    string `json:"border_color"`
	HighlightColor string `json:"highlight_color"`
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() Config {
	return Config{
		FFmpegPath:      "ffmpeg",
		FontSize:        24,
		FontColor:       "white",
		BackgroundColor: "black",
		TimestampFormat: "2006-01-02 15:04:05",
		ScreenWidth:     1280,
		ScreenHeight:    720,
		RecordPath:      "./recordings",
		ThemeName:       "default",
	}
}

// GetThemePresets returns predefined theme presets
func GetThemePresets() map[string]ThemePreset {
	return map[string]ThemePreset{
		"default": {
			Name:           "Default",
			FontColor:      "white",
			BackgroundColor: "black",
			BorderColor:    "gray",
			HighlightColor: "blue",
		},
		"hacker": {
			Name:           "Hacker",
			FontColor:      "lime",
			BackgroundColor: "black",
			BorderColor:    "green",
			HighlightColor: "red",
		},
		"solarized": {
			Name:           "Solarized",
			FontColor:      "#839496",
			BackgroundColor: "#002b36",
			BorderColor:    "#586e75",
			HighlightColor: "#268bd2",
		},
		"light": {
			Name:           "Light",
			FontColor:      "#222222",
			BackgroundColor: "#f9f9f9",
			BorderColor:    "#dddddd",
			HighlightColor: "#0066cc",
		},
		"monokai": {
			Name:           "Monokai",
			FontColor:      "#f8f8f2",
			BackgroundColor: "#272822",
			BorderColor:    "#75715e",
			HighlightColor: "#f92672",
		},
	}
}

// ApplyTheme applies a theme preset to the configuration
func (c *Config) ApplyTheme(themeName string) error {
	presets := GetThemePresets()
	theme, exists := presets[themeName]
	if !exists {
		return fmt.Errorf("theme '%s' not found", themeName)
	}
	
	c.ThemeName = themeName
	c.FontColor = theme.FontColor
	c.BackgroundColor = theme.BackgroundColor
	return nil
}

// SaveConfig saves the configuration to a file
func (c *Config) SaveConfig(filePath string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}
	
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}
	
	return nil
}

// LoadConfig loads the configuration from a file
func LoadConfig(filePath string) (Config, error) {
	config := GetDefaultConfig()
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return config, fmt.Errorf("error reading config file: %v", err)
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("error unmarshaling config: %v", err)
	}
	
	return config, nil
}

// ListThemes prints all available theme presets
func ListThemes() {
	presets := GetThemePresets()
	fmt.Println("Available themes:")
	for name, theme := range presets {
		fmt.Printf("- %s: Font: %s, Background: %s\n", 
			name, theme.FontColor, theme.BackgroundColor)
	}
}
