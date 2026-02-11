package ssh

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulfikawr/fm/internal/logger"
)

// ParseSSHConfig parses the ~/.ssh/config file.
func ParseSSHConfig() (map[string]*SSHConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(home, ".ssh", "config")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*SSHConfig), nil
		}
		return nil, err
	}
	defer logger.CloseAndLog(file, "SSH config file")

	configs := make(map[string]*SSHConfig)
	var currentHost []string
	var currentCfg *SSHConfig

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "host":
			// Save previous config if any
			if currentCfg != nil {
				for i := range currentHost {
					h := currentHost[i]
					// We don't handle wildcards well here, but it's a start
					if h != "*" {
						configs[h] = currentCfg
					}
				}
			}
			currentHost = parts[1:]
			currentCfg = &SSHConfig{}
		case "hostname":
			if currentCfg != nil {
				currentCfg.HostName = value
			}
		case "user":
			if currentCfg != nil {
				currentCfg.User = value
			}
		case "port":
			if currentCfg != nil {
				currentCfg.Port = value
			}
		case "identityfile":
			if currentCfg != nil {
				// Expand ~ if present
				if strings.HasPrefix(value, "~") {
					value = filepath.Join(home, value[1:])
				}
				currentCfg.IdentityFile = value
			}
		}
	}

	// Save last config
	if currentCfg != nil {
		for i := range currentHost {
			h := currentHost[i]
			if h != "*" {
				configs[h] = currentCfg
			}
		}
	}

	return configs, nil
}
