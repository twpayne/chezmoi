# knownhosts: enhanced Golang SSH known_hosts management

[![build status](https://img.shields.io/github/actions/workflow/status/skeema/knownhosts/tests.yml?branch=main)](https://github.com/skeema/knownhosts/actions)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/skeema/knownhosts)


> This repo is brought to you by [Skeema](https://github.com/skeema/skeema), a
> declarative pure-SQL schema management system for MySQL and MariaDB. Our
> premium products include extensive [SSH tunnel](https://www.skeema.io/docs/options/#ssh)
> functionality, which internally makes use of this package.

Go provides excellent functionality for OpenSSH known_hosts files in its
external package [golang.org/x/crypto/ssh/knownhosts](https://pkg.go.dev/golang.org/x/crypto/ssh/knownhosts). 
However, that package is somewhat low-level, making it difficult to implement full known_hosts management similar to command-line `ssh`'s behavior for `StrictHostKeyChecking=no` configuration.

This repo ([github.com/skeema/knownhosts](https://github.com/skeema/knownhosts)) is a thin wrapper package around [golang.org/x/crypto/ssh/knownhosts](https://pkg.go.dev/golang.org/x/crypto/ssh/knownhosts), adding the following functionality:

* Look up known_hosts public keys for any given host
* Auto-populate ssh.ClientConfig.HostKeyAlgorithms easily based on known_hosts, providing a solution for [golang/go#29286](https://github.com/golang/go/issues/29286)
* Write new known_hosts entries to an io.Writer
* Properly format/normalize new known_hosts entries containing ipv6 addresses, providing a solution for [golang/go#53463](https://github.com/golang/go/issues/53463)
* Determine if an ssh.HostKeyCallback's error corresponds to a host whose key has changed (indicating potential MitM attack) vs a host that just isn't known yet

## How host key lookup works

Although [golang.org/x/crypto/ssh/knownhosts](https://pkg.go.dev/golang.org/x/crypto/ssh/knownhosts) doesn't directly expose a way to query its known_host map, we use a subtle trick to do so: invoke the HostKeyCallback with a valid host but a bogus key. The resulting KeyError allows us to determine which public keys are actually present for that host.

By using this technique, [github.com/skeema/knownhosts](https://github.com/skeema/knownhosts) doesn't need to duplicate or re-implement any of the actual known_hosts management from [golang.org/x/crypto/ssh/knownhosts](https://pkg.go.dev/golang.org/x/crypto/ssh/knownhosts).

## Populating ssh.ClientConfig.HostKeyAlgorithms based on known_hosts

Hosts often have multiple public keys, each of a different type (algorithm). This can be [problematic](https://github.com/golang/go/issues/29286) in [golang.org/x/crypto/ssh/knownhosts](https://pkg.go.dev/golang.org/x/crypto/ssh/knownhosts): if a host's first public key is *not* in known_hosts, but a key of a different type *is*, the HostKeyCallback returns an error. The solution is to populate `ssh.ClientConfig.HostKeyAlgorithms` based on the algorithms of the known_hosts entries for that host, but 
[golang.org/x/crypto/ssh/knownhosts](https://pkg.go.dev/golang.org/x/crypto/ssh/knownhosts)
does not provide an obvious way to do so.

This package uses its host key lookup trick in order to make ssh.ClientConfig.HostKeyAlgorithms easy to populate:

```golang
import (
	"golang.org/x/crypto/ssh"
	"github.com/skeema/knownhosts"
)

func sshConfigForHost(hostWithPort string) (*ssh.ClientConfig, error) {
	kh, err := knownhosts.New("/home/myuser/.ssh/known_hosts")
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:              "myuser",
		Auth:              []ssh.AuthMethod{ /* ... */ },
		HostKeyCallback:   kh.HostKeyCallback(), // or, equivalently, use ssh.HostKeyCallback(kh)
		HostKeyAlgorithms: kh.HostKeyAlgorithms(hostWithPort),
	}
	return config, nil
}
```

## Writing new known_hosts entries

If you wish to mimic the behavior of OpenSSH's `StrictHostKeyChecking=no` or `StrictHostKeyChecking=ask`, this package provides a few functions to simplify this task. For example:

```golang
sshHost := "yourserver.com:22"
khPath := "/home/myuser/.ssh/known_hosts"
kh, err := knownhosts.New(khPath)
if err != nil {
	log.Fatal("Failed to read known_hosts: ", err)
}

// Create a custom permissive hostkey callback which still errors on hosts
// with changed keys, but allows unknown hosts and adds them to known_hosts
cb := ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
	err := kh(hostname, remote, key)
	if knownhosts.IsHostKeyChanged(err) {
		return fmt.Errorf("REMOTE HOST IDENTIFICATION HAS CHANGED for host %s! This may indicate a MitM attack.", hostname)
	} else if knownhosts.IsHostUnknown(err) {
		f, ferr := os.OpenFile(khPath, os.O_APPEND|os.O_WRONLY, 0600)
		if ferr == nil {
			defer f.Close()
			ferr = knownhosts.WriteKnownHost(f, hostname, remote, key)
		}
		if ferr == nil {
			log.Printf("Added host %s to known_hosts\n", hostname)
		} else {
			log.Printf("Failed to add host %s to known_hosts: %v\n", hostname, ferr)
		}
		return nil // permit previously-unknown hosts (warning: may be insecure)
	}
	return err
})

config := &ssh.ClientConfig{
	User:              "myuser",
	Auth:              []ssh.AuthMethod{ /* ... */ },
	HostKeyCallback:   cb,
	HostKeyAlgorithms: kh.HostKeyAlgorithms(sshHost),
}
```

## License

**Source code copyright 2023 Skeema LLC and the Skeema Knownhosts authors**

```text
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
