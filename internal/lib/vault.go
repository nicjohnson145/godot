package lib

import (
	log "github.com/sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"path/filepath"
	"fmt"
	"os"
)

func setVaultClient(userConf *UserConfig) {
	if userConf.VaultConfig.Address == "" {
		log.Debug("No vault address set, not setting vault client")
		return
	}

	config := vault.DefaultConfig()
	config.Address = userConf.VaultConfig.Address
	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("Unabled to initialize Vault client: %v", err)
	}

	tokenPath := filepath.Join(userConf.HomeDir, ".vault-token")
	if userConf.VaultConfig.TokenPath != "" {
		tokenPath = userConf.VaultConfig.TokenPath
	}

	exists, err := pathExists(replaceTilde(tokenPath, userConf.HomeDir))
	if err != nil {
		log.Fatalf("Unable to check path existance: %v", err)
	}

	if !exists {
		log.Fatalf("Vault token path %v does not exist", tokenPath)
	}

	b, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatalf("Error reading vault token file: %v", err)
	}

	client.SetToken(string(b))
	userConf.VaultConfig.Client = client
	log.Debug("Initialized vault client")
}

func ReadKey(client *vault.Client, path string, key string) string {
	secret, err := client.Logical().Read(path)
	if err != nil {
		log.Fatalf("Error reading %v from vault: %v", path, err)
	}

	data := map[string]string{}
	rawData := secret.Data["data"].(map[string]interface{})
	for key, value := range rawData {
		data[key] = fmt.Sprint(value)
	}

	value, ok := data[key]
	if !ok {
		log.Fatalf("Key %v not present at vault path %v", key, path)
	}

	return value
}

