package google

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

const (
	configPath                      = "config"
	clientIDConfigPropertyName      = "client_id"
	clientSecretConfigPropertyName  = "client_secret"
	fetchGroupsPropertyName         = "fetch_groups"
	impersonationPropertyName       = "impersonation"
	adminServiceAccountPropertyName = "admin_service_account"
	configEntry                     = "config"
)

func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var (
		clientID            = data.Get(clientIDConfigPropertyName).(string)
		clientSecret        = data.Get(clientSecretConfigPropertyName).(string)
		fetchGroups         = data.Get(fetchGroupsPropertyName).(bool)
		impersonation       = data.Get(impersonationPropertyName).(string)
		adminServiceAccount = data.Get(adminServiceAccountPropertyName).(string)
	)

	if fetchGroups {
		if impersonation == "" {
			return nil, fmt.Errorf("%s must be configured when %s is enabled", impersonationPropertyName, fetchGroupsPropertyName)
		}
		if adminServiceAccount == "" {
			return nil, fmt.Errorf("%s must be configured when %s is enabled", adminServiceAccountPropertyName, fetchGroupsPropertyName)
		}
	}

	entry, err := logical.StorageEntryJSON(configEntry, config{
		ClientID:            clientID,
		ClientSecret:        clientSecret,
		FetchGroups:         fetchGroups,
		Impersonation:       impersonation,
		AdminServiceAccount: adminServiceAccount,
	})
	if err != nil {
		return nil, err
	}

	return nil, req.Storage.Put(ctx, entry)
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	config, err := b.config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}

	configMap := map[string]interface{}{
		clientIDConfigPropertyName:      config.ClientID,
		clientSecretConfigPropertyName:  config.ClientSecret,
		fetchGroupsPropertyName:         config.FetchGroups,
		impersonationPropertyName:       config.Impersonation,
		adminServiceAccountPropertyName: config.AdminServiceAccount,
	}

	return &logical.Response{
		Data: configMap,
	}, nil
}

// Config returns the configuration for this backend.
func (b *backend) config(ctx context.Context, s logical.Storage) (*config, error) {
	entry, err := s.Get(ctx, configEntry)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var result config
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, fmt.Errorf("error reading configuration: %s", err)
	}

	return &result, nil
}

type config struct {
	ClientID            string `json:"client_id"`
	ClientSecret        string `json:"client_secret"`
	FetchGroups         bool   `json:"fetch_groups"`
	Impersonation       string `json:"impersonation"`
	AdminServiceAccount string `json:"admin_service_account"`
}

func (c *config) oauth2Config() *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}
	return config
}
