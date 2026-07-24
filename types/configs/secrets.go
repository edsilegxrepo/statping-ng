package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/statping-ng/statping-ng/utils"
)

func SaveSecrets(creds map[string]string) error {
	path := filepath.Join(utils.Directory, "statping.secrets")

	var content string
	content += "################################################################################\n"
	content += "# STATPING-NG AUTOMATICALLY GENERATED SECRETS                                  #\n"
	content += "# Generated on: " + utils.Now().Format("2006-01-02 15:04:05") + "                        #\n"
	content += "################################################################################\n\n"

	// Sort keys for consistent output
	keys := make([]string, 0, len(creds))
	for k := range creds {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, user := range keys {
		content += fmt.Sprintf("User:     %s\n", user)
		content += fmt.Sprintf("Password: %s\n", creds[user])
		content += "--------------------------------------------------------------------------------\n"
	}

	content += "\n# These secrets were generated because Statping was started with SAMPLE_DATA=true\n"
	content += "# or a fresh installation was performed without the setup page.\n"
	content += "# SECURE THIS FILE OR DELETE IT AFTER CHANGING PASSWORDS.\n"

	// Write file with 0400 permissions (read-only for owner)
	err := os.WriteFile(path, []byte(content), 0o400)
	if err != nil {
		return err
	}

	log.Infof("Statping secrets have been saved to: %s", path)
	return nil
}
