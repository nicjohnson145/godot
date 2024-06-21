package lib

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	vault "github.com/hashicorp/vault-client-go"
)

func setVaultClient(userConf *UserConfig) error {
	if userConf.VaultConfig.Address == "" {
		return nil
	}

	client, err := vault.New(
		vault.WithAddress(userConf.VaultConfig.Address),
		vault.WithRequestTimeout(15*time.Second),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	tokenPath := filepath.Join(userConf.HomeDir, ".vault-token")
	if userConf.VaultConfig.TokenPath != "" {
		tokenPath = userConf.VaultConfig.TokenPath
	}

	exists, err := pathExists(replaceTilde(tokenPath, userConf.HomeDir))
	if err != nil {
		return fmt.Errorf("unable to check path existance: %v", err)
	}

	if !exists {
		return fmt.Errorf("vault token path %v does not exist", tokenPath)
	}

	b, err := os.ReadFile(tokenPath)
	if err != nil {
		return fmt.Errorf("error reading vault token file: %v", err)
	}

	client.SetToken(string(b))
	userConf.VaultConfig.Client = RealVaultClient{Client: client}
	return nil
}

type VaultClient interface {
	Initialized() bool
	ReadKey(string, string) (string, error)
}

type RealVaultClient struct {
	Client *vault.Client
}

func (v RealVaultClient) Initialized() bool {
	return v.Client != nil
}

func (v RealVaultClient) ReadKey(path string, key string) (string, error) {
	secret, err := v.Client.Read(context.Background(), path)
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
