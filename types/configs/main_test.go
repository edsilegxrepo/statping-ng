package configs

import (
	"os"
	"testing"

	"github.com/statping-ng/statping-ng/utils"
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "statping-configs-test")
	if err != nil {
		os.Exit(1)
	}

	utils.Directory = tmpDir
	utils.InitEnvs()
	utils.Params.Set("STATPING_DIR", tmpDir)

	code := m.Run()

	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}
