package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)


var (
	TARGET_FPS       		int
	SSBU_TITLE_ID     		= "01006A800016E000"
	LOCAL_APP_DATA 			= os.Getenv("LOCALAPPDATA")
	ROAMING_APP_DATA 		= os.Getenv("APPDATA")
	YUZU_GLOBAL_CONFIG   	= filepath.Join(ROAMING_APP_DATA, "yuzu", "config", "qt-config.ini")
	SSBU_MOD_LOC  			= filepath.Join(ROAMING_APP_DATA, "yuzu", "sdmc", "yuzu", "load", SSBU_TITLE_ID)
	SSBU_CONFIG  			= filepath.Join(ROAMING_APP_DATA, "yuzu", "config", "custom", fmt.Sprintf("%s.ini", SSBU_TITLE_ID))
	DEFAULT_LAUNCH_DIR		= filepath.Join(LOCAL_APP_DATA, "yuzu_ssbu")
)

func main() {
	logFile, _ := os.OpenFile("yuzu_ssbu_launcher.log", os.O_WRONLY|os.O_CREATE, 0666)
	defer logFile.Close()
	os.Stdout = logFile
    os.Stderr = logFile

	ini.PrettySection = true
	ini.PrettyFormat = false
	ini.PrettyEqual = false
	
	if len(os.Args) < 2 {
		errorExit("Please provide an integer representing the FPS you want to run the game at (ex: 120)", nil, 1)
	}
	TARGET_FPS = parseInt(os.Args[1])
	fmt.Println("Target FPS:", TARGET_FPS);

	fmt.Println("Searching for SSBU Rom...")
	ssbuGamePath := findSSBURom()
	if ssbuGamePath == "" {
		errorExit("Unable to find SSBU Rom", nil, 1)
	}
	fmt.Println("Found SSBU ROM:", ssbuGamePath)

	fmt.Println("Updating Game Speed...")
	updateGameSpeed()
	fmt.Println("Updating FPS Mod...")
	updateFPSMod()
	fmt.Println("Starting Yuzu...")
	startYuzu(ssbuGamePath)
}

func findSSBURom() string {
	gameDirectories := []string{}
	
	globalConfig, err := ini.Load(YUZU_GLOBAL_CONFIG)
	if err != nil {
		errorExit("Error opening global config file", err, 1)
	}


	globalUIConfig := globalConfig.Section("UI")
	for _, key := range globalUIConfig.Keys() {
		if strings.HasPrefix(strings.ToLower(key.Name()), "paths\\gamedirs") {
			path := key.String()
			if fileInfo, err := os.Stat(path); err == nil && fileInfo.IsDir() {
				gameDirectories = append(gameDirectories, path)
			}
		}
	}

	if len(gameDirectories) == 0 {
		errorExit("Unable to find any yuzu game directories", err, 1)
	}

	for _, gameDirectory := range gameDirectories {
		dir, err := os.Open(gameDirectory)
		if err != nil {
			errorExit("Error opening game directory", err, 1)
		}
		defer dir.Close()

		fileInfos, err := dir.Readdir(-1)
		if err != nil {
			errorExit("Error reading game directory contents", err, 1)
		}
		
		for _, fileInfo := range fileInfos {
			if fileInfo.Mode().IsRegular() {
				fileName := fileInfo.Name()
				fileExt := filepath.Ext(fileName)
				containsSmashString := strings.Contains(fileName, SSBU_TITLE_ID) || 
									strings.Contains(fileName, "Super Smash Bros") ||
									strings.Contains(fileName, "SSBU")
				isROMFile := fileExt == ".xci" || fileExt == ".nsp"
				isBigFile := fileInfo.Size() >= 13000000000
				if containsSmashString && isROMFile && isBigFile {
					ssbuGamePath := filepath.Join(gameDirectory, fileName)
					return ssbuGamePath
				}
			}
		}
	}

	return ""
}

func updateGameSpeed() {
	gameSpeed := int((float64(TARGET_FPS) / 60.0) * 100)
	gameSpeedString := strconv.Itoa(gameSpeed)

	if _, err := os.Stat(SSBU_CONFIG); os.IsNotExist(err) {
		errorExit("SSBU config file doesn't exist", err, 1)
	}

	ssbuGameConfig, err := ini.Load(SSBU_CONFIG)
	if err != nil {
		errorExit("Error loading SSBU game config", err, 1)
	}

	ssbuGameSystemConfig := ssbuGameConfig.Section("Core")
	ssbuGameSystemConfig.Key("speed_limit\\use_global").SetValue("false")
	ssbuGameSystemConfig.Key("speed_limit\\default").SetValue("false")
	ssbuGameSystemConfig.Key("speed_limit").SetValue(gameSpeedString)

	err = ssbuGameConfig.SaveTo(SSBU_CONFIG)
	if err != nil {
		errorExit("Error saving SSBU game config", err, 1)
	}
}

func updateFPSMod() {
	internalGameFPS := 3600 / TARGET_FPS
	internalGameFPSHexFormatted := fmt.Sprintf("%08X", internalGameFPS)
		
	FPSModPath := filepath.Join(SSBU_MOD_LOC, "Custom FPS", "cheats")
	if _, err := os.Stat(FPSModPath); err != nil {
		err := os.MkdirAll(FPSModPath, 0777)
		if err != nil {
			errorExit("Error creating fps mod directory tree", err, 1)
		}
	}

	fpsCheatFile, err := os.OpenFile(filepath.Join(FPSModPath, "B9B166DF1DB90BAF.txt"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		errorExit("Error writing fps mod file", err, 1)
	}
	fpsCheatFile.WriteString(fmt.Sprintf("[%d FPS]", TARGET_FPS))
	fpsCheatFile.WriteString("\n")
	fpsCheatFile.WriteString("04000000 0523B004 ")
	fpsCheatFile.WriteString(internalGameFPSHexFormatted)
	fpsCheatFile.WriteString("\n")
	fpsCheatFile.Close()
}

func startYuzu(gamePath string) {
	if _, err := os.Stat("maintenancetool.exe"); err != nil {
		os.Chdir(DEFAULT_LAUNCH_DIR)
	}
	launcherPath, _ := filepath.Abs("maintenancetool.exe")
	yuzuPath, _ := filepath.Abs(filepath.Join("yuzu-windows-msvc", "yuzu.exe"))
	cmd := exec.Command(launcherPath, "--launcher", yuzuPath, "--launcher_arg", gamePath)
	err := cmd.Start()
	if err != nil {
		errorExit("Error starting yuzu:", err, 1)
	}
}

func parseInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		errorExit("Error parsing int", err, 1)
	}
	return val
}

func errorExit(message string, err error, exitCode int) {
	fmt.Println(message)
	fmt.Println(err)
	os.Exit(exitCode)
}