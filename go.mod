module github.com/twpayne/chezmoi/v2

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/ProtonMail/go-crypto v0.0.0-20210707164159-52430bf6b52c // indirect
	github.com/alecthomas/chroma v0.9.2 // indirect
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20201120212035-bb82daffcca2 // indirect
	github.com/bmatcuk/doublestar/v4 v4.0.2
	github.com/bradenhilton/mozillainstallhash v1.0.0
	github.com/charmbracelet/glamour v0.3.0
	github.com/coreos/go-semver v0.3.0
	github.com/danieljoos/wincred v1.1.1 // indirect
	github.com/go-git/go-git/v5 v5.4.2
	github.com/google/go-github/v36 v36.0.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gops v0.3.19
	github.com/google/renameio v1.0.1
	github.com/google/uuid v1.3.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.15 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/muesli/combinator v0.3.0
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.9.0 // indirect
	github.com/pelletier/go-toml v1.9.3
	github.com/rogpeppe/go-internal v1.8.0
	github.com/rs/zerolog v1.23.0
	github.com/sergi/go-diff v1.1.0
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spf13/afero v1.6.0
	github.com/spf13/cast v1.4.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/twpayne/go-shell v0.3.0
	github.com/twpayne/go-vfs/v3 v3.0.0
	github.com/twpayne/go-xdg/v6 v6.0.0
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	github.com/yuin/goldmark v1.4.0 // indirect
	github.com/zalando/go-keyring v0.1.1
	go.etcd.io/bbolt v1.3.6
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
)

exclude github.com/sergi/go-diff v1.2.0 // Produces incorrect diffs
