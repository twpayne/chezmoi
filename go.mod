module github.com/twpayne/chezmoi

go 1.24.4

tool (
	github.com/twpayne/chezmoi
	github.com/twpayne/chezmoi/internal/cmds/execute-template
	github.com/twpayne/chezmoi/internal/cmds/generate-commit
	github.com/twpayne/chezmoi/internal/cmds/generate-helps
	github.com/twpayne/chezmoi/internal/cmds/generate-install.sh
	github.com/twpayne/chezmoi/internal/cmds/generate-license
	github.com/twpayne/chezmoi/internal/cmds/lint-commit-messages
	github.com/twpayne/chezmoi/internal/cmds/lint-txtar
	github.com/twpayne/chezmoi/internal/cmds/lint-whitespace
	github.com/twpayne/find-typos
)

require (
	filippo.io/age v1.2.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.10.1
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets v1.4.0
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/Shopify/ejson v1.5.4
	github.com/alecthomas/assert/v2 v2.11.0
	github.com/aws/aws-sdk-go-v2 v1.36.4
	github.com/aws/aws-sdk-go-v2/config v1.29.16
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.35.6
	github.com/bmatcuk/doublestar/v4 v4.8.1
	github.com/bradenhilton/mozillainstallhash v1.0.1
	github.com/charmbracelet/bubbles v0.20.0
	github.com/charmbracelet/bubbletea v1.3.5
	github.com/charmbracelet/glamour v0.10.0
	github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834
	github.com/coreos/go-semver v0.3.1
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-git/go-git/v5 v5.16.2
	github.com/go-viper/mapstructure/v2 v2.2.1
	github.com/goccy/go-yaml v1.18.0
	github.com/google/go-github/v61 v61.0.0
	github.com/google/renameio/v2 v2.0.0
	github.com/gopasspw/gopass v1.15.16
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/itchyny/gojq v0.12.17
	github.com/klauspost/compress v1.18.0
	github.com/mattn/go-runewidth v0.0.16
	github.com/mitchellh/copystructure v1.2.0
	github.com/muesli/combinator v0.3.0
	github.com/muesli/termenv v0.16.0
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/rogpeppe/go-internal v1.14.1
	github.com/sergi/go-diff v1.4.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/tailscale/hujson v0.0.0-20250605163823-992244df8c5a
	github.com/tobischo/gokeepasslib/v3 v3.6.1
	github.com/twpayne/go-expect v0.0.2-0.20241130000624-916db2914efd
	github.com/twpayne/go-pinentry/v4 v4.0.0
	github.com/twpayne/go-shell v0.5.0
	github.com/twpayne/go-vfs/v5 v5.0.4
	github.com/twpayne/go-xdg/v6 v6.1.3
	github.com/ulikunitz/xz v0.5.12
	github.com/zalando/go-keyring v0.2.6
	github.com/zricethezav/gitleaks/v8 v8.27.2
	go.etcd.io/bbolt v1.4.0
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/crypto v0.39.0
	golang.org/x/oauth2 v0.30.0
	golang.org/x/sync v0.15.0
	golang.org/x/sys v0.33.0
	golang.org/x/term v0.32.0
	golang.org/x/text v0.26.0
	gopkg.in/ini.v1 v1.67.0
	howett.net/plist v1.0.1
	mvdan.cc/sh/v3 v3.11.0
)

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/BobuSumisu/aho-corasick v1.0.3 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/STARRY-S/zip v0.2.3 // indirect
	github.com/alecthomas/chroma/v2 v2.18.0 // indirect
	github.com/alecthomas/repr v0.4.0 // indirect
	github.com/andybalholm/brotli v1.1.2-0.20250424173009-453214e765f3 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.69 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.31 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.21 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/bradenhilton/cityhash v1.0.0 // indirect
	github.com/caspr-io/yamlpath v0.0.0-20200722075116-502e8d113a9b // indirect
	github.com/charmbracelet/colorprofile v0.3.1 // indirect
	github.com/charmbracelet/harmonica v0.2.0 // indirect
	github.com/charmbracelet/x/ansi v0.9.2 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13 // indirect
	github.com/charmbracelet/x/exp/slice v0.0.0-20250611152503-f53cdd7e01ef // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/creack/pty/v2 v2.0.0-20231209135443-03db72c7b76c // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dustin/gojson v0.0.0-20160307161227-2e71ec9dd5ad // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fatih/semgroup v1.3.0 // indirect
	github.com/gitleaks/go-gitdiff v0.9.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopasspw/gitconfig v0.0.1 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mholt/archives v0.1.2 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27 // indirect
	github.com/minio/minlz v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/nwaples/rardecode/v2 v2.1.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pjbgf/sha1cd v0.3.2 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/sorairolake/lzip-go v0.3.7 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/viper v1.20.1 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/tobischo/argon2 v0.1.0 // indirect
	github.com/twpayne/find-typos v0.0.3 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/wasilibs/go-re2 v1.10.0 // indirect
	github.com/wasilibs/wazero-helpers v0.0.0-20250123031827-cd30c44769bb // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/yuin/goldmark v1.7.12 // indirect
	github.com/yuin/goldmark-emoji v1.0.6 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/exp v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude (
	github.com/charmbracelet/bubbles v0.21.0 // https://github.com/twpayne/chezmoi/issues/4405
	go.etcd.io/bbolt v1.4.1
)
