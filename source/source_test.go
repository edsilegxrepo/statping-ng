package source

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dir string

func init() {
	dir = utils.Directory
	_ = utils.InitLogs()
	_ = Assets()
	_ = utils.DeleteDirectory(dir + "/assets")
	dir = utils.Params.GetString("STATPING_DIR")
}

// setupTempDir creates a temporary directory for isolated testing and returns
// the directory path and a cleanup function.
func setupTempDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "source_test_*")
	require.NoError(t, err)
	return tmpDir, func() {
		_ = os.RemoveAll(tmpDir)
	}
}

// TestUsingAssets_NoAssetsFolder tests UsingAssets returns false when no assets folder exists
func TestUsingAssets_NoAssetsFolder(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	result := UsingAssets(tmpDir)
	assert.False(t, result, "UsingAssets should return false when assets folder does not exist")
}

// TestUsingAssets_WithAssetsFolder tests UsingAssets returns true when assets folder exists
func TestUsingAssets_WithAssetsFolder(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	// Create the assets folder
	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.Mkdir(assetsDir, 0o750)
	require.NoError(t, err)

	result := UsingAssets(tmpDir)
	assert.True(t, result, "UsingAssets should return true when assets folder exists")
}

// TestMakePublicFolder_NewFolder tests creating a new folder
func TestMakePublicFolder_NewFolder(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	newFolder := filepath.Join(tmpDir, "newfolder")
	err := MakePublicFolder(newFolder)
	assert.NoError(t, err)
	assert.DirExists(t, newFolder)
}

// TestMakePublicFolder_ExistingFolder tests MakePublicFolder with existing folder (should not error)
func TestMakePublicFolder_ExistingFolder(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	existingFolder := filepath.Join(tmpDir, "existing")
	err := os.Mkdir(existingFolder, 0o750)
	require.NoError(t, err)

	// Should not error when folder already exists
	err = MakePublicFolder(existingFolder)
	assert.NoError(t, err)
	assert.DirExists(t, existingFolder)
}

// TestMakePublicFolder_NestedPath tests creating nested folders
func TestMakePublicFolder_NestedPath(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	// Create parent first
	parent := filepath.Join(tmpDir, "parent")
	err := MakePublicFolder(parent)
	require.NoError(t, err)

	// Create child
	child := filepath.Join(parent, "child")
	err = MakePublicFolder(child)
	assert.NoError(t, err)
	assert.DirExists(t, child)
}

// TestSaveAsset_Isolated tests SaveAsset in an isolated temp directory
func TestSaveAsset_Isolated(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	// Temporarily override utils.Directory
	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Create assets directory structure
	assetsDir := filepath.Join(tmpDir, "assets", "js")
	err := os.MkdirAll(assetsDir, 0o750)
	require.NoError(t, err)

	testData := []byte("console.log('test');")
	err = SaveAsset(testData, "js/test.js")
	assert.NoError(t, err)

	expectedPath := filepath.Join(tmpDir, "assets", "js", "test.js")
	assert.FileExists(t, expectedPath)

	// Verify content
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, testData, content)
}

// TestSaveAsset_NoDirectory tests SaveAsset when directory doesn't exist
func TestSaveAsset_NoDirectory(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Don't create directory - SaveAsset should fail
	testData := []byte("test")
	err := SaveAsset(testData, "nonexistent/test.txt")
	assert.Error(t, err, "SaveAsset should fail when directory doesn't exist")
}

// TestOpenAsset_Isolated tests OpenAsset in an isolated temp directory
func TestOpenAsset_Isolated(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Create assets and write a file
	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.MkdirAll(assetsDir, 0o750)
	require.NoError(t, err)

	testContent := "test content here"
	testFile := filepath.Join(assetsDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0o600)
	require.NoError(t, err)

	result := OpenAsset("test.txt")
	assert.Equal(t, testContent, result)
}

// TestOpenAsset_NonExistent tests OpenAsset with non-existent file
func TestOpenAsset_NonExistent(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Don't create any file - OpenAsset should return empty string
	result := OpenAsset("nonexistent.txt")
	assert.Empty(t, result, "OpenAsset should return empty string for non-existent file")
}

