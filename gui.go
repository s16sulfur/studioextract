package main

import (
	"fmt"
	"os"
	"path"

	"github.com/tadvi/winc"
)

var guiCurrDir string
var guiExtFlag int

func runGui(currDir string) {
	guiCurrDir = currDir

	mainWindow := winc.NewForm(nil)
	mainWindow.SetSize(400, 300)
	mainWindow.SetText("Studio Extract")

	mainWindow.EnableMaxButton(false)
	mainWindow.EnableMinButton(false)
	mainWindow.EnableSizable(false)
	mainWindow.EnableDragAcceptFiles(true)

	lFont := winc.NewFont("MS Shell Dlg 2", 20, winc.FontBold)
	label := winc.NewLabel(mainWindow)
	label.SetText("Drop scene file here")
	label.SetFont(lFont)
	label.SetPos(50, 100)
	label.SetSize(300, 40)

	allRadio := winc.NewRadioButton(mainWindow)
	allRadio.SetChecked(true)
	allRadio.SetText("All charater")
	allRadio.SetPos(80, 230)
	allRadio.SetSize(70, 20)
	allRadio.OnClick().Bind(wndOnAllClick)

	maleRadio := winc.NewRadioButton(mainWindow)
	maleRadio.SetText("Male Only")
	maleRadio.SetPos(160, 230)
	maleRadio.SetSize(70, 20)
	maleRadio.OnClick().Bind(wndOnMaleClick)

	femaleRadio := winc.NewRadioButton(mainWindow)
	femaleRadio.SetText("Female Only")
	femaleRadio.SetPos(240, 230)
	femaleRadio.SetSize(80, 20)
	femaleRadio.OnClick().Bind(wndOnFemaleClick)

	mainWindow.Center()
	mainWindow.Show()
	mainWindow.OnDropFiles().Bind(wndOnDropFiles)
	mainWindow.OnClose().Bind(wndOnClose)

	winc.RunMainLoop() // Must call to start event loop.
}

func wndOnAllClick(arg *winc.Event) {
	guiExtFlag = 0
}

func wndOnMaleClick(arg *winc.Event) {
	guiExtFlag = 1
}

func wndOnFemaleClick(arg *winc.Event) {
	guiExtFlag = 2
}

func wndOnDropFiles(arg *winc.Event) {
	dropData, ok := arg.Data.(*winc.DropFilesEventData)
	if ok {
		for _, file := range dropData.Files {
			_, fErr := os.Stat(file)
			if !os.IsNotExist(fErr) && path.Ext(file) == ".png" {
				total, write, err := extractScene(guiCurrDir, file, guiExtFlag, true)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println("extractScene >>", total, write)
			}
		}
	}
}

func wndOnClose(arg *winc.Event) {
	winc.Exit()
	os.Exit(0)
}
