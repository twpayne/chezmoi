# AWS Secrets Manager functions

The `awsSecretsManager*` functions return data from [AWS Secrets Manager][awssm]
using the [`GetSecretValue`][gsv] API.

The profile and region are pulled from the standard environment variables and
shared config files but can be overridden by setting `awsSecretsManager.profile`
and `awsSecretsManager.region` configuration variables respectively.

[awssm]: https://aws.amazon.com/secrets-manager/
[gsv]: https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
