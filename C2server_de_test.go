package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Task définit le modèle pour les tâches à exécuter
type Task struct {
	Type        string `json:"Type"`        // Le type de tâche, ex: COMMAND
	User        string `json:"User"`        // L'utilisateur associé à la tâche
	Name        string `json:"Name"`        // Le nom de la tâche
	Description string `json:"Description"` // La description de la tâche
	Data        string `json:"Data"`        // Les données associées à la tâche
	Timestamp   string `json:"Timestamp"`   // Le timestamp associé à la tâche
	Id          string `json:"Id"`          // L'ID de la tâche
	After       string `json:"After"`       // Moment après lequel la tâche peut être exécutée
}

// Request définit le modèle pour les requêtes envoyées au serveur C2
type Request struct {
	NextRequestTime string `json:"NextRequestTime"` // Le temps pour la prochaine requête
	Tasks           []Task `json:"Tasks"`           // Liste des tâches à exécuter
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var request Request
		err := decoder.Decode(&request)
		if err != nil {
			fmt.Println("Error decoding request:", err)
			return
		}

		// Traitement de chaque tâche
		for _, task := range request.Tasks {
			switch task.Type {
			case "COMMAND":
				// Vérifie que la commande est autorisée
				if task.Data != "ls -la" {
					fmt.Println("Command not authorized:", task.Data)
					continue
				}

				// Envoie la commande à l'agent
				fmt.Println("Executing command on agent:", task.Data)
				// TODO: envoyer la commande à l'agent
			default:
				fmt.Println("Unknown task type:", task.Type)
			}
		}

		// Réponse à envoyer à l'agent
		response := map[string]string{
			"Message": "Tasks received",
		}
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("C2 server started")
	http.ListenAndServe(":8000", nil)
}
