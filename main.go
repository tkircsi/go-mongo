package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongoadmin:secret@localhost:27017/quickstart?tls=false&authSource=admin"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	dbs, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Databases:", dbs)

	defer client.Disconnect(ctx)

	batchLoad(client.Database("podcastsdb"))

	// quickstartDatabase := client.Database("quickstart")
	// podcastsCollection := quickstartDatabase.Collection("podcasts")
	// episodesCollection := quickstartDatabase.Collection("episodes")
	// _ = episodesCollection

	// podcastResult, err := podcastsCollection.InsertOne(ctx, bson.D{
	// 	{Key: "title", Value: "The Polyglot Developer Podcast"},
	// 	{Key: "author", Value: "Nic Raboy"},
	// 	{Key: "tags", Value: bson.A{"development", "programming", "coding"}},
	// })

	// episodeResult, err := episodesCollection.InsertMany(ctx, []interface{}{
	// 	bson.D{
	// 		{Key: "podcast", Value: podcastResult.InsertedID},
	// 		{Key: "title", Value: "GraphQL for API Development"},
	// 		{Key: "description", Value: "Learn about GraphQL from the co-creator of GraphQL, Lee Byron."},
	// 		{Key: "duration", Value: 25},
	// 	},
	// 	bson.D{
	// 		{Key: "podcast", Value: podcastResult.InsertedID},
	// 		{Key: "title", Value: "Progressive Web Application Development"},
	// 		{Key: "description", Value: "Learn about PWA development with Tara Manicsic."},
	// 		{Key: "duration", Value: 32},
	// 	},
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Inserted %v documents into episode collection.\n", len(episodeResult.InsertedIDs))

}

func batchLoad(db *mongo.Database) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	podcastsCollection := db.Collection("podcasts")
	csvfile, err := os.Open("podcasts.csv")
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(csvfile)
	_, err = r.Read() // skip header row

	const batchSize int = 1000

Loop:
	for {
		var podcasts []interface{}
		for i := 0; i < batchSize; i++ {
			row, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			podcasts = append(podcasts, bson.D{
				{Key: "title", Value: row[1]},
				{Key: "author", Value: row[7]},
			})
			// fmt.Printf("Title: %s Author: %s\n", row[1], row[7])
		}
		podcastResult, err := podcastsCollection.InsertMany(ctx, podcasts)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Inserted %v documents into podcasts collection.\n", len(podcastResult.InsertedIDs))
		if len(podcastResult.InsertedIDs) < batchSize {
			break Loop
		}
	}

}
