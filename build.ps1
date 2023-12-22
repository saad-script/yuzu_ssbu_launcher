# get rsrc from https://github.com/akavel/rsrc
cd ".\build"
Copy-Item -Path "..\icons" -Destination ".\yuzu_ssbu_launcher" -Recurse -Force
..\rsrc_windows_amd64.exe -ico "..\icons\ssbu_red.ico" -o ".\rsrc.syso"
go build "..\" -o ".\yuzu_ssbu_launcher\yuzu_ssbu_launcher.exe"
Compress-Archive -Path ".\yuzu_ssbu_launcher" -DestinationPath ".\yuzu_ssbu_launcher.zip" -Force
