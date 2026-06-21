param(
  [string]$Message = "",
  [string]$Remote = "origin",
  [string]$Branch = "blog",
  [switch]$SkipAI,
  [switch]$SkipTests,
  [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Invoke-Checked {
  param(
    [string]$Label,
    [scriptblock]$Action
  )

  Write-Host ("==> " + $Label)
  & $Action
  if ($LASTEXITCODE -ne 0) {
    throw ($Label + " failed with exit code " + $LASTEXITCODE)
  }
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$repoName = Split-Path -Leaf $repoRoot
$buildOut = Join-Path ([System.IO.Path]::GetTempPath()) ("folio-publish-" + [guid]::NewGuid().ToString("N"))

Push-Location $repoRoot
try {
  Invoke-Checked "Checking git repository" { git rev-parse --is-inside-work-tree | Out-Null }

  $currentBranch = (git rev-parse --abbrev-ref HEAD).Trim()
  if (-not $SkipAI) {
    Invoke-Checked "Classifying changed posts with Codex" { & (Join-Path $PSScriptRoot "classify-posts.ps1") }
  }
  if (-not $SkipTests) {
    Invoke-Checked "Running tests" { go test ./... }
  }

  if (-not $SkipBuild) {
    Invoke-Checked "Running build check" { go run ./cmd/build -out $buildOut -base-path ("/" + $repoName) }
  }

  Invoke-Checked "Staging changes" { git add -A }

  & git diff --cached --quiet
  $hasStagedChanges = ($LASTEXITCODE -ne 0)

  if ($hasStagedChanges) {
    if ([string]::IsNullOrWhiteSpace($Message)) {
      $Message = "publish: " + (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
    }
    Invoke-Checked "Creating commit" { git commit -m $Message }
  } else {
    Write-Host "==> No staged changes to commit"
  }

  Invoke-Checked ("Pushing HEAD to " + $Remote + "/" + $Branch) { git push $Remote ("HEAD:" + $Branch) }

  Write-Host ""
  Write-Host ("Published from local branch: " + $currentBranch)
  Write-Host ("Remote target branch: " + $Branch)

  if (Test-Path "config.json") {
    try {
      $config = Get-Content "config.json" -Raw | ConvertFrom-Json
      if ($config.site_url) {
        Write-Host ("Site URL: " + $config.site_url)
      }
    } catch {
    }
  }
} finally {
  Pop-Location
  if (Test-Path $buildOut) {
    Remove-Item -LiteralPath $buildOut -Recurse -Force
  }
}
