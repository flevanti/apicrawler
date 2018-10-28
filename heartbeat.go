package main

import (
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"time"
)

var heartbeatSeconds int64 = 30
var heartbeatDatabaseID string
var heartbeatDatabaseObjectId objectid.ObjectID
var heartbeatDatabaseFilter *bson.Document

// heartbeatKeepAlive
func heartbeatKeepAlive(processKey string) bool {
	hbAlreadyAlive, err := heartbeatCheckIfAlreadyAlive(processKey)
	if err != nil {
		fmt.Printf("Something wrong while checking if heartbeat is already alive %s\n", err)
		return false
	}
	if hbAlreadyAlive {
		fmt.Printf("Heartbeat ID %s is already alive! \n", processKey)
		return false
	}

	if !heartbeatStart(processKey) {
		fmt.Printf("Unable to start the heartbeat ID %s! \n", processKey)
		return false
	}
	fmt.Printf("heartbeat ID started!!! %s (db id %s)\n", processKey, heartbeatDatabaseID)

	go func() {
		for {
			time.Sleep(time.Duration(heartbeatSeconds) * time.Second)
			fmt.Printf("Alive and kicking.... ❤️ \n")
			heartbeatUpdate(false)

			if err != nil {
				fmt.Printf("Error while updating heartbeat!❣️❣️❣️❣️❣️❣️ 💥💥💥💥%s \n", err)
			}
		}
	}()

	return true

}

func heartbeatUpdate(lastHeartbeat bool) error {
	record := bson.NewDocument()
	//https://docs.mongodb.com/manual/reference/operator/update/
	if !lastHeartbeat {
		record = record.Append(
			bson.EC.SubDocument("$set", bson.NewDocument(
				bson.EC.Int64("beat_last", getTimestamp()),
			)))
	} else {
		record = record.Append(bson.EC.SubDocument("$set", bson.NewDocument(
			bson.EC.Int64("beat_ended", getTimestamp()),
		)))
	}

	_, err := mongoCollectionHb.UpdateOne(nil, heartbeatDatabaseFilter, record)

	return err
}

func heartbeatStart(processKey string) bool {
	record := bson.NewDocument(bson.EC.String("process_name", processKey),
		bson.EC.Int64("beat_started", getTimestamp()),
		bson.EC.Int64("beat_last", getTimestamp()),
		bson.EC.Int64("beat_ended", 0))
	result, err := mongoCollectionHb.InsertOne(nil, record)
	if err != nil {
		fmt.Printf("Error while starting heartbeat %s\n", err)
		return false
	}
	//store some information...
	heartbeatDatabaseObjectId = result.InsertedID.(objectid.ObjectID)
	heartbeatDatabaseID = heartbeatDatabaseObjectId.Hex()
	heartbeatDatabaseFilter = bson.NewDocument(bson.EC.ObjectID("_id", heartbeatDatabaseObjectId))

	fmt.Printf("Hearbeat started... DB ID %s      ❤️\n", heartbeatDatabaseID)

	return true
}

func heartbeatEnd() {
	heartbeatUpdate(true)
	fmt.Printf("Hearbeat ended...  💔 \n")

}

func heartbeatCheckIfAlreadyAlive(processKey string) (bool, error) {
	//look for the heartbeat of the process name happened recently (heartbeat seconds * 2)
	timestampLimit := getTimestamp() - (heartbeatSeconds * 2)
	filter := bson.NewDocument(
		bson.EC.String("process_name", processKey),
		bson.EC.SubDocumentFromElements("beat_last", bson.EC.Int64("$gte", timestampLimit)))

	cur, err := mongoCollectionHb.Find(nil, filter)
	if err != nil {
		//SOME ERROR WHILE QUERY DB
		fmt.Printf("SOME ERROR WHILE QUERY DB %s\n", err)
		return true, err
	}

	if !cur.Next(nil) {
		if err = cur.Err(); err != nil {
			//SOME ERROR WHILE MOVING TO NEXT RECORD
			fmt.Printf("ERROR WHILE MOVING TO NEXT RECORD %s\n", cur.Err())
			return true, cur.Err()
		}

		//HOORAY!
		fmt.Printf("Looks like no other processes are working on this source\n")
		return false, nil
	}

	recordValue := bson.NewDocument()
	if err := cur.Decode(recordValue); err != nil {
		fmt.Printf("ERROR WHILE DECODING RECORD %s\n!", err)
		return true, err
	}

	elementValue, err := recordValue.LookupErr("beat_started")
	var beatStarted int64
	var beatLast int64
	if err != nil {
		beatStarted = 0
	} else {
		beatStarted = elementValue.Int64() //this does not perform casting so the type needs to be the same as the one saved!
	}
	beatLast = recordValue.Lookup("beat_last").Int64() //this does not perform casting so the type needs to be the same as the one saved!

	fmt.Printf("Another process is currently processing the same source! 👷 \n")
	fmt.Printf("The other process started %s and its last beat was %s\n",
		time.Unix(beatStarted, 0),
		time.Unix(beatLast, 0))

	return true, nil
}
