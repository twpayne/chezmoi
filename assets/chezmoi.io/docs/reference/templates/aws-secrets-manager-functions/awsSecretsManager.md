# `awsSecretsManager` *arn*

`awsSecretsManager` returns structured data retrieved from
[AWS Secrets Manager](https://aws.amazon.com/secrets-manager/). *arn* specifies the `SecretId` passed to
[`GetSecretValue`](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html). This can
either be the full ARN or the
[simpler name](https://docs.aws.amazon.com/secretsmanager/latest/userguide/troubleshoot.html#ARN_secretnamehyphen)
if applicable.

+++ 2.19.0
