# Start each service in a new window
$services = @(
    @{
        Name = "Master Service"
        Path = "services/master"
        Color = "Green"
    },
    @{
        Name = "World Generator"
        Path = "services/worldgen"
        Color = "Cyan"
    },
    @{
        Name = "Logging Service"
        Path = "services/logging"
        Color = "Yellow"
    },
    @{
        Name = "AI Service"
        Path = "services/ai"
        Color = "Magenta"
    }
)

# Create a function to start a service
function Start-GameService {
    param (
        [string]$Name,
        [string]$Path,
        [string]$Color
    )
    
    $workingDir = Join-Path $PSScriptRoot $Path
    Start-Process powershell.exe -ArgumentList "-NoExit", "-Command", "Write-Host 'Starting $Name' -ForegroundColor $Color; cd '$workingDir'; go run main.go"
}

# Start all services
foreach ($service in $services) {
    Start-GameService -Name $service.Name -Path $service.Path -Color $service.Color
}

# Wait for services to start
Write-Host "Waiting for services to initialize (10 seconds)..." -ForegroundColor White
Start-Sleep -Seconds 10

# Command window script
$commandScript = @'
Write-Host "Text Adventure Game Command Console" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Green
Write-Host "Starting game..." -ForegroundColor Yellow

# Initialize game state
Write-Host "Initializing game state..." -ForegroundColor Cyan
curl.exe -X POST http://localhost:8080/api/v1/game-state -d "{\\"name\\": \\"Brave Adventurer\\"}"

Write-Host "`nTry these commands:" -ForegroundColor Yellow
Write-Host "1. Look around:" -ForegroundColor White
Write-Host "curl.exe -X POST http://localhost:8080/api/v1/process-input -d {\\"input\\": \\"look around\\"}" -ForegroundColor Gray

Write-Host "`n2. Check inventory:" -ForegroundColor White
Write-Host "curl.exe -X POST http://localhost:8080/api/v1/process-input -d {\\"input\\": \\"inventory\\"}" -ForegroundColor Gray

Write-Host "`n3. Move around:" -ForegroundColor White
Write-Host "curl.exe -X POST http://localhost:8080/api/v1/process-input -d {\\"input\\": \\"go north\\"}" -ForegroundColor Gray

Write-Host "`n4. Check game state:" -ForegroundColor White
Write-Host "curl.exe -X GET http://localhost:8080/api/v1/game-state" -ForegroundColor Gray

Write-Host "`nHappy adventuring!" -ForegroundColor Green
'@

# Start the command window
Start-Process powershell.exe -ArgumentList "-NoExit", "-Command", $commandScript
