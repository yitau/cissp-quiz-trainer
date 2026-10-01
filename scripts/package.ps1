$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$bin = Join-Path $root 'build/bin'
$exe = Join-Path $bin 'CISSPQuizTrainer.exe'
if (-not (Test-Path -LiteralPath $exe)) { throw 'Run wails build -clean first.' }
Copy-Item -LiteralPath (Join-Path $root 'samples/demo-questions.json') -Destination (Join-Path $bin 'demo-questions.json')
Copy-Item -LiteralPath (Join-Path $root 'docs/user-guide.md') -Destination (Join-Path $bin '使用说明.md')
$archive = Join-Path $bin 'CISSPQuizTrainer-win-x64.zip'
Compress-Archive -LiteralPath $exe,(Join-Path $bin 'demo-questions.json'),(Join-Path $bin '使用说明.md') -DestinationPath $archive -Force
$lines = @($exe,$archive) | ForEach-Object { $hash = Get-FileHash -LiteralPath $_ -Algorithm SHA256; '{0}  {1}' -f $hash.Hash.ToLowerInvariant(),(Split-Path $_ -Leaf) }
$lines | Set-Content -LiteralPath (Join-Path $bin 'checksums.txt') -Encoding utf8
Get-Item -LiteralPath $exe,$archive | Select-Object FullName,Length
