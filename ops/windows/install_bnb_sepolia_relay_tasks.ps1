Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ensureScript = Join-Path $scriptDir "ensure_bnb_sepolia_relay.ps1"

if (-not (Test-Path $ensureScript)) {
    throw "Ensure script not found: $ensureScript"
}

$taskNameBoot = "BNB Sepolia Relay Boot"
$taskNameHeal = "BNB Sepolia Relay Heal"
$taskCommand = "powershell.exe -NoProfile -ExecutionPolicy Bypass -File `"$ensureScript`""

schtasks /Create /F /TN $taskNameBoot /SC ONLOGON /RL HIGHEST /TR $taskCommand | Out-Null
schtasks /Create /F /TN $taskNameHeal /SC MINUTE /MO 5 /RL HIGHEST /TR $taskCommand | Out-Null
schtasks /Run /TN $taskNameBoot | Out-Null
