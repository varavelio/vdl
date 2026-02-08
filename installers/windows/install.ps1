<# 
.SYNOPSIS
  VDL Installer Script (Windows)

.DESCRIPTION
  Downloads and installs the VDL binary for Windows.

.EXAMPLE
  irm https://get.varavel.com/vdl.ps1 | iex

.EXAMPLE
  $env:VERSION = "v0.4.0"; irm https://get.varavel.com/vdl.ps1 | iex

.EXAMPLE
  $env:INSTALL_DIR = "$HOME\.local\bin"; $env:QUIET = "true"; irm https://get.varavel.com/vdl.ps1 | iex

.NOTES
  Options (environment variables):
    VERSION      : Version to install (e.g., v0.4.0). Defaults to "latest".
    INSTALL_DIR  : Install directory. Defaults to "$env:LOCALAPPDATA\Programs\vdl".
    QUIET        : Set to "true" to suppress output.
#>

$ErrorActionPreference = "Stop"

$script:Repo = "varavelio/vdl"
$script:BinaryName = "vdl"
$script:InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\vdl" }
$script:Version = if ($env:VERSION) { $env:VERSION } else { "" }
$script:Quiet = $env:QUIET -eq "true"
$script:TmpDir = $null
$script:UseColors = $Host.UI.SupportsVirtualTerminal -and -not $script:Quiet

function Write-Info($Message) {
  if (-not $script:Quiet) {
    if ($script:UseColors) {
      Write-Host "[INFO] " -ForegroundColor Green -NoNewline
      Write-Host $Message
    } else {
      Write-Host "[INFO] $Message"
    }
  }
}

function Write-Warn($Message) {
  if (-not $script:Quiet) {
    if ($script:UseColors) {
      Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
      Write-Host $Message
    } else {
      Write-Host "[WARN] $Message"
    }
  }
}

function Write-Err($Message) {
  if (-not $script:Quiet) {
    if ($script:UseColors) {
      Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
      Write-Host $Message
    } else {
      Write-Host "[ERROR] $Message"
    }
  }
}

function Show-Banner {
  if (-not $script:Quiet) {
    $banner = @"
██╗  ██╗█████╗ ██╗
██║  ██║██╔═██╗██║
╚██╗██╔╝██║ ██║██║
 ╚███╔╝ █████╔╝█████╗
  ╚══╝  ╚════╝ ╚════╝
"@
    if ($script:UseColors) {
      Write-Host $banner.TrimStart() -ForegroundColor Blue
    } else {
      Write-Host $banner.TrimStart()
    }
  }
}

function Invoke-Cleanup {
  if ($script:TmpDir -and (Test-Path $script:TmpDir)) {
    Remove-Item -Recurse -Force $script:TmpDir -ErrorAction SilentlyContinue
  }
}

function Get-Platform {
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLower()
  switch ($arch) {
    "x64" { return "amd64" }
    "arm64" { return "arm64" }
    default {
      Write-Err "Unsupported architecture: $arch"
      exit 1
    }
  }
}

function Get-LatestVersion {
  if ([string]::IsNullOrEmpty($script:Version) -or $script:Version -eq "latest") {
    Write-Info "Fetching latest version..."
    try {
      $null = Invoke-WebRequest -Uri "https://github.com/$script:Repo/releases/latest" -Method Head -MaximumRedirection 0 -ErrorAction Stop -UseBasicParsing
    } catch {
      $response = $_.Exception.Response
      if ($response -and $response.Headers.Location) {
        $location = $response.Headers.Location
        if ($location -is [array]) { $location = $location[0] }
        if ($location -match "/tag/v?(.+)$") {
          $script:Version = "v$($Matches[1])"
        }
      }
    }

    if ([string]::IsNullOrEmpty($script:Version)) {
      Write-Err "Failed to determine latest version."
      exit 1
    }
  }

  if (-not $script:Version.StartsWith("v")) {
    $script:Version = "v$($script:Version)"
  }
}

function Test-IsAdmin {
  $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
  $principal = New-Object Security.Principal.WindowsPrincipal($identity)
  return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Install-Vdl {
  $script:TmpDir = Join-Path $env:TEMP "vdl-install-$(Get-Random)"
  New-Item -ItemType Directory -Path $script:TmpDir -Force | Out-Null

  try {
    $arch = Get-Platform
    $filename = "${script:BinaryName}_windows_${arch}.zip"
    $downloadUrl = "https://github.com/$script:Repo/releases/download/$script:Version/$filename"
    $checksumsUrl = "https://github.com/$script:Repo/releases/download/$script:Version/checksums.txt"
    $zipPath = Join-Path $script:TmpDir $filename
    $checksumsPath = Join-Path $script:TmpDir "checksums.txt"

    Write-Info "Installing version: $script:Version"
    Write-Info "Downloading $filename..."

    try {
      Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing
      Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing
    } catch {
      Write-Err "Download failed: $_"
      exit 1
    }

    Write-Info "Verifying checksum..."
    $expectedLine = Get-Content $checksumsPath | Where-Object { $_ -match $filename }
    if ($expectedLine) {
      $expectedHash = ($expectedLine -split "\s+")[0].ToUpper()
      $actualHash = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToUpper()
      if ($expectedHash -ne $actualHash) {
        Write-Err "Checksum verification failed!"
        exit 1
      }
    } else {
      Write-Warn "Checksum entry not found. Skipping verification."
    }

    Write-Info "Extracting..."
    Expand-Archive -Path $zipPath -DestinationPath $script:TmpDir -Force

    $binSource = Join-Path $script:TmpDir "$script:BinaryName.exe"
    if (-not (Test-Path $binSource)) {
      Write-Err "Binary not found in archive."
      exit 1
    }

    Write-Info "Installing to $script:InstallDir..."

    # Create directory if needed
    if (-not (Test-Path $script:InstallDir)) {
      try {
        New-Item -ItemType Directory -Path $script:InstallDir -Force | Out-Null
      } catch {
        if (Test-IsAdmin) {
          Write-Err "Cannot create directory: $script:InstallDir"
          exit 1
        } else {
          Write-Err "Cannot create directory. Try running as Administrator or use a user-writable path."
          Write-Info "Example: `$env:INSTALL_DIR = `"`$HOME\.local\bin`"; irm ... | iex"
          exit 1
        }
      }
    }

    # Copy binary
    $binDest = Join-Path $script:InstallDir "$script:BinaryName.exe"
    try {
      Copy-Item -Path $binSource -Destination $binDest -Force
    } catch {
      if (-not (Test-IsAdmin)) {
        Write-Err "Cannot write to $script:InstallDir. Try running as Administrator."
        exit 1
      }
      throw
    }

    # Update PATH
    $pathScope = "User"
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", $pathScope)
    if ($currentPath -notlike "*$script:InstallDir*") {
      Write-Info "Adding $script:InstallDir to PATH..."
      $newPath = if ($currentPath) { "$currentPath;$script:InstallDir" } else { $script:InstallDir }
      try {
        [Environment]::SetEnvironmentVariable("PATH", $newPath, $pathScope)
        $env:PATH = "$env:PATH;$script:InstallDir"
      } catch {
        Write-Warn "Could not update PATH automatically. Add manually: $script:InstallDir"
      }
    }

    Write-Info "Installation complete!"
    Write-Info "Run '$script:BinaryName --version' to verify."
    Write-Info "Restart your terminal to use vdl."

  } finally {
    Invoke-Cleanup
  }
}

Show-Banner
Get-LatestVersion
Install-Vdl
