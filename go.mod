module github.com/twpayne/chezmoi

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/alecthomas/chroma v0.8.0 // indirect
	github.com/bmatcuk/doublestar v1.3.2
	github.com/charmbracelet/glamour v0.2.0
	github.com/coreos/go-semver v0.3.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-git/go-git/v5 v5.1.0
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-github/v26 v26.1.3
	github.com/google/renameio v0.1.0
	github.com/google/uuid v1.1.1 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/microcosm-cc/bluemonday v1.0.4 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/muesli/termenv v0.7.2 // indirect
	github.com/pelletier/go-toml v1.8.0
	github.com/pkg/diff v0.0.0-20190930165518-531926345625
	github.com/rogpeppe/go-internal v1.6.1
	github.com/sergi/go-diff v1.1.0
	github.com/spf13/afero v1.3.4 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/twpayne/go-shell v0.3.0
	github.com/twpayne/go-vfs v1.7.0
	github.com/twpayne/go-vfsafero v1.0.0
	github.com/twpayne/go-xdg/v3 v3.1.0
	github.com/yuin/goldmark v1.2.1 // indirect
	go.etcd.io/bbolt v1.3.5
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200828194041-157a740278f4
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/ini.v1 v1.60.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

// Temporary while waiting for https://github.com/rogpeppe/go-internal/pull/106 to be merged.
replace github.com/rogpeppe/go-internal => github.com/twpayne/go-internal v1.5.3-0.20200706163000-4426ab554b0a
