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
	userConf.VaultConfig.Client = VaultClient{Client: client}
	log.Debug("Initialized vault client")
}

type VaultClient struct {
	Client *vault.Client
}


func (v VaultClient) Initialized() bool {
	return v.Client != nil
}


func (v VaultClient) ReadKey(path string, key string) (string, error) {
	secret, err := v.Client.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("Error reading path %v from vault: %w", path, err)
	}

	data := map[string]string{}
	rawData := secret.Data["data"].(map[string]interface{})
	for key, value := range rawData {
		data[key] = fmt.Sprint(value)
	}

	value, ok := data[key]
	if !ok {
		return "", fmt.Errorf("Key %v not present at vault path %v", key, path)
	}

	return value, nil
}

func (v VaultClient) ReadKeyOrDie(path string, key string) string {
	val, err := v.ReadKey(path, key)
	if err != nil {
		log.Fatal(err)
	}

	return val
}
