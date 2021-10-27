module github.com/twpayne/chezmoi/v2

go 1.16

require (
	filippo.io/age v1.0.0
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/ProtonMail/go-crypto v0.0.0-20210920160938-87db9fbc61c7 // indirect
	github.com/alecthomas/chroma v0.9.4 // indirect
	github.com/bmatcuk/doublestar/v4 v4.0.2
	github.com/bradenhilton/mozillainstallhash v1.0.0
	github.com/charmbracelet/glamour v0.3.0
	github.com/coreos/go-semver v0.3.0
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-github/v39 v39.2.0
	github.com/google/renameio/v2 v2.0.0
	github.com/google/uuid v1.3.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.16 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.2
	github.com/muesli/combinator v0.3.0
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.9.0 // indirect
	github.com/pelletier/go-toml v1.9.4
	github.com/rogpeppe/go-internal v1.8.0
	github.com/rs/zerolog v1.25.0
	github.com/sergi/go-diff v1.1.0
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/twpayne/go-shell v0.3.1
	github.com/twpayne/go-vfs/v4 v4.1.0
	github.com/twpayne/go-xdg/v6 v6.0.0
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	github.com/yuin/goldmark v1.4.1 // indirect
	github.com/zalando/go-keyring v0.1.1
	go.etcd.io/bbolt v1.3.6
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1
	golang.org/x/sys v0.0.0-20211025201205-69cdffdb9359
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
)

require (
	github.com/Microsoft/go-winio v0.5.1 // indirect
	github.com/godbus/dbus/v5 v5.0.5 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/shopspring/decimal v1.3.0 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/text v0.3.7 // indirect
)

exclude github.com/sergi/go-diff v1.2.0 // Produces incorrect diffs
