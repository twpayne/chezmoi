# `awsSecretsManager` *arn*

`awsSecretsManager` returns structured data retrieved from [AWS Secrets
Manager][awssm]. *arn* specifies the `SecretId` passed to
[`GetSecretValue`][gsv]. This can either be the full ARN or the [simpler
name][name] if applicable.

[awssm]: https://aws.amazon.com/secrets-manager/
[gsv]: https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
[name]: https://docs.aws.amazon.com/secretsmanager/latest/userguide/troubleshoot.html#ARN_secretnamehyphen
