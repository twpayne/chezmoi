enum LogLevel {
    Debug = 3
    Info = 2
    Error = 1
    Critical = 0
}

$ErrorActionPreference = 'Stop' # throw an exception if anything bad happens

# If the environment isn't correct for running this script, try to give people
# some idea of how to fix it

if (($PSVersionTable.PSVersion.Major) -lt 5) {
    Write-Warning "PowerShell 5 or later is required to run this install script."
    Write-Warning "Please upgrade: https://docs.microsoft.com/en-us/powershell/scripting/setup/installing-windows-powershell"
    break
}

# show notification to change execution policy:
$allowedExecutionPolicy = @('Unrestricted', 'RemoteSigned', 'Bypass')
if ((Get-ExecutionPolicy).ToString() -notin $allowedExecutionPolicy) {
    Write-Warning "PowerShell requires an execution policy in [$($allowedExecutionPolicy -join ", ")] to run this install script."
    Write-Warning "For example, to set the execution policy to 'RemoteSigned' please run :"
    Write-Warning "'Set-ExecutionPolicy RemoteSigned -scope CurrentUser'"
    break
}

# globals
$tempdir = (Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName()));

# Helper functions
function Fetch-FileFromWeb(
    [string]$url,
    [string]$path
) {
    $cl = New-Object Net.WebClient
    $cl.Headers['User-Agent'] = 'System.Net.WebClient'
    $cl.DownloadFile($url, $path)
}

function Fetch-StringFromWeb(
    [string]$url
) {
    $cl = New-Object Net.WebClient
    $cl.Headers['User-Agent'] = 'System.Net.WebClient'
    return $cl.DownloadString($url)
}

function Fetch-JsonFromWeb(
    [string]$url
) {
    $cl = New-Object Net.WebClient
    $cl.Headers['User-Agent'] = 'System.Net.WebClient'
    $cl.Headers['Accept'] = 'application/json'
    return $cl.DownloadString($url) | ConvertFrom-Json
}

function Fetch-DataFromWeb(
    [string]$url
) {
    $cl = New-Object Net.WebClient
    $cl.Headers['User-Agent'] = 'System.Net.WebClient'
    return $cl.DownloadData($url)
}

function log {
    [CmdletBinding(PositionalBinding=$false)]
    param(
        [LogLevel] $MessageLevel,
        [string] $Message
    )

    if ([int]$LogLevel -ge [int]$MesageLevel) {
        Write-Host $Message
    }
}

function log-debug {
    param(
        [string] $Message
    )
    log -MessageLevel Debug -Message $Message
}

function log-info {
    param(
        [string] $Message
    )
    log -MessageLevel Info -Message $Message
}

function log-error {
    param(
        [string] $Message
    )
    log -MessageLevel Error -Message $Message
}

function log-critical {
    param(
        [string] $Message
    )
    log -MessageLevel Critical -Message $Message
}

function get_goos {
    $ri = [System.Runtime.InteropServices.RuntimeInformation]::IsOSPlatform;
    # if $ri isn't defined, then we must be running in Powershell 5.1, which only works on Windows.
    if (-not $ri -or $ri.Invoke([Runtime.InteropServices.OSPlatform]::Windows)) {
        return "windows"
    } elseif ($ri.Invoke([Runtime.InteropServices.OSPlatform]::Linux)) {
        return "linux"
    } elseif ($ri.Invoke([Runtime.InteropServices.OSPlatform]::OSX)) {
        return "darwin"
    } elseif ($ri.Invoke([Runtime.InteropServices.OSPlatform]::FreeBSD)) {
        return "freebsd"
    } else {
        throw "unsupported platform"
    }
}

function get_goarch {
    $arch = "$([Runtime.InteropServices.RuntimeInformation]::OSArchitecture)";
    $wmi_arch = $null;

    if ((-not $arch) -and [Environment]::Is64BitOperatingSystem) {
        $arch = "X64";
    }

    # [Environment]::Is64BitOperatingSystem is only available on .net 4 or
    # newer, so if we still don't know, try another method
    if (-not $arch) {
        if (Get-Command "Get-WmiObject" -ErrorAction SilentlyContinue) {
            $wmi_arch = (Get-WmiObject -Class Win32_OperatingSystem | Select-Object *).OSArchitecture
            if ($wmi_arch.StartsWith("64")) {
                $arch = "X64";
            } elseif ($wmi_arch.StartsWith("32")) {
                $arch = "X86";
            }
        }
    }

    switch ($arch) {
        "Arm" {
            return "arm"
        }
        "Arm64" {
            return "arm64"
        }
        "X86" {
            return "i386"
        }
        "X64" {
            return "amd64"
        }

        default {
            log-debug "Unsupported architecture: $arch (wmi: $wmi_arch)"
            throw "unsupported architecture"
        }
    }
}

