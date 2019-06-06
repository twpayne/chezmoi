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
	github.com/kr/text v0.1.0
	github.com/mattn/go-isatty v0.0.7
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.3.2
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/twpayne/go-difflib v1.3.0
	github.com/twpayne/go-shell v0.0.1
	github.com/twpayne/go-vfs v1.1.0
	github.com/twpayne/go-vfsafero v1.0.0
	github.com/twpayne/go-xdg/v3 v3.1.0
	github.com/zalando/go-keyring v0.0.0-20180221093347-6d81c293b3fb
	go.etcd.io/bbolt v1.3.3-0.20190510211640-4af6cfab7010
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sys v0.0.0-20190310054646-10058d7d4faa // indirect
	gopkg.in/yaml.v2 v2.2.2
)

// Temporary while waiting for https://github.com/spf13/cobra/pull/754 to be merged
replace github.com/spf13/cobra => github.com/0robustus1/cobra v0.0.4-0.20190522074606-64400adf086c
