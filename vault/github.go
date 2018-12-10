package vault

import (
	"fmt"
	"github.com/hashicorp/vault/api"
)

func isGithubAuthBackendPresent(client *api.Client, path string) (bool, error) {
	auths, err := client.Sys().ListAuth()
	if err != nil {
		return false, fmt.Errorf("error reading from Vault: %s", err)
	}

	configuredPath := path + "/"

	for authBackendPath, auth := range auths {

		if auth.Type == "github" && authBackendPath == configuredPath {
			return true, nil
		}
	}

	return false, nil
}

func githubConfigEndpoint(path string) string {
	return fmt.Sprintf("/auth/%s/config", path)
}
