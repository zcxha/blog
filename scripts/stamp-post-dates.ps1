param(
  [string[]]$Paths = @(),
  [switch]$All
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest
$script:OutputEncoding = [System.Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

function Get-ChangedHTMLPostPaths {
  $changed = @()
  & git rev-parse --verify HEAD *> $null
  if ($LASTEXITCODE -eq 0) {
    $changed += @(& git diff --name-only --diff-filter=ACMR HEAD -- posts)
  }
  $changed += @(& git ls-files --others --exclude-standard -- posts)
  return @($changed | Where-Object { $_ -match '^posts[\\/].+\.html$' } | Sort-Object -Unique)
}

function ConvertTo-YamlJSONString {
  param([string]$Value)
  return ($Value | ConvertTo-Json -Compress)
}

function Get-FileDateStamp {
  param([string]$Path)

  $item = Get-Item -LiteralPath $Path
  $date = [DateTimeOffset]::new($item.LastWriteTime)
  return $date.ToString("yyyy-MM-ddTHH:mm:sszzz")
}

function Add-PostDateIfMissing {
  param([string]$Path)

  $content = [System.IO.File]::ReadAllText($Path, [System.Text.UTF8Encoding]::new($false))
  $newline = if ($content.Contains("`r`n")) { "`r`n" } else { "`n" }
  $normalized = $content.Replace("`r`n", "`n")
  $lines = [System.Collections.Generic.List[string]]::new()
  $normalized.Split("`n") | ForEach-Object { $lines.Add($_) }

  $dateLine = "date: " + (ConvertTo-YamlJSONString (Get-FileDateStamp $Path))

  if ($lines.Count -gt 0 -and $lines[0].TrimStart([char]0xFEFF) -eq "---") {
    $end = -1
    for ($i = 1; $i -lt $lines.Count; $i++) {
      if ($lines[$i].Trim() -eq "---") { $end = $i; break }
    }
    if ($end -lt 0) { throw "Unclosed front matter: $Path" }

    for ($i = 1; $i -lt $end; $i++) {
      if ($lines[$i] -match '^date\s*:') { return $false }
    }

    $lines.Insert($end, $dateLine)
  } else {
    $prefix = @("---", $dateLine, "---", "")
    for ($i = $prefix.Count - 1; $i -ge 0; $i--) { $lines.Insert(0, $prefix[$i]) }
  }

  $updated = [string]::Join($newline, $lines)
  [System.IO.File]::WriteAllText($Path, $updated, [System.Text.UTF8Encoding]::new($false))
  return $true
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
Push-Location $repoRoot
try {
  if ($All) {
    $Paths = @(Get-ChildItem posts -File -Filter *.html | ForEach-Object { "posts/" + $_.Name })
  } elseif ($Paths.Count -eq 0) {
    $Paths = @(Get-ChangedHTMLPostPaths)
  }

  $Paths = @($Paths | ForEach-Object { $_.Replace('\', '/') } | Where-Object { Test-Path -LiteralPath $_ } | Sort-Object -Unique)
  if ($Paths.Count -eq 0) {
    Write-Host "==> Date stamping: no changed HTML posts"
    return
  }

  $stamped = @()
  foreach ($path in $Paths) {
    if ([System.IO.Path]::GetExtension($path) -ne ".html") { continue }
    if (Add-PostDateIfMissing -Path $path) {
      $stamped += $path
    }
  }

  if ($stamped.Count -eq 0) {
    Write-Host "==> Date stamping: all HTML posts already have dates"
    return
  }

  Write-Host ("==> Date stamping: " + $stamped.Count + " HTML post(s)")
  foreach ($path in $stamped) {
    Write-Host ("    " + $path)
  }
} finally {
  Pop-Location
}
