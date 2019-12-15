module github.com/twpayne/chezmoi

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/go-github/v26 v26.1.3
	github.com/google/renameio v0.1.0
	github.com/google/uuid v1.1.1 // indirect
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95 // indirect
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kr/text v0.1.0
	github.com/mattn/go-isatty v0.0.10
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/diff v0.0.0-20190930165518-531926345625
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.5.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/twpayne/go-shell v0.1.1
	github.com/twpayne/go-vfs v1.3.6
	github.com/twpayne/go-vfsafero v1.0.0
	github.com/twpayne/go-xdg/v3 v3.1.0
	github.com/zalando/go-keyring v0.0.0-20190913082157-62750a1ff80d
	go.etcd.io/bbolt v1.3.3
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf
	golang.org/x/net v0.0.0-20191101175033-0deb6923b6d9 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20191029155521-f43be2a4598c // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/yaml.v2 v2.2.4
)

// Temporary while waiting for https://github.com/spf13/cobra/pull/754 to be merged
replace github.com/spf13/cobra => github.com/0robustus1/cobra v0.0.4-0.20190522074606-64400adf086c
