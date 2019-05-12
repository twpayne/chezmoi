module github.com/twpayne/chezmoi

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.17.1+incompatible
	github.com/aokoli/goutils v1.1.0 // indirect
	github.com/coreos/go-semver v0.2.0
	github.com/danieljoos/wincred v1.0.1 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/google/go-github/v25 v25.0.1
	github.com/google/renameio v0.1.0
	github.com/google/uuid v1.1.0 // indirect
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kr/text v0.1.0
	github.com/mattn/go-isatty v0.0.7
	github.com/pmezard/go-difflib v1.0.0
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/twpayne/go-shell v0.0.1
	github.com/twpayne/go-vfs v1.0.6
	github.com/twpayne/go-vfsafero v1.0.0
	github.com/twpayne/go-xdg/v3 v3.1.0
	github.com/zalando/go-keyring v0.0.0-20180221093347-6d81c293b3fb
	go.etcd.io/bbolt v1.3.2
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sys v0.0.0-20190310054646-10058d7d4faa // indirect
	gopkg.in/yaml.v2 v2.2.2
)

// github.com/pmezard/go-difflib is unmaintained, so use a fork that includes
// colored diff support.
replace github.com/pmezard/go-difflib => github.com/twpayne/go-difflib v1.2.1

// go.etcd.io/bbolt requires a couple of fixes before it can be used, so use a
// fork with the fixes. This replace can be removed once
// https://github.com/etcd-io/bbolt/pull/157 is merged and a new version of
// go.etcd.io/bbolt is tagged.
replace go.etcd.io/bbolt => github.com/twpayne/bbolt v1.3.3
