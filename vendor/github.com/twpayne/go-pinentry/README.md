# go-pinentry

[![PkgGoDev](https://pkg.go.dev/badge/github.com/twpayne/go-pinentry)](https://pkg.go.dev/github.com/twpayne/go-pinentry)

Package `pinentry` provides a client to [GnuPG's
pinentry](https://www.gnupg.org/related_software/pinentry/index.html).

## Key Features

* Support for all `pinentry` features.
* Idiomatic Go API.
* Well tested.

## Example

```go
	client, err := pinentry.NewClient(
		pinentry.WithBinaryNameFromGnuPGAgentConf(),
		pinentry.WithDesc("My description"),
		pinentry.WithGPGTTY(),
		pinentry.WithPrompt("My prompt:"),
		pinentry.WithTitle("My title")
	)
	if err != nil {
		return err
	}
	defer client.Close()

	switch pin, fromCache, err := client.GetPIN(); {
	case pinentry.IsCancelled(err):
		fmt.Println("Cancelled")
	case err != nil:
		return err
	case fromCache:
		fmt.Printf("PIN: %s (from cache)\n", pin)
	default:
		fmt.Printf("PIN: %s\n", pin)
	}
```

## Comparison with related packages

Compared to
[`github.com/gopasspw/pinentry`](https://github.com/gopasspw/pinentry), this
package:
* Implements all `pinentry` features.
* Includes tests.
* Implements a full parser of the underlying Assuan protocol for better
  compatibility with all `pinentry` implementations.

## License

MIT