function get-real-tag {
    param(
        [string] $tag
    )

    log-debug "checking GitHub for tag $tag"

    $release_url = "https://github.com/twpayne/chezmoi/releases/$tag"
    $real_tag = (Fetch-JsonFromWeb $release_url).tag_name

    log-debug "found tag $real_tag for $tag"
    return $real_tag
}

function verify-hash {
    param(
        [string] $target,
        [string] $checksums
    )

    $basename = [IO.Path]::GetFileName($target);

    # what checksum are we looking for?
    $want = (Get-Content $checksums | ForEach-Object {
        $line = $_;
        if ($line -match "$($basename)$") {
            $hash, $name = ($line -split "\s+");
            return $hash;
        }
    } | Select-Object -First 1).ToLower();

    $got = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash.ToLower();

    if ($want -ne $got) {
        Write-Error "Wanted: $want"
        Write-Error "Got: $got"
        throw "Checksum mismatch!"
    }
}

function unpack-file {
    param(
        [string] $file
    )

    if ($file.EndsWith(".tar.gz") -or $file.EndsWith(".tgz")) {
        tar -xzf $file
    } elseif ($file.EndsWith(".tar")) {
        tar -xf $file
    } elseif ($file.EndsWith(".zip")) {
        Expand-Archive -LiteralPath $file -DestinationPath .
    } else {
        throw "can't unpack unknown format for $file"
    }
}

<#
    .SYNOPSIS
    Install the Chezmoi dotfile manager

    .DESCRIPTION
    Installs Chezmoi to the given directory, defaulting to ./bin

    You can specify a particular git tag using the -Tag option.

    Examples:
    '$params = "-BinDir ~/bindir"', (iwr https://get.chezmoi.io/ps1).Content | powershell -c -
    '$params = "-Tag v1.8.10"', (iwr https://get.chezmoi.io/ps1).Content | powershell -c -
#>
function Install-Chezmoi {
    [CmdletBinding(PositionalBinding=$false)]
    param(
        [Parameter(Mandatory = $false)]
        [string]
        $BinDir = (Join-Path (Resolve-Path '.') 'bin'),

        [Parameter(Mandatory = $false)]
        [string] $Tag = 'latest',

        [LogLevel] $LogLevel = [LogLevel]::Info,

        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$ExecArgs
    )

    # some sub-functions (ie, get_goarch, likely others) require fetching of
    # non-existent properties to not error
    Set-StrictMode -off

    # $BinDir = Resolve-Path $BinDir

    $os = get_goos
    $arch = get_goarch
    $real_tag = get-real-tag $Tag
    $version = if ($real_tag.StartsWith("v")) { $real_tag.Substring(1) } else { $real_tag };

    log-info "found version $version for $Tag/$os/$arch"

    $binsuffix = ""
    $format = "tar.gz"

    if ($os -eq "windows") {
        $binsuffix = ".exe"
        $format = "zip"
    }

    $github_download = "https://github.com/twpayne/chezmoi/releases/download"
    New-Item -Type Directory -Path $tempdir | Out-Null

    # download tarball
    $name="chezmoi_$($version)_$($os)_$($arch)"
    $tarball="$name.$format"
    $tarball_url="$($github_download)/$real_tag/$tarball"

    $tmp_tarball = (Join-Path $tempdir $tarball)
    Fetch-FileFromWeb $tarball_url $tmp_tarball

    # download checksums
    $checksums = "chezmoi_$($version)_checksums.txt"
    $checksums_url = "$($github_download)/$($real_tag)/$($checksums)"

    $tmp_checksums = (Join-Path $tempdir $checksums)
    Fetch-FileFromWeb $checksums_url $tmp_checksums

    # verify checksums
    verify-hash $tmp_tarball $tmp_checksums

    Push-Location $tempdir

    unpack-file $tarball

    # install the binary
    if (-not (Test-Path $BinDir)) {
        New-Item -Type Directory -Path $BinDir | Out-Null
    }

    $binary = "chezmoi$($binsuffix)";
    $tmp_binary = (Join-Path $tempdir $binary);

    Move-Item -Force -Path $tmp_binary -Destination $BinDir

    log-info "Installed $($BinDir)/$($binary)"

    if ($ExecArgs) {
        & "$($BinDir)/$($binary)" $ExecArgs
    }
}

try {
    Invoke-Expression ("Install-Chezmoi " + $params)
} catch {
    Write-Host "An error occurred while installing: $_"
} finally {
    Pop-Location

    if (Test-Path $tempdir) {
        Remove-Item -LiteralPath $tempdir -Recurse -Force
    }
}
