Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent (Split-Path -Parent $scriptDir)
$relayScript = Join-Path $repoRoot "scripts\sepolia_rpc_relay.py"
$pythonExe = (Get-Command python).Source
$sshExe = "$env:WINDIR\System32\OpenSSH\ssh.exe"
$stateDir = Join-Path $env:ProgramData "bnb-sepolia-relay"
$logDir = Join-Path $stateDir "logs"
$relayStdout = Join-Path $logDir "relay.stdout.log"
$relayStderr = Join-Path $logDir "relay.stderr.log"
$tunnelStdout = Join-Path $logDir "tunnel.stdout.log"
$tunnelStderr = Join-Path $logDir "tunnel.stderr.log"

New-Item -ItemType Directory -Force -Path $logDir | Out-Null

function Test-LoopbackPortListening {
    param(
        [int]$Port
    )

    $listener = Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue
    return $null -ne $listener
}

function Find-ProcessByCommandLine {
    param(
        [string]$Pattern
    )

    Get-CimInstance Win32_Process |
        Where-Object { $_.CommandLine -and $_.CommandLine -like "*$Pattern*" }
}

if (-not (Test-Path $relayScript)) {
    throw "Relay script not found: $relayScript"
}

if (-not (Test-LoopbackPortListening -Port 28545)) {
    Start-Process -FilePath $pythonExe `
        -ArgumentList @(
            $relayScript,
            "--listen-host", "127.0.0.1",
            "--listen-port", "28545",
            "--upstream", "https://sepolia.infura.io/v3/aa4778679f4e4e64a48621c2b6c0c8b8",
            "--timeout", "120"
        ) `
        -RedirectStandardOutput $relayStdout `
        -RedirectStandardError $relayStderr `
        -WindowStyle Hidden
}

$tunnelPattern = "-R 127.0.0.1:28545:127.0.0.1:28545 bnb-server"
if (-not (Find-ProcessByCommandLine -Pattern $tunnelPattern)) {
    Start-Process -FilePath $sshExe `
        -ArgumentList @(
            "-N",
            "-o", "ServerAliveInterval=30",
            "-o", "ServerAliveCountMax=3",
            "-o", "ExitOnForwardFailure=yes",
            "-R", "127.0.0.1:28545:127.0.0.1:28545",
            "bnb-server"
        ) `
        -RedirectStandardOutput $tunnelStdout `
        -RedirectStandardError $tunnelStderr `
        -WindowStyle Hidden
}