// TestDeleteAllAssets_Isolated tests DeleteAllAssets in an isolated temp directory
func TestDeleteAllAssets_Isolated(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	// Create assets folder with some content
	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.MkdirAll(filepath.Join(assetsDir, "css"), 0o750)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(assetsDir, "test.txt"), []byte("test"), 0o600)
	require.NoError(t, err)

	// Verify assets folder exists
	assert.DirExists(t, assetsDir)

	// Delete all assets
	err = DeleteAllAssets(tmpDir)
	assert.NoError(t, err)

	// Verify assets folder is gone
	assert.NoDirExists(t, assetsDir)
}

// TestDeleteAllAssets_NoAssetsFolder tests DeleteAllAssets when no assets folder exists
func TestDeleteAllAssets_NoAssetsFolder(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	// Don't create assets folder - should not error (os.RemoveAll handles this)
	err := DeleteAllAssets(tmpDir)
	assert.NoError(t, err)
}

// TestCopyToPublic_Isolated tests CopyToPublic with rice box in isolated directory
func TestCopyToPublic_Isolated(t *testing.T) {
	// Skip if TmplBox is nil (rice box not loaded)
	if TmplBox == nil {
		t.Skip("TmplBox not loaded, skipping CopyToPublic test")
	}

	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Create assets folder
	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.MkdirAll(assetsDir, 0o750)
	require.NoError(t, err)

	// Copy robots.txt (no subpath)
	err = CopyToPublic(TmplBox, "", "robots.txt")
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(assetsDir, "robots.txt"))
}

// TestCopyToPublic_WithSubpath tests CopyToPublic with a subpath
func TestCopyToPublic_WithSubpath(t *testing.T) {
	if TmplBox == nil {
		t.Skip("TmplBox not loaded, skipping CopyToPublic test")
	}

	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Create assets/scss folder
	scssDir := filepath.Join(tmpDir, "assets", "scss")
	err := os.MkdirAll(scssDir, 0o750)
	require.NoError(t, err)

	// CopyToPublic uses box.String(file) where file is the path in the rice box.
	// The rice box stores files with paths like "scss/index.scss" relative to dist.
	// When path is provided, it becomes part of the output path: assets/{path}/{file}
	err = CopyToPublic(TmplBox, "", "scss/index.scss")
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(tmpDir, "assets", "scss", "index.scss"))
}

// TestCopyToPublic_NonExistentFile tests CopyToPublic with non-existent file
func TestCopyToPublic_NonExistentFile(t *testing.T) {
	if TmplBox == nil {
		t.Skip("TmplBox not loaded, skipping CopyToPublic test")
	}

	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Create assets folder
	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.MkdirAll(assetsDir, 0o750)
	require.NoError(t, err)

	// Try to copy non-existent file
	err = CopyToPublic(TmplBox, "", "nonexistent_file_xyz.txt")
	assert.Error(t, err, "CopyToPublic should fail for non-existent file")
}

