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

### Second approach
* We stick with the referee and cowboys separation in separate projects/modules from the first approach
* Referee, 
    * Keeps the state of cowboys, but this time we use psql
    * Starts the battle 
* Cowboys 
    * just shoot a message in a message queue, and a random cowboy picks it up. 
    * If a message does not get acknowledged anymore, the sending cowboy knows, that he is the only one left.
    * Health updates including death are sent as message 
### Review before coding
    *  
### Insights While Coding 
    * There is always one message in the queue left with this approach, so we just purge the queue right after starting the application. Similar situation with the last cowboy in the database. 
### Concept
    * We have two components, referee and cowboy, where at we have 1 referee and several cowboys
    * Both reside in different folders 
    * The referee takes a list of cowboys stores them in a psql database, then spins several cowboys by running child processes start the cowboy binary
    * We use a queue for cowboy to cowboy communication, ie shooting
    * We use a queue for cowboy to referee communication to update health information
    * Changes on the cowboy code needs to be followed by `go build` in the cowboy directory
* Solution is on branch `second`

## Run the application
* Change to `referee` directory
* Optionally set `GIN_MODE` to `test` in order to not have that many logs of the webserver framework
* Optionally set `PORT` to change the port, the application runs at
* Optionally set `LOG_LEVEL` to prod in order to have only logs with level info and above. 
* `docker run -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres` to start a database 
* `docker run -d -p 15672:15672 -p 5672:5672 rabbitmq:management` to run a rabbitmq including [management ui](http://localhost:15672)
* Optionally set `DB_HOST` if the host is not localhost. If not set, defaults to localhost
* Optionally set `RABBIT_HOST` if the host is not localhost. If not set, defaults to localhost
* `go run main.go`
* **Alternative to the last 5 steps is to use `docker-compose up` :)**
* use endppoint to save cowboys
* use endpoint to start battle

## Referee API 
* POST <host>:<port>/cowboy - saves the cowboys from body json 
* POST <host>:<port>/startShooting - starts the battle 

## Todo
* Logging for extracted stuff in common (debug)
* Having the cowboy as global variable does not feel right. 
* Don't add cowboy, if there is already a cowboy with that name -> use name as pk?!