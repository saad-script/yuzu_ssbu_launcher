Remove-Item '.\build' -Recurse -Force

Copy-Item -Path ".\icons" -Destination ".\build\yuzu_ssbu_launcher\icons" -Recurse -Force

# get rsrc from https://github.com/akavel/rsrc
.\rsrc_windows_amd64.exe -ico ".\icons\ssbu_red.ico" -o ".\rsrc.syso"

go build -o ".\yuzu_ssbu_launcher.exe"

Move-Item -Path ".\yuzu_ssbu_launcher.exe" -Destination ".\build\yuzu_ssbu_launcher"

Compress-Archive -Path ".\build\yuzu_ssbu_launcher" -DestinationPath ".\build\yuzu_ssbu_launcher.zip" -Force
