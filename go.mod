module github.com/twpayne/chezmoi/v2

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/alecthomas/chroma v0.9.1 // indirect
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20201120212035-bb82daffcca2 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/charmbracelet/glamour v0.3.0
	github.com/coreos/go-semver v0.3.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-git/go-git/v5 v5.3.0
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-github/v35 v35.2.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/renameio v1.0.1
	github.com/google/uuid v1.2.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/microcosm-cc/bluemonday v1.0.9 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.1
	github.com/muesli/combinator v0.3.0
	github.com/pelletier/go-toml v1.9.0
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0
	github.com/rs/zerolog v1.21.0
	github.com/sergi/go-diff v1.1.0
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/twpayne/go-shell v0.3.0
	github.com/twpayne/go-vfs/v2 v2.0.0
	github.com/twpayne/go-vfsafero/v2 v2.0.0
	github.com/twpayne/go-xdg/v4 v4.0.0
	github.com/yuin/goldmark v1.3.5 // indirect
	github.com/zalando/go-keyring v0.1.1
	go.etcd.io/bbolt v1.3.5
	go.uber.org/multierr v1.7.0
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf // indirect
	golang.org/x/net v0.0.0-20210508051633-16afe75a6701 // indirect
	golang.org/x/oauth2 v0.0.0-20210427180440-81ed05c6b58c
	golang.org/x/sys v0.0.0-20210507161434-a76c4d0a0096
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
)

exclude github.com/sergi/go-diff v1.2.0 // Produces incorrect diffs
