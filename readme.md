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

### Review before coding
* How are cowboys represented? Process that run in parallel? That would be the simplest approach
* How can they find each other?
    * Cowboys need to be single applications as they need to have APIs 
    * Orchestrator knows the health state, let's ask him
* Orchestrator is actually the wrong name, it is more a referee
* While Coding 
    * While http and file store might be the simplest approach, I was thinking of 

### Concept
    * We have two components, referee and cowboy, where at we have 1 referee and several cowboys
    * Both reside in different folders 
    * The referee takes a list of cowboys and spins up several cowboys by running child processes start the cowboy binary
    * Each cowboy runs on a different port that gets stored together with other properties of the cowboy in a file
    * Changes on the cowboy code needs to be followed by `go build` in the cowboy directory
* Solution is on branch `first`

###  Critical reflection
* Spinning up the cowboys through cmd.execute using the cowboy binary feels kind of unusual, but works ^^
* That fact made the developing and debugging a little bit hard at some point because
    * Cowboy child processes remained running when something went wrong and had to be cleaned with kill -9
    * When something in the interaction between the referee and the cowboys did not work as expected it was kind of hard to find the problem. -> Logging helps of course. 
* Maybe we can simplify stuff here using for example goroutines, but somehow that feels wrong. It would be a vertical scaling vs a horizontal scaling which feels more natural somehow. 


### Second approach
* We stick with the referee and cowboys separation in separate projects/modules
* Referee, 
    * Keeps the state of cowboys, but this time we use psql
    * Starts the battle 
* Cowboys 
    * just shoot a message in a message queue, and a random cowboy picks it up. 
    * If a message does not get acknowledged anymore, the sending cowboy knows, that he is the only one left.
    * Health updates including death are sent as message 
* Review before coding
    *  
* Concept
    * We have two components, referee and cowboy, where at we have 1 referee and several cowboys
    * Both reside in different folders 
    * The referee takes a list of cowboys stores them in a psql database, then spins several cowboys by running child processes start the cowboy binary
    * Changes on the cowboy code needs to be followed by `go build` in the cowboy directory
* Solution is on branch `second`

### Final Thoughts
* Running cowboys as threads inside the main app does not feel right. 
* Maybe we could abstract it with some kind of strategy pattern. 
    * As thread (routine) 
    * As Binary 
    * As Container
    * As Kubernetes Manifest
* Using Messaging introduces a SPOF
* The first solution does not work if a cowboy dies, we could use pm2 to keep it up an running and then we could instead of rely on the global variable, ask the referee every time to get the current infos about the cowboy. For the second solution the same. We would need to get the health infos every time from referee instead of relying on the memory. 
* When I think of everything running in docker-compose or kubernetes I get the impression that the cowboys should more likely run inside an own container. But how would I do it?  
    * But how can I transform the cowboy input list to running cowboy docker images? 