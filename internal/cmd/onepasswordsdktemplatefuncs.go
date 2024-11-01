package cmd

import (
	"context"
	"os"

	"github.com/1password/onepassword-sdk-go"
)

type onepasswordSDKConfig struct {
	Token               string `json:"token"       mapstructure:"token"       yaml:"token"`
	TokenEnvVar         string `json:"tokenEnvVar" mapstructure:"tokenEnvVar" yaml:"tokenEnvVar"`
	itemsGetCache       map[string]onepasswordSDKItem
	secretsResolveCache map[string]string
	client              *onepassword.Client
	clientErr           error
}

type onepasswordSDKItem struct {
	ID       string
	Title    string
	Category onepassword.ItemCategory
	VaultID  string
	Fields   map[string]onepassword.ItemField
	Sections map[string]onepassword.ItemSection
}

func (c *Config) onepasswordSDKItemsGet(vaultID, itemID string) onepasswordSDKItem {
	key := vaultID + "\x00" + itemID
	if result, ok := c.OnepasswordSDK.itemsGetCache[key]; ok {
		return result
	}

	ctx := context.TODO()

	client, err := c.onepasswordSDKClient(ctx)
	if err != nil {
		panic(err)
	}

	item, err := client.Items.Get(ctx, vaultID, itemID)
	if err != nil {
		panic(err)
	}

	if c.OnepasswordSDK.itemsGetCache == nil {
		c.OnepasswordSDK.itemsGetCache = make(map[string]onepasswordSDKItem)
	}

	fields := make(map[string]onepassword.ItemField)
	for _, field := range item.Fields {
		fields[field.ID] = field
	}

	sections := make(map[string]onepassword.ItemSection)
	for _, section := range item.Sections {
		sections[section.ID] = section
	}

	onepasswordSDKItem := onepasswordSDKItem{
		ID:       item.ID,
		Title:    item.Title,
		Category: item.Category,
		VaultID:  item.VaultID,
		Fields:   fields,
		Sections: sections,
	}

	c.OnepasswordSDK.itemsGetCache[key] = onepasswordSDKItem

	return onepasswordSDKItem
}

func (c *Config) onepasswordSDKSecretsResolve(secretReference string) string {
	if result, ok := c.OnepasswordSDK.secretsResolveCache[secretReference]; ok {
		return result
	}

	ctx := context.TODO()

	client, err := c.onepasswordSDKClient(ctx)
	if err != nil {
		panic(err)
	}

	secret, err := client.Secrets.Resolve(ctx, secretReference)
	if err != nil {
		panic(err)
	}

	if c.OnepasswordSDK.secretsResolveCache == nil {
		c.OnepasswordSDK.secretsResolveCache = make(map[string]string)
	}
	c.OnepasswordSDK.secretsResolveCache[secretReference] = secret

	return secret
}

func (c *Config) onepasswordSDKClient(ctx context.Context) (*onepassword.Client, error) {
	if c.OnepasswordSDK.client != nil || c.OnepasswordSDK.clientErr != nil {
		return c.OnepasswordSDK.client, c.OnepasswordSDK.clientErr
	}

	token := c.OnepasswordSDK.Token
	if token == "" {
		token = os.Getenv(c.OnepasswordSDK.TokenEnvVar)
	}

	version := c.versionInfo.Version
	if version == "" {
		version = c.versionInfo.Commit
	}
	if version == "" {
		version = onepassword.DefaultIntegrationVersion
	}

	c.OnepasswordSDK.client, c.OnepasswordSDK.clientErr = onepassword.NewClient(
		ctx,
		onepassword.WithIntegrationInfo("chezmoi", version),
		onepassword.WithServiceAccountToken(token),
	)

	return c.OnepasswordSDK.client, c.OnepasswordSDK.clientErr
}
