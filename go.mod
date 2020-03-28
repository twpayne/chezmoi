module github.com/twpayne/chezmoi

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/alecthomas/chroma v0.7.1 // indirect
	github.com/charmbracelet/glamour v0.1.0
	github.com/coreos/go-semver v0.3.0
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/google/go-github/v26 v26.1.3
	github.com/google/renameio v0.1.0
	github.com/google/uuid v1.1.1 // indirect
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381 // indirect
	github.com/mattn/go-isatty v0.0.11 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/pelletier/go-toml v1.6.0
	github.com/pkg/diff v0.0.0-20190930165518-531926345625
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.7
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.2
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/twpayne/go-shell v0.1.1
	github.com/twpayne/go-vfs v1.3.6
	github.com/twpayne/go-vfsafero v1.0.0
	github.com/twpayne/go-xdg/v3 v3.1.0
	github.com/yuin/goldmark v1.1.26 // indirect
	github.com/zalando/go-keyring v0.0.0-20200121091418-667557018717
	go.etcd.io/bbolt v1.3.4
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect; indirectq
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200327173247-9dae0f8f5775
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/ini.v1 v1.55.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

// Temporary while waiting for https://github.com/spf13/cobra/pull/1048 to be merged
replace github.com/spf13/cobra => github.com/twpayne/cobra v0.0.8
