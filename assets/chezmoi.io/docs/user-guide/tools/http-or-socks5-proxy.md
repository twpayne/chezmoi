# HTTP or SOCKS5 proxy

chezmoi supports HTTP, HTTPS, and SOCKS5 proxies. Set the `HTTP_PROXY`,
`HTTPS_PROXY`, and `NO_PROXY` environment variables, or their lowercase
equivalents, for example:

```sh
HTTP_PROXY=socks5://127.0.0.1:1080 chezmoi apply --refresh-externals
```
