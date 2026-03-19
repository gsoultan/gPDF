package system

import (
	"gpdf/font"
	"gpdf/font/truetype"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FontScanner defines an interface for discovering font files on the system.
type FontScanner interface {
	// Scan returns a slice of absolute paths to discovered font files.
	Scan() ([]string, error)
	// RegisterAll scans the system and registers all discovered fonts into the provided registry.
	RegisterAll(r font.Registry) error
}

// GetDefaultScanner returns a FontScanner appropriate for the current operating system.
func GetDefaultScanner() FontScanner {
	switch runtime.GOOS {
	case "windows":
		return &WindowsScanner{}
	case "darwin":
		return &MacOSScanner{}
	default:
		return &LinuxScanner{}
	}
}

// WindowsScanner scans C:\Windows\Fonts for .ttf, .otf, and .ttc files.
type WindowsScanner struct{}

func (s *WindowsScanner) Scan() ([]string, error) {
	root := filepath.Join(os.Getenv("SystemRoot"), "Fonts")
	return scanDir(root)
}

func (s *WindowsScanner) RegisterAll(r font.Registry) error {
	paths, err := s.Scan()
	if err != nil {
		return err
	}
	registerPaths(r, paths)
	return nil
}

// MacOSScanner scans standard macOS font directories.
type MacOSScanner struct{}

func (s *MacOSScanner) Scan() ([]string, error) {
	dirs := []string{
		"/Library/Fonts",
		"/System/Library/Fonts",
		filepath.Join(os.Getenv("HOME"), "Library/Fonts"),
	}
	var all []string
	for _, d := range dirs {
		paths, _ := scanDir(d)
		all = append(all, paths...)
	}
	return all, nil
}

func (s *MacOSScanner) RegisterAll(r font.Registry) error {
	paths, err := s.Scan()
	if err != nil {
		return err
	}
	registerPaths(r, paths)
	return nil
}

// LinuxScanner scans standard Linux font directories.
type LinuxScanner struct{}

func (s *LinuxScanner) Scan() ([]string, error) {
	dirs := []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
		filepath.Join(os.Getenv("HOME"), ".local/share/fonts"),
		filepath.Join(os.Getenv("HOME"), ".fonts"),
	}
	var all []string
	for _, d := range dirs {
		paths, _ := scanDir(d)
		all = append(all, paths...)
	}
	return all, nil
}

func (s *LinuxScanner) RegisterAll(r font.Registry) error {
	paths, err := s.Scan()
	if err != nil {
		return err
	}
	registerPaths(r, paths)
	return nil
}

func registerPaths(r font.Registry, paths []string) {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		// Try collection first
		if fonts, err := truetype.ParseCollection(data); err == nil {
			for _, f := range fonts {
				r.Register(f)
			}
			continue
		}
		// Try single font
		if f, err := truetype.Parse(data); err == nil {
			r.Register(f)
		}
	}
}

func scanDir(root string) ([]string, error) {
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't access
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".ttf" || ext == ".otf" || ext == ".ttc" {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
