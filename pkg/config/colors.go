package config

import (
	"os"
	"path/filepath"

	"github.com/r2unit/openpasswd/pkg/toml"
)

type ColorScheme struct {
	SearchBar    string `toml:"search_bar"`
	Sidebar      string `toml:"sidebar"`
	MainBox      string `toml:"main_box"`
	Header       string `toml:"header"`
	Selected     string `toml:"selected"`
	SelectedBg   string `toml:"selected_bg"`
	Normal       string `toml:"normal"`
	SidebarTitle string `toml:"sidebar_title"`
	Help         string `toml:"help"`
	Version      string `toml:"version"`
	Error        string `toml:"error"`
	Success      string `toml:"success"`
}

type UIConfig struct {
	Colors ColorScheme `toml:"colors"`
}

var DefaultColorScheme = ColorScheme{
	SearchBar:    "#5FAFFF",
	Sidebar:      "#585858",
	MainBox:      "#585858",
	Header:       "#FFFFFF",
	Selected:     "#FFFFFF",
	SelectedBg:   "#5FAFFF",
	Normal:       "#A8A8A8",
	SidebarTitle: "#FFAF00",
	Help:         "#585858",
	Version:      "#3A3A3A",
	Error:        "#FF5F5F",
	Success:      "#00D75F",
}

func LoadColorScheme() (*ColorScheme, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return &DefaultColorScheme, nil
	}

	configPath := filepath.Join(configDir, "config.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &DefaultColorScheme, nil
	}

	var uiConfig UIConfig
	if _, err := toml.DecodeFile(configPath, &uiConfig); err != nil {
		return &DefaultColorScheme, nil
	}

	scheme := &uiConfig.Colors

	if scheme.SearchBar == "" {
		scheme.SearchBar = DefaultColorScheme.SearchBar
	}
	if scheme.Sidebar == "" {
		scheme.Sidebar = DefaultColorScheme.Sidebar
	}
	if scheme.MainBox == "" {
		scheme.MainBox = DefaultColorScheme.MainBox
	}
	if scheme.Header == "" {
		scheme.Header = DefaultColorScheme.Header
	}
	if scheme.Selected == "" {
		scheme.Selected = DefaultColorScheme.Selected
	}
	if scheme.SelectedBg == "" {
		scheme.SelectedBg = DefaultColorScheme.SelectedBg
	}
	if scheme.Normal == "" {
		scheme.Normal = DefaultColorScheme.Normal
	}
	if scheme.SidebarTitle == "" {
		scheme.SidebarTitle = DefaultColorScheme.SidebarTitle
	}
	if scheme.Help == "" {
		scheme.Help = DefaultColorScheme.Help
	}
	if scheme.Version == "" {
		scheme.Version = DefaultColorScheme.Version
	}
	if scheme.Error == "" {
		scheme.Error = DefaultColorScheme.Error
	}
	if scheme.Success == "" {
		scheme.Success = DefaultColorScheme.Success
	}

	return scheme, nil
}

func CreateDefaultConfig() error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.toml")

	if _, err := os.Stat(configPath); err == nil {
		return nil
	}

	defaultConfig := `# OpenPasswd Configuration File
# You can customize the color scheme and keybindings here

[colors]
# Colors use hex format: #RRGGBB
# Search bar border color
search_bar = "#5FAFFF"

# Sidebar border color
sidebar = "#585858"

# Main password box border color
main_box = "#585858"

# Header text color
header = "#FFFFFF"

# Selected item text color
selected = "#FFFFFF"

# Selected item background color
selected_bg = "#5FAFFF"

# Normal text color
normal = "#A8A8A8"

# Sidebar section title color (Recent/Common)
sidebar_title = "#FFAF00"

# Help text color
help = "#585858"

# Version text color
version = "#3A3A3A"

# Error message color
error = "#FF5F5F"

# Success message color
success = "#00D75F"

[keybindings]
# Quit the application (nvim-style)
quit = ":q"

# Alternative quit command (Ctrl+C)
quit_alt = "ctrl+c"

# Go back / escape current view
back = "esc"

# Navigate up
up = "up"
up_alt = "k"

# Navigate down
down = "down"
down_alt = "j"

# Select / confirm
select = "enter"
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0600)
}
