<#
.SYNOPSIS
Install (and optionally run) chezmoi.

.PARAMETER BinDir
Specifies the installation directory. "./bin" is the default. Alias: b

.PARAMETER Tag
Specifies the version of chezmoi to install. "latest" is the default. Alias: t

.PARAMETER EnableDebug
If specified, print debug output. Alias: d

.PARAMETER ChezmoiArgs
If specified, execute chezmoi with these arguments after successful installation. This parameter can be provided
positionally without specifying its name.

.EXAMPLE
PS> install.ps1 -b '~/bin'
PS> iex "&{$(irm 'https://get.chezmoi.io/ps1')} -b '~/bin'"

.EXAMPLE
PS> install.ps1 -- init --apply <DOTFILE_REPO_URL>
PS> iex "&{$(irm 'https://get.chezmoi.io/ps1')} -- init --apply <DOTFILE_REPO_URL>"
#>
[CmdletBinding()]
param (
    [Parameter(Mandatory = $false)]
    [Alias('b')]
    [string]
    $BinDir = (Join-Path -Path (Resolve-Path -Path '.') -ChildPath 'bin'),

    [Parameter(Mandatory = $false)]
    [Alias('t')]
    [string]
    $Tag = 'latest',

    [Parameter(Mandatory = $false)]
    [Alias('d')]
    [switch]
    $EnableDebug,

    [Parameter(Position = 0, ValueFromRemainingArguments = $true)]
    [string[]]
    $ChezmoiArgs
)

function Write-DebugVariable {
    param (
        [string[]]$variables
    )
    foreach ($variable in $variables) {
        $debugVariable = Get-Variable -Name $variable
        Write-Debug "$( $debugVariable.Name ): $( $debugVariable.Value )"
    }
}

function Invoke-CleanUp ($directory) {
    if (($null -ne $directory) -and (Test-Path -Path $directory)) {
        Write-Debug "removing ${directory}"
        Remove-Item -Path $directory -Recurse -Force
    }
}

function Invoke-FileDownload ($uri, $path) {
    Write-Debug "downloading ${uri}"
    $wc = [System.Net.WebClient]::new()
    $wc.Headers.Add('Accept', 'application/octet-stream')
    $wc.DownloadFile($uri, $path)
    $wc.Dispose()
}

function Invoke-StringDownload ($uri) {
    Write-Debug "downloading ${uri} as string"
    $wc = [System.Net.WebClient]::new()
    $wc.DownloadString($uri)
    $wc.Dispose()
}

