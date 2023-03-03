# C2 Server Agent

Welcome to the C2 Server Agent repository! This agent is responsible for communication between equipment and the command and control server. This repository contains all the necessary source code to configure and run the agent on your equipment.

# Prerequisites

go installed on your PC
```sudo apt install golang```

# Installation


To configure the agent on your equipment, please follow the following steps:

Download the source code of this repository using the git clone command or by downloading a zip file.

```git clone https://github.com/evaris237/Agent.git```
Configure the IP address and port of the C2 server in the configuration file.
C2agent.go

## Compile the C2agent.go file
 ```go build C2agent.go```
Run the agent.go file to start the agent.
```./C2agent```

# Usage
Once the agent is installed and configured, it will communicate with the C2 server to transmit the necessary information. You can monitor and control the equipment using the C2 server control panel.



# Licence

This project is published under the xxxxxxx license.

Please refer to the LICENSE file for more information on terms of use and distribution.
