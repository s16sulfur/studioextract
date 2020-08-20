package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Version variable for application version name
var Version = "development"

// Build type for application ["full", "gui"]
var Build = "gui"

var isDebug = (Version == "development")

func printHelp(exeName string) {

	if Build == "full" {
		fmt.Println("Extract charater card from Studio scene card.")

		fmt.Println("Supported games:")
		fmt.Println("\tAI Shoujo")
		fmt.Println("\tHoney Select")
		fmt.Println("\tHoney Select 2")
		fmt.Println("\tKoikatsu")
		fmt.Println("\tPlayHome")

	} else {
		fmt.Println("Extract charater card from Studio NEO / Studio NEO V2 scene card.")
	}

	fmt.Println("\nUsage:")
	fmt.Println("\t", exeName, "file [-options]")
	fmt.Println("\t", exeName, "-h | --help")
	fmt.Println("\t", exeName, "-v | --version")

	fmt.Println("\nOptions:")
	fmt.Println("\t-h --help\tShow this screen.")
	fmt.Println("\t-v --version\tShow version.")
	fmt.Println("\t-m --male\tExtract male charater only.")
	fmt.Println("\t-f --female\tExtract female charater only.")

	fmt.Println("")
}

func parseArgs() (file string, flag int) {
	file = ""
	flag = 0

	args := os.Args
	exePath := args[0]
	exeName := strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))

	aLen := len(args)
	if aLen > 1 {
		switch args[1] {
		case "-h", "--help":
			printHelp(exeName)
			os.Exit(0)
		case "-v", "--version":
			fmt.Println(exeName, "version", Version)
			os.Exit(0)
		case "-m", "--male":
			flag = 1
		case "-f", "--female":
			flag = 2
		default:
			file = args[1]
		}

		if aLen > 2 {
			switch args[2] {
			case "-m", "--male":
				flag = 1
			case "-f", "--female":
				flag = 2
			default:
				file = args[2]
			}
		}

	} else {
		printHelp(exeName)
		os.Exit(0)
	}
	return
}

func init() {
	exe, exErr := os.Executable()
	if exErr == nil {
		isDebug = strings.HasSuffix(exe, "__debug_bin")
	}
}

func main() {
	currDir, err := os.Getwd()
	if err != nil {
		if isDebug {
			panic(err)

		} else {
			printError(err)
			os.Exit(0)
			return
		}
	}

	if Build == "gui" {
		runGui(currDir)

	} else {
		var filePath string
		var flag int

		if isDebug {
			filePath = path.Join(currDir, "temp", "ph_665209fc29e5ffb.png")

		} else {
			filePath, flag = parseArgs()
		}

		_, fErr := os.Stat(filePath)
		if os.IsNotExist(fErr) {
			printError(errors.New("File '" + filePath + "' not found."))
			return
		}

		total, write, err := extractScene(currDir, filePath, flag, Build == "full")
		if err != nil {
			printError(err)
			return
		}

		fmt.Println("\033[30;102m SUCCESS \033[0m", "Extract success.")
		if total == 0 {
			fmt.Println("\t", "No charater found in scene card.")
		} else {
			fmt.Println("\t", total, "charater(s) found and", write, "charater(s) extracted.")
		}
	}
}
