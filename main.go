package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path/filepath"
)

func main() {

	myApp := app.New()
	myWindow := myApp.NewWindow("Base64 Transformer")

	var selectedFiles []string

	fileList := widget.NewList(func() int {
		return len(selectedFiles)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("Item")
	}, func(id widget.ListItemID, object fyne.CanvasObject) {
		object.(*widget.Label).SetText(selectedFiles[id])
	})

	fileList.OnSelected = func(id widget.ListItemID) {
		selectedFilePath := selectedFiles[id]
		dialog.ShowInformation("Selected File", selectedFilePath, myWindow)

	}

	var saveLocation string

	saveJsonFileButton := widget.NewButtonWithIcon("Save JSON File Path", theme.FileIcon(), func() {

		//jsonDialog := dialog.NewFolderOpen(func(closer fyne.URIWriteCloser, err error) {
		//	if err == nil && closer != nil {
		//		saveLocation = closer.URI().Path()
		//		closer.Close()
		//		dialog.ShowInformation("Successful", "JSON file saved.", myWindow)
		//	}
		//
		//}, myWindow)
		//jsonDialog.Show()

		jsonDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				saveLocation = uri.Path()
				dialog.ShowInformation("Successful", "JSON file saved.", myWindow)
			}
		}, myWindow)

		jsonDialog.Show()
	})
	convertButton := widget.NewButton("Transform and Save", func() {
		var data []map[string]string

		for _, fileName := range selectedFiles {
			content, err := os.ReadFile(fileName)
			if err != nil {
				log.Println("File read error:", err)
				return
			}

			base64Content := base64.StdEncoding.EncodeToString(content)
			fileName = filepath.Clean(fileName)
			_, fileName := filepath.Split(fileName)
			extension := filepath.Ext(fileName)
			fileNameWithoutExt := fileName[:len(fileName)-len(extension)]
			fmt.Println(fileNameWithoutExt)
			data = append(data, map[string]string{
				"FileName":      fileNameWithoutExt,
				"Base64Content": base64Content})

		}

		jsonData, err := json.MarshalIndent(data, "", "")
		if err != nil {
			log.Println("JSON creation error:", err)
			return
		}

		jsonFilePath := filepath.Join(saveLocation, "content.json")
		err = os.WriteFile(jsonFilePath, jsonData, 0666)
		if err != nil {
			log.Println("Error writing JSON file:", err)
			return
		}

		dialog.ShowInformation("Process Completed", "Files converted to base64 and saved to content.json",
			myWindow)

	})

	selectButton := widget.NewButton("Select File", func() {

		//fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		//	if err == nil && reader != nil {
		//		filePath := reader.URI().Path()
		//
		//		selectedFiles = append(selectedFiles, filePath)
		//		fileList.Refresh()
		//	}
		//}, myWindow)
		//
		//fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".fbx", ".Fbx", ".glb", ".png", ".jpg"}))
		//fileDialog.Show()

		filePaths, _, err := FileMultiSelect("Select File", ".fbx .Fbx .glb .png .jpg")
		if err != nil {
			fmt.Println("File Selection Error: ", err)
			return
		}

		for _, path := range filePaths {
			selectedFiles = append(selectedFiles, path)
		}

		fileList.Refresh()
	})

	content := container.NewVBox(
		layout.NewSpacer(),
		widget.NewLabel("Select Files and Transform to Base64"),
		saveJsonFileButton,
		layout.NewSpacer(),
		fileList,
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			selectButton,
			convertButton,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 300))
	myWindow.ShowAndRun()
}