function Get-GoOS {
    if ($PSVersionTable.PSEdition -eq 'Desktop') {
        return 'windows'
    }

    $isOSPlatform = [System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform
    $osPlatform = [System.Runtime.InteropServices.OSPlatform]

    if ($isOSPlatform.Invoke($osPlatform::Windows)) { return 'windows' }
    if ($isOSPlatform.Invoke($osPlatform::Linux)) { return 'linux' }
    if ($isOSPlatform.Invoke($osPlatform::OSX)) { return 'darwin' }

    Write-Error 'unable to determine GOOS'
}

function Get-GoArch {
    if ($PSVersionTable.PSEdition -eq 'Core') {
        $goArch = @{
            'Arm'   = 'arm'
            'Arm64' = 'arm64'
            'X86'   = 'i386'
            'X64'   = 'amd64'
        }
        $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
        $result = $goArch[$arch]
        if (-not $result) {
            throw "Unsupported OS architecture: $arch"
        }
        return $result
    }

    $cpuArch = (Get-CimInstance -ClassName Win32_Processor).Architecture

    if ([System.Environment]::Is64BitOperatingSystem) {
        switch ($cpuArch) {
            9  { return 'amd64' }
            12 { return 'arm64' }
            default {
                throw "Unsupported CPU architecture ($cpuArch) on a 64-bit OS."
            }
        }
    } else {
        switch ($cpuArch) {
            0 { return 'i386' }
            9 { return 'i386' } # AMD64 CPU running 32-bit OS
            5 { return 'arm' }
            12 { return 'arm' } # ARM64 CPU running 32-bit OS
            default {
                throw "Unsupported CPU architecture ($cpuArch) on a 32-bit OS."
            }
        }
    }
}

function Get-RealTag ($tag) {
    Write-Debug "checking GitHub for tag ${tag}"
    $releaseUrl = "${BaseUrl}/${tag}"
    $json = try {
        Invoke-RestMethod -Uri $releaseUrl -Headers @{ 'Accept' = 'application/json' }
    } catch {
        Write-Error "error retrieving GitHub release ${tag}"
    }
    $realTag = $json.tag_name
    Write-Debug "found tag ${realTag} for ${tag}"
    return $realTag
}

function Get-LibC {
    $libcOutput = ''
    if (Get-Command -CommandType Application ldd -ErrorAction SilentlyContinue) {
        $libcOutput = (ldd --version 2>&1) -join [System.Environment]::NewLine
    } elseif (Get-Command -CommandType Application getconf -ErrorAction SilentlyContinue) {
        $libcOutput = (getconf GNU_LIBC_VERSION 2>&1) -join [System.Environment]::NewLine
    }
    Write-DebugVariable 'libcOutput'
    switch -Wildcard ($libcOutput) {
        '*glibc*' { return 'glibc' }
        '*gnu libc*' { return 'glibc' }
        '*musl*' { return 'musl' }
    }
    Write-Error 'unable to determine libc'
}

function Get-Checksums ($tag, $version) {
    $checksumsText = Invoke-StringDownload "${BaseUrl}/download/${tag}/chezmoi_${version}_checksums.txt"

    $checksums = @{}
    $lines = $checksumsText -split '\r?\n' | Where-Object { $_ }
    foreach ($line in $lines) {
        $value, $key = $line -split '\s+'
        $checksums[$key] = $value
    }
    $checksums
}

function Confirm-Checksum ($target, $checksums) {
    $basename = [System.IO.Path]::GetFileName($target)
    if (-not $checksums.ContainsKey($basename)) {
        Write-Error "unable to find checksum for ${target} in checksums"
    }
    $want = $checksums[$basename].ToLower()
    $got = (Get-FileHash -Path $target -Algorithm SHA256).Hash.ToLower()
    if ($want -ne $got) {
        Write-Error "checksum for ${target} did not verify ${want} vs ${got}"
    }
}

function Expand-ChezmoiArchive ($path) {
    $parent = Split-Path -Path $path -Parent
    Write-Debug "extracting ${path} to ${parent}"
    if ($path.EndsWith('.tar.gz')) {
        & tar --extract --gzip --file $path --directory $parent
    }
    if ($path.EndsWith('.zip')) {
        Expand-Archive -Path $path -DestinationPath $parent
    }
}

Set-StrictMode -Version 3.0

[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12

$script:ErrorActionPreference = 'Stop'
$script:InformationPreference = 'Continue'
if ($EnableDebug) {
    $script:DebugPreference = 'Continue'
}

trap {
    Invoke-CleanUp $tempDir
    break
}

$BaseUrl = 'https://github.com/twpayne/chezmoi/releases'

# convert $BinDir to an absolute path
$BinDir = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($BinDir)

$tempDir = ''
do {
    $tempDir = Join-Path -Path ([System.IO.Path]::GetTempPath()) -ChildPath ([System.Guid]::NewGuid())
} while (Test-Path -Path $tempDir)
New-Item -ItemType Directory -Path $tempDir | Out-Null

Write-DebugVariable 'BinDir', 'Tag', 'ChezmoiArgs', 'tempDir'

$goOS = Get-GoOS
$goArch = Get-GoArch
foreach ($variableName in @('goOS', 'goArch')) {
    Write-DebugVariable $variableName
}

$realTag = Get-RealTag $Tag
$version = $realTag.TrimStart('v')
Write-Information "found version ${version} for ${Tag}/${goOS}/${goArch}"

$binarySuffix = ''
$archiveFormat = 'tar.gz'
$goOSExtra = ''
switch ($goOS) {
    'linux' {
        $goOSExtra = "-$( Get-LibC )"
        break
    }
    'windows' {
        $binarySuffix = '.exe'
        $archiveFormat = 'zip'
        break
    }
}
Write-DebugVariable 'binarySuffix', 'archiveFormat', 'goOSExtra'

$archiveFilename = "chezmoi_${version}_${goOS}${goOSExtra}_${goArch}.${archiveFormat}"
$tempArchivePath = Join-Path -Path $tempDir -ChildPath $archiveFilename
Write-DebugVariable 'archiveFilename', 'tempArchivePath'
Invoke-FileDownload "${BaseUrl}/download/${realTag}/${archiveFilename}" $tempArchivePath

$checksums = Get-Checksums $realTag $version
Confirm-Checksum $tempArchivePath $checksums

Expand-ChezmoiArchive $tempArchivePath

$binaryFilename = "chezmoi${binarySuffix}"
$tempBinaryPath = Join-Path -Path $tempDir -ChildPath $binaryFilename
Write-DebugVariable 'binaryFilename', 'tempBinaryPath'
[System.IO.Directory]::CreateDirectory($BinDir) | Out-Null
$binary = Join-Path -Path $BinDir -ChildPath $binaryFilename
Write-DebugVariable 'binary'
Move-Item -Path $tempBinaryPath -Destination $binary -Force
Write-Information "installed ${binary}"

Invoke-CleanUp $tempDir

if (($null -ne $ChezmoiArgs) -and ($ChezmoiArgs.Count -gt 0)) {
    & $binary $ChezmoiArgs
}
