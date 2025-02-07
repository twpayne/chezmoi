# AWS Secrets Manager functions

The `awsSecretsManager*` functions return data from [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)
using the [`GetSecretValue`](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html)
API.

The profile and region are pulled from the standard environment variables and shared config files but can be
overridden by setting `awsSecretsManager.profile` and `awsSecretsManager.region` configuration variables respectively.

+++ 2.19.0
