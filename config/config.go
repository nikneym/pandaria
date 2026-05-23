package config

import (
	"os"
)

type Config struct {
	RootDir   string
	ToolsDir  string
	BiomePath string
}

var C Config

// Setup sets up global config.
func Setup() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rootDir := homeDir + "/.pandaria"
	toolsDir := rootDir + "/tools"
	biomePath := toolsDir + "/biome"

	C = Config{
		RootDir:   rootDir,
		ToolsDir:  toolsDir,
		BiomePath: biomePath,
	}

	// Create necessary directories if not exist.
	os.Mkdir(rootDir, 0755)
	os.Mkdir(toolsDir, 0755)

	return nil
}
