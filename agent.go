package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/andlabs/ui"
	"github.com/google/uuid"
)

func downloadCommand() {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://192.168.1.1:888/web/scripts/", nil)
	req.Header.Add("Command-ID", uuid.New().String())
	req.Header.Add("File-Name", "example.txt")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Println("Data from server:", string(body))
}

func uploadCommand() {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://192.168.1.1:888/web/scripts/", nil)
	req.Header.Add("Command-ID", uuid.New().String())
	req.Header.Add("File-Name", "example.txt")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Println("Data from server:", string(body))
}

func Command() {
	
}

func main() {
	ui.Main(func() {
		window := ui.NewWindow("C2-EX-machena", 200, 100, false)
		downloadButton := ui.NewButton("Download")
		downloadButton.OnClicked(func(*ui.Button) {
			downloadCommand()
		})
		uploadButton := ui.NewButton("Upload")
		uploadButton.OnClicked(func(*ui.Button) {
			uploadCommand()
		})
		resetButton := ui.NewButton("Reset")
		resetButton.OnClicked(func(*ui.Button) {
			resetCommand()
		})
		vbox := ui.NewVerticalBox()
		vbox.Append(downloadButton, false)
		vbox.Append(uploadButton, false)
		vbox.Append(resetButton, false)
		window.SetChild(vbox)
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
}
