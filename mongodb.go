package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/core/command"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var mongoClient *mongo.Client
var mongoDb *mongo.Database
var mongoCollectionsList = make(map[string]bool)
var mongoCollectionHb *mongo.Collection
var saveRecordsErrors = 0
var saveRecordsErrorsLimit = 10

// saveRecords
// VERY IMPORTANT THE WAIT GROUP... USE THE POINTER OTHERWISE WE CREATE A COPY OF IT AND IT WILL BE A LOOONG WAIT
func saveRecordsGrtukri(collectionName string, data []interface{}, page int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Request to save page %d ...\n", page)
	_, err := mongoDb.Collection(collectionName).InsertMany(context.Background(), data)
	if err != nil {
		fmt.Printf("💥💥💥💥💥💥 Error while saving to mongo: %s\n", err)
		saveRecordsErrors++
		return
	}
	fmt.Printf("Page %d saved!\n", page)

}

// initialiseMongo
func initialiseMongo() bool {
	fmt.Printf("Initialising Mongo... 🍃 \n")

	var err error
	var uri, db, collhb string = os.Getenv("MONGO_URI"),
		os.Getenv("MONGO_DATABASE"),
		os.Getenv("MONGO_COLLECTION_HB")

	if uri == "" || db == "" || collhb == "" {
		fmt.Printf("Db parameters missing\n")
		return false
	}

	// CREATE CLIENT
	mongoClient, err = mongo.Connect(context.Background(), uri, nil)
	if err != nil {
		fmt.Printf("Mongo db client creation failed\n")
		return false
	}
	fmt.Printf("Mongo db client created\n")

	/*

		// THIS BIT OF CODE WAS INITIALLY NEEDED.. NOW IT IS NOT ANYMORE
		// NOT SURE WHY...
		// LEFT FOR FUTURE REFERENCE
		// TODO REMOVE BEFORE PRODUCTION

		//CONNECT TO CLIENT
		err = mongoClient.Connect(context.Background())
		if err != nil {
			fmt.Printf("Db client connection failed %s\n", err)
			return false
		}
		fmt.Printf("Mongo client connection ok\n")
	*/

	// SELECT DATABASE
	dbExists, err := databaseExists(db)
	if err != nil {
		fmt.Printf("Error while retrieving databases list 💥 : %s\n", err)
		return false
	}
	if !dbExists {
		fmt.Printf("Mongo db %s does not exists\n", db)
		return false
	}
	fmt.Printf("Mongo db %s found\n", db)

	mongoDb = mongoClient.Database(db)

	// RETRIEVE COLLECTIONS LIST
	err = retrieveCollectionsList()
	if err != nil {
		fmt.Printf("Mongo db collections list not retrieved 💥: %s\n", err)
		return false
	}

	fmt.Printf("Mongo db collections list retrieved\n")

	// SELECT COLLECTION FOR HB
	if !collectionExists(collhb) {
		fmt.Printf("Mongo db collection %s does not exist\n", collhb)
		return false
	}
	fmt.Printf("Mongo db collection %s exist, this is good...\n", collhb)
	mongoCollectionHb = mongoDb.Collection(collhb)

	return true
}

func retrieveCollectionsList() (error) {
	var cur command.Cursor
	var err error
	cnt := context.Background()

	cur, err = mongoDb.ListCollections(context.Background(), nil)
	if err != nil {
		fmt.Printf("Mongo db collections list not retrieved\n")
		return err
	}

	for cur.Next(cnt) {
		elem := bson.NewDocument()
		if err := cur.Decode(elem); err != nil {
			fmt.Printf("Unable to decode element while reading collections list\n")
			return err
		}
		name := elem.Lookup("name").StringValue()
		fmt.Printf("collection found %s\n", name)
		mongoCollectionsList[name] = true
	}

	if err := cur.Err(); err != nil {
		fmt.Printf("Cursor error while reading collections list\n")
		return err
	}

	return nil
}

func databaseExists(databaseName string) (bool, error) {
	var databasesList []string
	var err error

	databasesList, err = mongoClient.ListDatabaseNames(context.Background(), nil)
	if err != nil {
		return false, err
	}
	for _, v := range databasesList {
		if v == databaseName {
			return true, nil
		}
	}
	return false, nil
}

func collectionExists(collectionName string) bool {
	_, exists := mongoCollectionsList[collectionName]
	return exists
}

func closeMongo() {
	fmt.Printf("Disconnecting Mongo client... 🍂 \n")
	mongoCollectionHb = nil
	mongoClient.Disconnect(context.Background())
	mongoClient = nil
	mongoDb = nil
	saveRecordsErrors = 0
}
