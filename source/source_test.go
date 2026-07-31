package source

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = utils.InitLogs()
	_ = Assets()
}

// setupTempDir creates a temporary directory for isolated testing
func setupTempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestAssets(t *testing.T) {
	err := Assets()
	assert.NoError(t, err)
	assert.NotNil(t, TmplBox)
}

func TestAssets_DisableHTTP(t *testing.T) {
	origValue := utils.Params.GetBool("DISABLE_HTTP")
	utils.Params.Set("DISABLE_HTTP", true)
	defer func() {
		utils.Params.Set("DISABLE_HTTP", origValue)
	}()

	err := Assets()
	assert.NoError(t, err)
}

func TestReadFile(t *testing.T) {
	data, err := ReadFile("robots.txt")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestReadFile_NotFound(t *testing.T) {
	_, err := ReadFile("nonexistent.txt")
	assert.Error(t, err)
}

func TestReadFileString(t *testing.T) {
	content, err := ReadFileString("robots.txt")
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "User-agent")
}

func TestUsingAssets_NoAssetsFolder(t *testing.T) {
	tmpDir := setupTempDir(t)

	result := UsingAssets(tmpDir)
	assert.False(t, result)
}

func TestUsingAssets_WithAssetsFolder(t *testing.T) {
	tmpDir := setupTempDir(t)

	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.Mkdir(assetsDir, 0o750)
	require.NoError(t, err)

	result := UsingAssets(tmpDir)
	assert.True(t, result)
}

func TestCustomCSS_FullCycle(t *testing.T) {
	tmpDir := setupTempDir(t)

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Initially no custom CSS
	assert.False(t, HasCustomCSS())

	// Load should return empty string when no file exists
	css, err := LoadCustomCSS()
	assert.NoError(t, err)
	assert.Empty(t, css)

	// Save custom CSS
	testCSS := "body { background: red; }"
	err = SaveCustomCSS(testCSS)
	assert.NoError(t, err)

	// Now HasCustomCSS should return true
	assert.True(t, HasCustomCSS())

	// Load should return the saved CSS
	css, err = LoadCustomCSS()
	assert.NoError(t, err)
	assert.Equal(t, testCSS, css)

	// CustomCSSPath should point to the correct file
	expectedPath := filepath.Join(tmpDir, "assets", "custom.css")
	assert.Equal(t, expectedPath, CustomCSSPath())
	assert.FileExists(t, expectedPath)

	// Delete custom CSS
	err = DeleteCustomCSS()
	assert.NoError(t, err)

	// Verify deleted
	assert.False(t, HasCustomCSS())
	css, err = LoadCustomCSS()
	assert.NoError(t, err)
	assert.Empty(t, css)
}

func TestSaveCustomCSS_CreatesAssetsDir(t *testing.T) {
	tmpDir := setupTempDir(t)

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Assets dir doesn't exist yet
	assetsDir := filepath.Join(tmpDir, "assets")
	assert.NoDirExists(t, assetsDir)

	// Save should create it
	err := SaveCustomCSS(".test { color: blue; }")
	assert.NoError(t, err)
	assert.DirExists(t, assetsDir)
}

func TestDeleteCustomCSS_NoFile(t *testing.T) {
	tmpDir := setupTempDir(t)

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Should not error when file doesn't exist
	err := DeleteCustomCSS()
	assert.NoError(t, err)
}
