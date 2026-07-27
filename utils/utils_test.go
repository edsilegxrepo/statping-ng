package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLog(t *testing.T) {
	if Directory == "" {
		Directory, _ = os.Getwd()
	}
	err := createLog(Directory)
	assert.Nil(t, err)
}

func TestReplaceValue(t *testing.T) {
	assert.Equal(t, true, replaceVal(true))
	assert.Equal(t, 42, replaceVal(42))
	assert.Equal(t, "hello world", replaceVal("hello world"))
	assert.Equal(t, "5s", replaceVal(5*time.Second))
}

func TestReplaceValComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		// Basic types passed through
		{"bool true", true, true},
		{"bool false", false, false},
		{"int", 42, 42},
		{"float64", 3.14, 3.14},

		// Short strings passed through
		{"short string", "hello", "hello"},
		{"empty string", "", ""},

		// Long strings truncated
		{"500 char string", string(make([]byte, 500)), string(make([]byte, 500))},

		// time.Time converted to string
		{"time.Duration", 5 * time.Second, "5s"},
		{"time.Duration hours", 2 * time.Hour, "2h0m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceVal(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test long string truncation separately
	t.Run("long string truncated", func(t *testing.T) {
		longStr := string(make([]byte, 600))
		for i := range longStr {
			longStr = longStr[:i] + "a" + longStr[i+1:]
		}
		// Create a 600 character string
		input := make([]byte, 600)
		for i := range input {
			input[i] = 'a'
		}
		result := replaceVal(string(input)).(string)
		assert.Len(t, result, 500+len("... (truncated in logs)"))
		assert.Contains(t, result, "... (truncated in logs)")
	})

	// Test time.Time
	t.Run("time.Time", func(t *testing.T) {
		now := time.Now()
		result := replaceVal(now)
		assert.Equal(t, now.String(), result)
	})
}

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		expected  bool
	}{
		{"password", "password", true},
		{"Password uppercase", "Password", true},
		{"api_key", "api_key", true},
		{"ApiKey camel", "ApiKey", true},
		{"secret", "secret", true},
		{"token", "token", true},
		{"credential", "credential", true},
		{"pass in name", "userpass", true},
		{"regular field", "username", false},
		{"email field", "email", false},
		{"id field", "id", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveField(tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogRow(t *testing.T) {
	t.Run("newLogRow creates row with current time", func(t *testing.T) {
		before := time.Now()
		row := newLogRow("test message")
		after := time.Now()

		assert.NotNil(t, row)
		assert.Equal(t, "test message", row.Line)
		assert.True(t, row.Date.After(before) || row.Date.Equal(before))
		assert.True(t, row.Date.Before(after) || row.Date.Equal(after))
	})

	t.Run("lineAsString with string", func(t *testing.T) {
		row := newLogRow("test message")
		assert.Equal(t, "test message", row.lineAsString())
	})

	t.Run("lineAsString with error", func(t *testing.T) {
		row := newLogRow(fmt.Errorf("test error"))
		assert.Equal(t, "test error", row.lineAsString())
	})

	t.Run("lineAsString with bytes", func(t *testing.T) {
		row := newLogRow([]byte("byte message"))
		assert.Equal(t, "byte message", row.lineAsString())
	})

	t.Run("lineAsString with other type", func(t *testing.T) {
		row := newLogRow(12345)
		assert.Equal(t, "", row.lineAsString())
	})

	t.Run("FormatForHtml", func(t *testing.T) {
		row := newLogRow("test message")
		formatted := row.FormatForHtml()
		assert.Contains(t, formatted, "test message")
		assert.Contains(t, formatted, ":")
		// Check date format pattern
		assert.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`, formatted)
	})
}

func TestPushAndGetLastLine(t *testing.T) {
	// Clear existing lines
	LockLines.Lock()
	LastLines = nil
	LockLines.Unlock()

	t.Run("GetLastLine returns nil when empty", func(t *testing.T) {
		LockLines.Lock()
		LastLines = nil
		LockLines.Unlock()
		result := GetLastLine()
		assert.Nil(t, result)
	})

	t.Run("pushLastLine adds line", func(t *testing.T) {
		LockLines.Lock()
		LastLines = nil
		LockLines.Unlock()

		pushLastLine("first line")
		result := GetLastLine()
		assert.NotNil(t, result)
		assert.Equal(t, "first line", result.Line)
	})

	t.Run("GetLastLine returns most recent", func(t *testing.T) {
		LockLines.Lock()
		LastLines = nil
		LockLines.Unlock()

		pushLastLine("line 1")
		pushLastLine("line 2")
		pushLastLine("line 3")

		result := GetLastLine()
		assert.Equal(t, "line 3", result.Line)
	})

	t.Run("pushLastLine respects 1000 line limit", func(t *testing.T) {
		LockLines.Lock()
		LastLines = nil
		LockLines.Unlock()

		// Add 1050 lines
		for i := 0; i < 1050; i++ {
			pushLastLine(fmt.Sprintf("line %d", i))
		}

		LockLines.Lock()
		lineCount := len(LastLines)
		firstLine := LastLines[0].Line
		LockLines.Unlock()

		assert.Equal(t, 1000, lineCount)
		// First line should be "line 50" (oldest 50 were removed)
		assert.Equal(t, "line 50", firstLine)
	})
}

func TestCheckVerboseMode(t *testing.T) {
	// Save original value
	originalMode := VerboseMode
	defer func() { VerboseMode = originalMode }()

	tests := []struct {
		mode          int
		expectedLevel string
	}{
		{1, "warning"},
		{2, "info"},
		{3, "debug"},
		{4, "trace"},
		{0, "info"},  // default
		{99, "info"}, // unknown defaults to info
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("mode_%d", tt.mode), func(t *testing.T) {
			VerboseMode = tt.mode
			checkVerboseMode()
			assert.Equal(t, tt.expectedLevel, Log.GetLevel().String())
		})
	}
}

func TestInitLogs(t *testing.T) {
	assert.Nil(t, InitLogs())
	require.NotEmpty(t, Params.GetString("STATPING_DIR"))
	require.False(t, Params.GetBool("DISABLE_LOGS"))

	Log.Infoln("this is a test")
	assert.FileExists(t, Directory+"/logs/statping.log")
}

func TestDir(t *testing.T) {
	assert.Contains(t, Directory, "statping-ng")
}

func TestCommand(t *testing.T) {
	t.SkipNow()
	_, out, err := Command("/bin/echo", "\"statping testing\"")
	assert.Nil(t, err)
	assert.Contains(t, out, "statping")
}

func TestCommandCrossplatform(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping command tests in CI environment")
	}

	t.Run("simple echo command", func(t *testing.T) {
		var stdout, stderr string
		var err error

		if os.PathSeparator == '\\' {
			// Windows
			stdout, stderr, err = Command("cmd", "/c", "echo", "hello")
		} else {
			// Unix
			stdout, stderr, err = Command("echo", "hello")
		}

		assert.NoError(t, err)
		assert.Contains(t, stdout, "hello")
		_ = stderr // may or may not have content
	})

	t.Run("command not found", func(t *testing.T) {
		_, _, err := Command("nonexistent_command_12345")
		assert.Error(t, err)
	})

	t.Run("command with exit error", func(t *testing.T) {
		var err error

		if os.PathSeparator == '\\' {
			// Windows - exit with code 1
			_, _, err = Command("cmd", "/c", "exit", "1")
		} else {
			// Unix
			_, _, err = Command("false")
		}

		assert.Error(t, err)
	})
}

func TestToInt(t *testing.T) {
	assert.Equal(t, int64(55), ToInt("55"))
	assert.Equal(t, int64(55), ToInt(55))
	assert.Equal(t, int64(55), ToInt(55.0))
	assert.Equal(t, int64(55), ToInt([]byte("55")))
}

func TestToIntComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		// String conversions
		{"string positive", "123", 123},
		{"string negative", "-456", -456},
		{"string zero", "0", 0},
		{"string empty", "", 0},
		{"string invalid", "abc", 0},
		{"string with spaces", " 42 ", 0}, // strconv.Atoi fails on leading/trailing spaces

		// Byte slice conversions
		{"bytes positive", []byte("789"), 789},
		{"bytes negative", []byte("-100"), -100},
		{"bytes empty", []byte(""), 0},

		// Float conversions
		{"float32 positive", float32(42.7), 42},
		{"float32 negative", float32(-42.7), -42},
		{"float64 positive", float64(99.9), 99},
		{"float64 negative", float64(-99.9), -99},
		{"float64 zero", float64(0.0), 0},

		// Integer conversions
		{"int positive", 100, 100},
		{"int negative", -100, -100},
		{"int16", int16(32000), 32000},
		{"int32", int32(2000000000), 2000000000},
		{"int64", int64(9000000000000000000), 9000000000000000000},

		// Unsigned int
		{"uint small", uint(500), 500},

		// Unknown types return 0
		{"bool true", true, 0},
		{"bool false", false, 0},
		{"nil", nil, 0},
		{"struct", struct{}{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToInt(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToString(t *testing.T) {
	assert.Equal(t, "55", ToString(55))
	assert.Equal(t, "55.000000", ToString(55.0))
	assert.Equal(t, "55", ToString([]byte("55")))
	dir, _ := time.ParseDuration("55s")
	assert.Equal(t, "55s", ToString(dir))
	assert.Equal(t, "true", ToString(true))
	assert.Equal(t, Now().Format("Monday January _2, 2006 at 03:04PM"), ToString(Now()))
}

func ExampleToString() {
	amount := 42
	fmt.Print(ToString(amount))
	// Output: 42
}

func TestSaveFile(t *testing.T) {
	assert.Nil(t, SaveFile(Directory+"/test.txt", []byte("testing saving a file")))
}

func TestOpenFile(t *testing.T) {
	f, err := OpenFile(Directory + "/test.txt")
	require.Nil(t, err)
	assert.Equal(t, "testing saving a file", f)
}

func TestOpenFileComprehensive(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("read existing file", func(t *testing.T) {
		testFile := tmpDir + "/test_read.txt"
		content := "test content for reading"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := OpenFile(testFile)
		assert.NoError(t, err)
		assert.Equal(t, content, result)
	})

	t.Run("read empty file", func(t *testing.T) {
		testFile := tmpDir + "/empty.txt"
		err := os.WriteFile(testFile, []byte(""), 0644)
		require.NoError(t, err)

		result, err := OpenFile(testFile)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("read non-existent file returns error", func(t *testing.T) {
		_, err := OpenFile(tmpDir + "/nonexistent.txt")
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("read file with unicode content", func(t *testing.T) {
		testFile := tmpDir + "/unicode.txt"
		content := "Hello, 世界! 🌍 Ümlauts"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := OpenFile(testFile)
		assert.NoError(t, err)
		assert.Equal(t, content, result)
	})

	t.Run("read binary-like content", func(t *testing.T) {
		testFile := tmpDir + "/binary.txt"
		content := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
		err := os.WriteFile(testFile, content, 0644)
		require.NoError(t, err)

		result, err := OpenFile(testFile)
		assert.NoError(t, err)
		assert.Equal(t, string(content), result)
	})
}

func TestFileExists(t *testing.T) {
	assert.True(t, FileExists(Directory+"/test.txt"))
	assert.False(t, FileExists(Directory+"fake.txt"))
}

func TestFileExistsComprehensive(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("existing file returns true", func(t *testing.T) {
		testFile := tmpDir + "/exists.txt"
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		assert.True(t, FileExists(testFile))
	})

	t.Run("non-existent file returns false", func(t *testing.T) {
		assert.False(t, FileExists(tmpDir+"/does_not_exist.txt"))
	})

	t.Run("directory returns true", func(t *testing.T) {
		// FileExists returns true for directories too (os.Stat succeeds)
		assert.True(t, FileExists(tmpDir))
	})

	t.Run("empty path returns false", func(t *testing.T) {
		assert.False(t, FileExists(""))
	})
}

func TestFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"simple extension", "file.txt", "txt"},
		{"multiple dots", "file.test.json", "json"},
		{"no extension", "file", "file"},
		{"hidden file", ".gitignore", "gitignore"},
		{"path with extension", "/path/to/file.go", "go"},
		{"empty string", "", ""},
		{"dot only", ".", ""},
		{"ends with dot", "file.", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExtension(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFolderExists(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("existing directory returns true", func(t *testing.T) {
		assert.True(t, FolderExists(tmpDir))
	})

	t.Run("non-existent directory returns false", func(t *testing.T) {
		assert.False(t, FolderExists(tmpDir+"/nonexistent"))
	})

	t.Run("file returns false", func(t *testing.T) {
		testFile := tmpDir + "/file.txt"
		err := os.WriteFile(testFile, []byte("content"), 0644)
		require.NoError(t, err)

		assert.False(t, FolderExists(testFile))
	})
}

func TestCreateAndDeleteDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("create new directory", func(t *testing.T) {
		newDir := tmpDir + "/new_dir"
		err := CreateDirectory(newDir)
		assert.NoError(t, err)
		assert.True(t, FolderExists(newDir))
	})

	t.Run("create existing directory returns nil", func(t *testing.T) {
		existingDir := tmpDir + "/existing"
		err := os.Mkdir(existingDir, 0750)
		require.NoError(t, err)

		// Creating again should not error
		err = CreateDirectory(existingDir)
		assert.NoError(t, err)
	})

	t.Run("delete directory", func(t *testing.T) {
		delDir := tmpDir + "/to_delete"
		err := os.Mkdir(delDir, 0750)
		require.NoError(t, err)

		err = DeleteDirectory(delDir)
		assert.NoError(t, err)
		assert.False(t, FolderExists(delDir))
	})

	t.Run("delete non-existent directory returns nil", func(t *testing.T) {
		err := DeleteDirectory(tmpDir + "/nonexistent_dir")
		assert.NoError(t, err) // os.RemoveAll returns nil for non-existent
	})
}

func TestSaveFileComprehensive(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("save new file", func(t *testing.T) {
		testFile := tmpDir + "/new_file.txt"
		content := []byte("new file content")

		err := SaveFile(testFile, content)
		assert.NoError(t, err)

		// Verify content
		data, err := os.ReadFile(testFile)
		assert.NoError(t, err)
		assert.Equal(t, content, data)
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		testFile := tmpDir + "/overwrite.txt"
		err := os.WriteFile(testFile, []byte("old content"), 0644)
		require.NoError(t, err)

		newContent := []byte("new content")
		err = SaveFile(testFile, newContent)
		assert.NoError(t, err)

		data, err := os.ReadFile(testFile)
		assert.NoError(t, err)
		assert.Equal(t, newContent, data)
	})

	t.Run("save empty file", func(t *testing.T) {
		testFile := tmpDir + "/empty.txt"
		err := SaveFile(testFile, []byte(""))
		assert.NoError(t, err)

		data, err := os.ReadFile(testFile)
		assert.NoError(t, err)
		assert.Empty(t, data)
	})
}

func TestDeleteFile(t *testing.T) {
	assert.Nil(t, DeleteFile(Directory+"/test.txt"))
	assert.Error(t, DeleteFile(Directory+"/missingfilehere.txt"))
}

func TestFormatDuration(t *testing.T) {
	dur, _ := time.ParseDuration("158s")
	assert.Equal(t, "2 minutes 38 seconds", FormatDuration(dur))
	dur, _ = time.ParseDuration("-65s")
	assert.Equal(t, "-1 minute 5 seconds", FormatDuration(dur))
	dur, _ = time.ParseDuration("3s")
	assert.Equal(t, "3 seconds", FormatDuration(dur))
	dur, _ = time.ParseDuration("48h")
	assert.Equal(t, "2 days", FormatDuration(dur))
	dur, _ = time.ParseDuration("12h")
	assert.Equal(t, "12 hours", FormatDuration(dur))
}

func ExampleDurationReadable() {
	dur, _ := time.ParseDuration("25m")
	readable := DurationReadable(dur)
	fmt.Print(readable)
	// Output: 25 minutes
}

func TestDurationReadableComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		// Hours
		{"1 hour", 1 * time.Hour, "1 hours"},
		{"2 hours", 2 * time.Hour, "2 hours"},
		{"24 hours", 24 * time.Hour, "24 hours"},
		{"1.5 hours", 90 * time.Minute, "2 hours"},

		// Minutes
		{"1 minute", 1 * time.Minute, "1 minutes"},
		{"30 minutes", 30 * time.Minute, "30 minutes"},
		{"59 minutes", 59 * time.Minute, "59 minutes"},

		// Seconds
		{"1 second", 1 * time.Second, "1 seconds"},
		{"30 seconds", 30 * time.Second, "30 seconds"},
		{"59 seconds", 59 * time.Second, "59 seconds"},

		// Sub-second (uses d.String())
		{"500 milliseconds", 500 * time.Millisecond, "500ms"},
		{"100 microseconds", 100 * time.Microsecond, "100µs"},
		{"10 nanoseconds", 10 * time.Nanosecond, "10ns"},
		{"zero duration", 0, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DurationReadable(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHumanMicro(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		// Microseconds (< 10000)
		{"zero", 0, "0 ms"},
		{"1 microsecond", 1, "1 μs"},
		{"100 microseconds", 100, "100 μs"},
		{"9999 microseconds", 9999, "9999 μs"},

		// Milliseconds (>= 10000)
		{"10000 microseconds", 10000, "10 ms"},
		{"45619 microseconds", 45619, "46 ms"},
		{"1000000 microseconds", 1000000, "1000 ms"},

		// Negative values
		{"negative small", -100, "-100 μs"},
		{"negative large", -50000, "-50 ms"},
		{"-9999 microseconds", -9999, "-9999 μs"},
		{"-10000 microseconds", -10000, "-10 ms"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HumanMicro(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogHTTP(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	assert.Nil(t, err)
	assert.NotNil(t, req)
}

func TestStringInt(t *testing.T) {
	assert.Equal(t, "1", ToString("1"))
}

func TestHashPassword(t *testing.T) {
	pass := HashPassword("password123")
	assert.Equal(t, 60, len(pass))
	assert.True(t, CheckHash("password123", pass))
	assert.False(t, CheckHash("wrongpasswd", pass))
}

func TestHuman(t *testing.T) {
	assert.Equal(t, "10 seconds", Duration{10 * time.Second}.Human())
	assert.Equal(t, "1 day 12 hours", Duration{36 * time.Hour}.Human())
	assert.Equal(t, "45 minutes", Duration{45 * time.Minute}.Human())
}

func TestSha256Hash(t *testing.T) {
	assert.Equal(t, "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f", Sha256Hash("password123"))
}

func TestNotNumbber(t *testing.T) {
	assert.True(t, NotNumber("notint"))
	assert.True(t, NotNumber("1293notanint922"))
	assert.False(t, NotNumber("0"))
	assert.False(t, NotNumber("5"))
}

func TestNewSHA1Hash(t *testing.T) {
	hash := NewSHA256Hash()
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
	assert.Len(t, NewSHA256Hash(), 64)
	assert.NotEqual(t, hash, NewSHA256Hash())
}

func TestRandomString(t *testing.T) {
	assert.NotEmpty(t, RandomString(5))
}

func TestDeleteDirectory(t *testing.T) {
	// Close log file handles before deleting logs directory (required on Windows)
	CloseLogs()
	assert.Nil(t, DeleteDirectory(Directory+"/logs"))
}

func TestRenameDirectory(t *testing.T) {
	assert.Nil(t, CreateDirectory(Directory+"/example"))
	require.DirExists(t, Directory+"/example")
	assert.Nil(t, RenameDirectory(Directory+"/example", Directory+"/renamed_example"))
	require.DirExists(t, Directory+"/renamed_example")
	assert.Nil(t, os.RemoveAll(Directory+"/renamed_example"))
}

func TestHttpRequest(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "/")
		assert.Equal(t, req.Header["Aaa"], []string{"bbbb="})
		assert.Equal(t, req.Header["Ccc"], []string{"ddd"})
		// Send response to be tested
		_, _ = rw.Write([]byte(`OK`))
	}))
	// Close the server when test finishes
	defer server.Close()

	body, resp, err := HttpRequest(server.URL, "GET", "application/json", []string{"aaa=bbbb=", "ccc=ddd"}, nil, 2*time.Second, false, nil)

	assert.Nil(t, err)
	assert.Equal(t, []byte("OK"), body)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestConfigLoad(t *testing.T) {
	err := InitLogs()
	require.Nil(t, err)
	InitEnvs()

	s := Params.GetString
	b := Params.GetBool

	Params.Set("DB_CONN", "sqlite")
	Params.Set("SAMPLE_DATA", true)
	Params.Set("ALLOW_REPORTS", true)

	assert.Equal(t, "sqlite", s("DB_CONN"))
	assert.Equal(t, Directory, s("STATPING_DIR"))
	assert.True(t, b("SAMPLE_DATA"))
	assert.True(t, b("ALLOW_REPORTS"))
}

func TestPerlin(t *testing.T) {
	p := NewPerlin(2, 2, 5, Now().UnixNano())
	require.NotNil(t, p)

	for hi := 1.; hi <= 100.; hi++ {
		assert.NotZero(t, p.Noise1D(hi/500))
	}
}

func TestPing(t *testing.T) {
	duration, error := Ping("localhost", 1)

	assert.Nil(t, error)
	assert.NotEqual(t, 0, duration)
}

func TestIsHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"bcrypt 2a hash", "$2a$14$somehashvaluehere", true},
		{"bcrypt 2b hash", "$2b$10$anotherhashvalue", true},
		{"bcrypt 2y hash", "$2y$12$yetanotherhash", true},
		{"plain password", "mypassword123", false},
		{"empty string", "", false},
		{"partial prefix", "$2a", false},
		{"wrong prefix", "$1$md5hash", false},
		{"sha256 hash", "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHash(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComplexityCheck(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"valid 30 char password", "Abcdefghij1234567890abcdefghij", true},
		{"valid long password", "ThisIsAVeryLongPassword123456789", true},
		{"too short", "Short1", false},
		{"29 chars with complexity", "Abcdefghij123456789abcdefghi", false},
		{"30 chars no uppercase", "abcdefghij1234567890abcdefghij", false},
		{"30 chars no lowercase", "ABCDEFGHIJ1234567890ABCDEFGHIJ", false},
		{"30 chars no digits", "AbcdefghijKLMNOPQRSTUabcdefghij", false},
		{"empty string", "", false},
		{"exactly 30 chars valid", "Aa1234567890123456789012345678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComplexityCheck(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}
