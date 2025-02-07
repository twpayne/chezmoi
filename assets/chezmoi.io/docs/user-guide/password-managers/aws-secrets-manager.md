# AWS Secrets Manager

chezmoi includes support for [AWS Secrets Manager][awssm].

Structured data can be retrieved with the `awsSecretsManager` template function,
for example:

```text
exampleUsername = {{ (awsSecretsManager "my-secret-name").username }}
examplePassword = {{ (awsSecretsManager "my-secret-name").password }}
```

For retrieving unstructured data, the `awsSecretsManagerRaw` template function
can be used. For example:

```text
exampleSecretString = {{ awsSecretsManagerRaw "my-secret-string" }}
```

The AWS shared profile name and region can be specified in chezmoi's config file
with `awsSecretsManager.profile` and `awsSecretsManager.region` respectively. By
default, these values will be picked up from the standard environment variables
and config files used by the standard AWS tooling.

```toml title="~/.config/chezmoi/chezmoi.toml"
[awsSecretsManager]
    profile = myWorkProfile
    region = us-east-2
```

[awssm]: https://aws.amazon.com/secrets-manager/