// TestScssRendered tests the scssRendered helper function
func TestScssRendered(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard scss path",
			input:    filepath.Join("assets", "scss", "index.scss"),
			expected: filepath.Join("assets", "css", "index.css"),
		},
		{
			name:     "nested scss path",
			input:    filepath.Join("home", "user", "assets", "scss", "base.scss"),
			expected: filepath.Join("home", "user", "assets", "css", "base.css"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scssRendered(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSaveAsset_OverwriteExisting tests that SaveAsset overwrites existing files
func TestSaveAsset_OverwriteExisting(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Create assets directory
	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.MkdirAll(assetsDir, 0o750)
	require.NoError(t, err)

	// Write initial content
	err = SaveAsset([]byte("initial"), "test.txt")
	require.NoError(t, err)

	// Overwrite with new content
	err = SaveAsset([]byte("updated"), "test.txt")
	require.NoError(t, err)

	// Verify content was overwritten
	result := OpenAsset("test.txt")
	assert.Equal(t, "updated", result)
}

// TestOpenAsset_EmptyFile tests OpenAsset with an empty file
func TestOpenAsset_EmptyFile(t *testing.T) {
	tmpDir, cleanup := setupTempDir(t)
	defer cleanup()

	origDir := utils.Directory
	utils.Directory = tmpDir
	defer func() { utils.Directory = origDir }()

	// Create assets directory and empty file
	assetsDir := filepath.Join(tmpDir, "assets")
	err := os.MkdirAll(assetsDir, 0o750)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(assetsDir, "empty.txt"), []byte{}, 0o600)
	require.NoError(t, err)

	result := OpenAsset("empty.txt")
	assert.Empty(t, result, "OpenAsset should return empty string for empty file")
}

// TestAssets_DisableHTTP tests that Assets returns nil when DISABLE_HTTP is true
func TestAssets_DisableHTTP(t *testing.T) {
	// Save original value
	origValue := utils.Params.GetBool("DISABLE_HTTP")

	// Set DISABLE_HTTP to true
	utils.Params.Set("DISABLE_HTTP", true)
	defer func() {
		utils.Params.Set("DISABLE_HTTP", origValue)
	}()

	err := Assets()
	assert.NoError(t, err, "Assets should return nil when DISABLE_HTTP is true")
}

func assertFiles(t *testing.T, exist bool) {
	for _, f := range RequiredFiles {
		if exist {
			assert.FileExists(t, dir+"/assets/"+f)
		} else {
			assert.NoFileExists(t, dir+"/assets/"+f)
		}
	}
}

func TestCore_UsingAssets(t *testing.T) {
	assert.False(t, UsingAssets(dir))
	assertFiles(t, false)
}

func TestCreateAssets(t *testing.T) {
	assert.Nil(t, CreateAllAssets(dir))
	assert.True(t, UsingAssets(dir))
	assert.Nil(t, CompileSASS())
	assertFiles(t, true)
}

func TestCopyAllToPublic(t *testing.T) {
	err := CopyAllToPublic(TmplBox)
	require.Nil(t, err)
	assertFiles(t, true)
}

func TestCompileSASS(t *testing.T) {
	err := CompileSASS()
	require.Nil(t, err)
	assert.True(t, UsingAssets(dir))
	assertFiles(t, true)
}

func TestSaveAndCompileAsset(t *testing.T) {
	vars := OpenAsset("scss/variables.scss")
	vars += "$testingcolor: #b1b2b3;"

	err := SaveAsset([]byte(vars), "scss/variables.scss")
	require.Nil(t, err)
	assert.FileExists(t, dir+"/assets/scss/variables.scss")

	scssData := OpenAsset("scss/base.scss")
	scssData += "BODY { color: $testingcolor; }"
	err = SaveAsset([]byte(scssData), "scss/base.scss")
	require.Nil(t, err)
	assert.FileExists(t, dir+"/assets/scss/base.scss")

	asset := OpenAsset("scss/variables.scss")
	assert.NotEmpty(t, asset)
	assert.Equal(t, vars, asset)

	asset = OpenAsset("scss/base.scss")
	assert.NotEmpty(t, asset)
	assert.Equal(t, scssData, asset)

	err = CompileSASS()
	require.Nil(t, err)
	assertFiles(t, true)

	themeCSS, err := utils.OpenFile(dir + "/assets/css/index.css")
	require.Nil(t, err)

	assert.Contains(t, themeCSS, `color: #b1b2b3;`)
}

func TestOpenAsset(t *testing.T) {
	for _, f := range RequiredFiles {
		assert.FileExists(t, dir+"/assets/"+f)
		assert.NotEmpty(t, OpenAsset(f))
	}
}

func TestDeleteAssets(t *testing.T) {
	assert.True(t, UsingAssets(dir))
	assert.Nil(t, DeleteAllAssets(dir))
	assert.False(t, UsingAssets(dir))
	assertFiles(t, false)
}

func ExampleSaveAsset() {
	data := []byte("alert('helloooo')")
	_ = SaveAsset(data, "js/test.js")
}

func ExampleOpenAsset() {
	OpenAsset("js/main.js")
}
