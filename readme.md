# Goboys 

Homework for application process at Cast AI

## Requirements

* We have a set of cowboys.
* Each cowboy has a unique name, health points and damage points.
* Cowboys list must be stored in persistent storage (File, Database etc).
* Each cowboy should run in it's own isolated process, workload or replica.
* All communication between cowboys should happen via your preferred networking solution (TCP, gRPC, HTTP, MQ etc). 
* Cowboys encounter starts at the same time in parallel. Each cowboys selects random target and shoots.
* Subtract shooter damage points from target health points.
  * If target cowboy health points are 0 or lower, then target is dead.
  * Cowboys don't shoot themselves and don't shoot dead cowboys.
* After the shot shooter sleeps for 1 second.
* Last standing cowboy is the winner.
* Outcome of the task is to print log of every action and winner should log that he won.
* Kubernetes, Docker-compose or any other container orchestration solution is preferred, but optional for final deployment manifests. 
* Provide startup and usage instructions in Readme.MD

## Thinking Process

### First approach
* We need some kind of orchestrator that
    * takes the Input and starts the battle
    * stores the list of cowboys (first in a file for simplicity)
    * keeps track of the health state of the cowboys
    * notices when the battle is over
* We need instances of a cowboy that 
    * can shoot, which includes randomly choosing an enemy
    * can be shot, meaning substracting health and informing the orchestrator
    * sleeps inbetween shots
* Review before coding
    * How are cowboys represented? Process that run in parallel? That would be the simplest approach
    * How can they find each other?
        * Cowboys need to be single applications as they need to have APIs 
        * Orchestrator knows the health state, let's ask him
    * Orchestrator is actually the wrong name, it is more a referee
* While Coding 
    * While http and file store might be the simplest approach, I was thinking of 
* Concept
    * We have two components, referee and cowboy, where at we have 1 referee and several cowboys
    * Both reside in different folders 
    * The referee takes a list of cowboys and spins up several cowboys by running child processes start the cowboy binary
    * Each cowboy runs on a different port that gets stored together with other properties of the cowboy in a file
    * Changes on the cowboy code needs to be followed by `go build` in the cowboy directory
* Solution is on branch `first`

### Run the application
* Optionally set `GIN_MODE` to `test` in order to not have that many logs of the webserver framework
* Optionally set `PORT` to change the port, the application runs at
* Optionally set `LOG_LEVEL` to `prod` which only shows info and above 
* `go run main.go`
* use endppoint to save cowboys
* use endpoint to start battle

### Referee API 
* POST <host>:<port>/cowboy - saves the cowboys from body json 
* GET <host>:<port>/cowboy - gets list of cowboys with their properties
* DELETE <host>:<port>/cowboy/<name> - deletes a cowboy
* PUT <host>:<port>/cowboy/<name> - updates a cowboy
* POST <host>:<port>/startShooting - starts the battle. Takes the list of cowboys as request body. 