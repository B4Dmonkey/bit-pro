package claude

import (
	"fmt"
	"os"
	"path/filepath"
)

const settings = `{
  "extraKnownMarketplaces": {
    "bit-pro": {
      "source": {
        "source": "github",
        "repo": "B4Dmonkey/bit-pro"
      }
    }
  },
  "enabledPlugins": {
    "bit@bit-pro": true
  }
}
`

func WriteSettings(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(settings), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
