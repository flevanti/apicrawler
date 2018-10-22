package main

import (
	"fmt"
	"time"
)

var heartbeatSeconds int64 = 30
var heartbeatDatabaseID = ""

// heartbeatKeepAlive
func heartbeatKeepAlive(processKey string) bool {

	if heartbeatCheckIfAlreadyAlive() {
		fmt.Printf("Heartbeat ID %s is already alive! \n", processKey)
		return false
	}

	if !heartbeatStart() {
		fmt.Printf("Unable to start the heartbeat ID %s! \n", processKey)
		return false
	}
	fmt.Printf("heartbeat ID started!!! %s (db id %s)\n", processKey, heartbeatDatabaseID)

	go func() {
		for {
			time.Sleep(time.Duration(heartbeatSeconds) * time.Second)
			fmt.Printf("Alive and kicking.... ❤️ \n")
			//TODO ADD HEARTBEAT TO DB
		}
	}()

	return true

}

func heartbeatStart() bool {
	//TODO INITIALISE HEARTBEAT
	heartbeatDatabaseID = "dummyID"
	fmt.Printf("Hearbeat started... ❤️\n")

	return true
}

func heartbeatEnd() {
	//TODO CLOSE HEARTBEAT
	fmt.Printf("Hearbeat ended...  💔 \n")
	heartbeatDatabaseID = ""

}

func heartbeatCheckIfAlreadyAlive() bool {
	// TODO ADD LOGIC
	return false
}
