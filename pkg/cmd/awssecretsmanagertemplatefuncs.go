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
	Region  string
	Profile string

	svc       *secretsmanager.Client
	cache     map[string]string
	jsonCache map[string]map[string]interface{}
}

func (c *Config) awsSecretsManagerRawTemplateFunc(arn string) string {
	if secret, ok := c.AwsSecretsManager.cache[arn]; ok {
		return secret
	}

	if c.AwsSecretsManager.svc == nil {
		var opts []func(*config.LoadOptions) error
		if region := c.AwsSecretsManager.Region; len(region) > 0 {
			opts = append(opts, config.WithRegion(region))
		}
		if profile := c.AwsSecretsManager.Profile; len(profile) > 0 {
			opts = append(opts, config.WithSharedConfigProfile(profile))
		}

		opts = append(opts, config.WithRetryMaxAttempts(1))

		cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
		if err != nil {
			panic(err)
		}

		c.AwsSecretsManager.svc = secretsmanager.NewFromConfig(cfg)
	}

	result, err := c.AwsSecretsManager.svc.GetSecretValue(context.Background(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
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

	if c.AwsSecretsManager.cache == nil {
		c.AwsSecretsManager.cache = make(map[string]string)
	}

	c.AwsSecretsManager.cache[arn] = secret
	return secret
}

func (c *Config) awsSecretsManagerTemplateFunc(arn string) map[string]interface{} {
	if secret, ok := c.AwsSecretsManager.jsonCache[arn]; ok {
		return secret
	}

	raw := c.awsSecretsManagerRawTemplateFunc(arn)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		panic(err)
	}

	if c.AwsSecretsManager.jsonCache == nil {
		c.AwsSecretsManager.jsonCache = make(map[string]map[string]interface{})
	}

	c.AwsSecretsManager.jsonCache[arn] = data
	return data
}
