#Requires -Version 5.1
<#
.SYNOPSIS
  Forma local development launcher (Windows PowerShell).

.EXAMPLE
  .\scripts\forma\local\forma-local.ps1 doctor
  .\scripts\forma\local\forma-local.ps1 start
  .\scripts\forma\local\forma-local.ps1 status
  .\scripts\forma\local\forma-local.ps1 logs backend
  .\scripts\forma\local\forma-local.ps1 stop
  .\scripts\forma\local\forma-local.ps1 stop --all
#>
[CmdletBinding()]
param(
  [Parameter(Position = 0)]
  [string]$Command = 'help',

  [Parameter(ValueFromRemainingArguments = $true)]
  [string[]]$Rest
)

$ErrorActionPreference = 'Stop'
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Core = Join-Path $ScriptDir 'forma-local.mjs'

if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
  Write-Host 'BLOCKED: Node.js is required (rush.json requires >=21).' -ForegroundColor Red
  exit 2
}

if (-not (Test-Path -LiteralPath $Core)) {
  Write-Host "BLOCKED: missing shared launcher: $Core" -ForegroundColor Red
  exit 2
}

$argsList = @($Core, $Command) + @($Rest)
& node @argsList
exit $LASTEXITCODE
