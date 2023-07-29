package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type awsSecretsManagerConfig struct {
	Region    string `json:"region"  mapstructure:"region"  yaml:"region"`
	Profile   string `json:"profile" mapstructure:"profile" yaml:"profile"`
	svc       *secretsmanager.Client
	cache     map[string]string
	jsonCache map[string]map[string]any
}

func (c *Config) awsSecretsManagerRawTemplateFunc(arn string) string {
	if secret, ok := c.AWSSecretsManager.cache[arn]; ok {
		return secret
	}

	if c.AWSSecretsManager.svc == nil {
		var opts []func(*config.LoadOptions) error
		if region := c.AWSSecretsManager.Region; len(region) > 0 {
			opts = append(opts, config.WithRegion(region))
		}
		if profile := c.AWSSecretsManager.Profile; len(profile) > 0 {
			opts = append(opts, config.WithSharedConfigProfile(profile))
		}

		opts = append(opts, config.WithRetryMaxAttempts(1))

		cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
		if err != nil {
			panic(err)
		}

		c.AWSSecretsManager.svc = secretsmanager.NewFromConfig(cfg)
	}

	result, err := c.AWSSecretsManager.svc.GetSecretValue(
		context.Background(),
		&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(arn),
		},
	)
	if err != nil {
		panic(err)
	}

	var secret string
	if result.SecretString != nil {
		secret = *result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		length, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			panic(err)
		}

		secret = string(decodedBinarySecretBytes[:length])
	}

	if c.AWSSecretsManager.cache == nil {
		c.AWSSecretsManager.cache = make(map[string]string)
	}

	c.AWSSecretsManager.cache[arn] = secret
	return secret
}

func (c *Config) awsSecretsManagerTemplateFunc(arn string) map[string]any {
	if secret, ok := c.AWSSecretsManager.jsonCache[arn]; ok {
		return secret
	}

	raw := c.awsSecretsManagerRawTemplateFunc(arn)

	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		panic(err)
	}

	if c.AWSSecretsManager.jsonCache == nil {
		c.AWSSecretsManager.jsonCache = make(map[string]map[string]any)
	}

	c.AWSSecretsManager.jsonCache[arn] = data
	return data
}
