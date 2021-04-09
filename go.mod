module github.com/twpayne/chezmoi

go 1.16

require (
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20201120212035-bb82daffcca2 // indirect
	github.com/bmatcuk/doublestar/v3 v3.0.0
	github.com/charmbracelet/glamour v0.3.0
	github.com/coreos/go-semver v0.3.0
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-git/go-git/v5 v5.3.0
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-github/v34 v34.0.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/renameio v1.0.0
	github.com/google/uuid v1.2.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/mitchellh/copystructure v1.1.2 // indirect
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
	github.com/zalando/go-keyring v0.1.1
	go.etcd.io/bbolt v1.3.5
	go.uber.org/multierr v1.6.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	golang.org/x/sys v0.0.0-20210403161142-5e06dd20ab57
	golang.org/x/term v0.0.0-20210406210042-72f3dc4e9b72
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
)

exclude github.com/sergi/go-diff v1.2.0 // Produces incorrect diffs
