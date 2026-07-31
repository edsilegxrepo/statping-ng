package source

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/statping-ng/statping-ng/utils"
)

//go:embed dist/*
var distFS embed.FS

var (
	log           = utils.Log.WithField("type", "source")
	TmplBox       fs.FS // Embedded files from 'source/dist' directory
	RequiredFiles = []string{
		"robots.txt",
		"base.gohtml",
	}
)

// Assets will initialize the embedded filesystem for templates and assets.
func Assets() error {
	if utils.Params.GetBool("DISABLE_HTTP") {
		return nil
	}
	var err error
	TmplBox, err = fs.Sub(distFS, "dist")
	if err != nil {
		return err
	}
	return nil
}

// ReadFile reads a file from the embedded filesystem
func ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(TmplBox, name)
}

// ReadFileString reads a file from the embedded filesystem as a string
func ReadFileString(name string) (string, error) {
	data, err := ReadFile(name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UsingAssets returns true if the '/assets' folder is found in the directory
func UsingAssets(folder string) bool {
	if _, err := os.Stat(filepath.Join(folder, "assets")); err == nil {
		return true
	}
	return false
}

// CustomCSSPath returns the path to the custom CSS file
func CustomCSSPath() string {
	return filepath.Join(utils.Directory, "assets", "custom.css")
}

// HasCustomCSS returns true if a custom.css file exists
func HasCustomCSS() bool {
	_, err := os.Stat(CustomCSSPath())
	return err == nil
}

// SaveCustomCSS saves custom CSS to the assets folder
func SaveCustomCSS(css string) error {
	assetsDir := filepath.Join(utils.Directory, "assets")
	if !utils.FolderExists(assetsDir) {
		if err := utils.CreateDirectory(assetsDir); err != nil {
			return fmt.Errorf("failed to create assets directory: %w", err)
		}
	}
	return utils.SaveFile(CustomCSSPath(), []byte(css))
}

// LoadCustomCSS reads the custom CSS file
func LoadCustomCSS() (string, error) {
	if !HasCustomCSS() {
		return "", nil
	}
	return utils.OpenFile(CustomCSSPath())
}

// DeleteCustomCSS removes the custom CSS file
func DeleteCustomCSS() error {
	if !HasCustomCSS() {
		return nil
	}
	return os.Remove(CustomCSSPath())
}
