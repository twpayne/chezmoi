module chezmoi.io/chezmoi

go 1.25.7

tool (
	chezmoi.io/chezmoi
	chezmoi.io/chezmoi/internal/cmds/execute-template
	chezmoi.io/chezmoi/internal/cmds/generate-commit
	chezmoi.io/chezmoi/internal/cmds/generate-helps
	chezmoi.io/chezmoi/internal/cmds/generate-install.sh
	chezmoi.io/chezmoi/internal/cmds/generate-license
	chezmoi.io/chezmoi/internal/cmds/hexencode
	chezmoi.io/chezmoi/internal/cmds/lint-commit-messages
	chezmoi.io/chezmoi/internal/cmds/lint-txtar
	chezmoi.io/chezmoi/internal/cmds/lint-whitespace
	github.com/editorconfig-checker/editorconfig-checker/v3/cmd/editorconfig-checker
	github.com/google/capslock/cmd/capslock
	github.com/josephspurrier/goversioninfo/cmd/goversioninfo
	github.com/rhysd/actionlint/cmd/actionlint
	github.com/twpayne/find-typos
	github.com/twpayne/go-jsonstruct/v3/cmd/gojsonstruct
)

require (
	filippo.io/age v1.3.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets v1.4.0
	github.com/BurntSushi/toml v1.6.0
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/Shopify/ejson v1.5.4
	github.com/alecthomas/assert/v2 v2.11.0
	github.com/aws/aws-sdk-go-v2 v1.41.2
	github.com/aws/aws-sdk-go-v2/config v1.32.10
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.41.2
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/bradenhilton/mozillainstallhash v1.0.1
	github.com/charmbracelet/bubbles v0.20.0
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/glamour v0.10.0
	github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834
	github.com/coreos/go-semver v0.3.1
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-git/go-git/v5 v5.17.0
	github.com/go-sprout/sprout v1.0.3
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/goccy/go-yaml v1.19.2
	github.com/google/go-github/v61 v61.0.0
	github.com/google/renameio/v2 v2.0.2
	github.com/gopasspw/gopass v1.16.1
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/itchyny/gojq v0.12.18
	github.com/klauspost/compress v1.18.4
	github.com/mattn/go-runewidth v0.0.20
	github.com/mitchellh/copystructure v1.2.0
	github.com/muesli/combinator v0.3.0
	github.com/muesli/termenv v0.16.0
	github.com/nwaples/rardecode/v2 v2.2.2
	github.com/rogpeppe/go-internal v1.14.1
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/tailscale/hujson v0.0.0-20250605163823-992244df8c5a
	github.com/tobischo/gokeepasslib/v3 v3.6.2
	github.com/twpayne/go-expect v0.0.2-0.20241130000624-916db2914efd
	github.com/twpayne/go-pinentry/v4 v4.0.1
	github.com/twpayne/go-shell v0.5.0
	github.com/twpayne/go-vfs/v5 v5.0.5
	github.com/twpayne/go-xdg/v6 v6.1.3
	github.com/ulikunitz/xz v0.5.15
	github.com/zalando/go-keyring v0.2.6
	github.com/zricethezav/gitleaks/v8 v8.30.0
	go.etcd.io/bbolt v1.4.3
	golang.org/x/crypto v0.48.0
	golang.org/x/oauth2 v0.35.0
	golang.org/x/sync v0.19.0
	golang.org/x/sys v0.41.0
	golang.org/x/term v0.40.0
	golang.org/x/text v0.34.0
	gopkg.in/ini.v1 v1.67.1
	howett.net/plist v1.0.1
	mvdan.cc/sh/v3 v3.12.0
	znkr.io/diff v1.0.0-beta.4
)

require (
	al.essio.dev/pkg/shellescape v1.6.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	filippo.io/hpke v0.4.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.21.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.2.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/BobuSumisu/aho-corasick v1.0.3 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/STARRY-S/zip v0.2.3 // indirect
	github.com/akavel/rsrc v0.10.2 // indirect
	github.com/alecthomas/chroma/v2 v2.23.1 // indirect
	github.com/alecthomas/repr v0.5.2 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.18 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.7 // indirect
	github.com/aws/smithy-go v1.24.1 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/bradenhilton/cityhash v1.0.0 // indirect
	github.com/caspr-io/yamlpath v0.0.0-20200722075116-502e8d113a9b // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/charmbracelet/colorprofile v0.4.2 // indirect
	github.com/charmbracelet/harmonica v0.2.0 // indirect
	github.com/charmbracelet/x/ansi v0.11.6 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/exp/slice v0.0.0-20260225200202-61df8bc4b903 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/creack/pty/v2 v2.0.0-20231209135443-03db72c7b76c // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/danieljoos/wincred v1.2.3 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/dustin/gojson v0.0.0-20160307161227-2e71ec9dd5ad // indirect
	github.com/editorconfig-checker/editorconfig-checker/v3 v3.6.1 // indirect
	github.com/editorconfig/editorconfig-core-go/v2 v2.6.4 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fatih/semgroup v1.3.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/gitleaks/go-gitdiff v0.9.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.8.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/capslock v0.3.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gopasspw/gitconfig v0.0.4 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.7 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/josephspurrier/goversioninfo v1.5.0 // indirect
	github.com/kevinburke/ssh_config v1.6.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mholt/archives v0.1.5 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27 // indirect
	github.com/mikelolasagasti/xz v1.0.1 // indirect
	github.com/minio/minlz v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/pjbgf/sha1cd v0.5.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/rhysd/actionlint v1.7.11 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/skeema/knownhosts v1.3.2 // indirect
	github.com/sorairolake/lzip-go v0.3.8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tetratelabs/wazero v1.11.0 // indirect
	github.com/tobischo/argon2 v0.1.0 // indirect
	github.com/twpayne/find-typos v0.0.3 // indirect
	github.com/twpayne/go-jsonstruct/v3 v3.3.0 // indirect
	github.com/urfave/cli/v2 v2.27.7 // indirect
	github.com/wasilibs/go-re2 v1.10.0 // indirect
	github.com/wasilibs/wazero-helpers v0.0.0-20250123031827-cd30c44769bb // indirect
	github.com/wlynxg/chardet v1.0.4 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/xrash/smetrics v0.0.0-20250705151800-55b8f293f342 // indirect
	github.com/yuin/goldmark v1.7.16 // indirect
	github.com/yuin/goldmark-emoji v1.0.6 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.3 // indirect
	go4.org v0.0.0-20260112195520-a5071408f32f // indirect
	golang.org/x/exp v0.0.0-20260218203240-3dfff04db8fa // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude (
	github.com/charmbracelet/bubbles v0.21.0 // https://github.com/twpayne/chezmoi/issues/4405
	github.com/charmbracelet/bubbles v0.21.1 // https://github.com/twpayne/chezmoi/issues/4405
	github.com/charmbracelet/bubbles v1.0.0 // https://github.com/twpayne/chezmoi/issues/4405
)
