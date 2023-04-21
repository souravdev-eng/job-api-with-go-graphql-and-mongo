package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/souravdev-eng/gql-mongo/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Fatal(err)
	}

	return &DB{
		client: client,
	}
}

func HandelError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DB) GetJob(id string) *model.JobListing {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	var jobListing model.JobListing

	err := jobCollection.FindOne(ctx, filter).Decode(&jobListing)

	// Custom error handler
	HandelError(err)

	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	var jobListings []*model.JobListing
	cursor, err := jobCollection.Find(ctx, bson.D{})

	HandelError(err)

	if err = cursor.All(context.TODO(), &jobListings); err != nil {
		panic(err)
	}

	return jobListings
}

func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	inserted, err := jobCollection.InsertOne(ctx, bson.M{
		"title":       jobInfo.Title,
		"description": jobInfo.Description,
		"url":         jobInfo.URL,
		"company":     jobInfo.Company,
	})

	HandelError(err)

	//! What is this line of code mean
	insertedId := inserted.InsertedID.(primitive.ObjectID).Hex()
	returnJobListing := model.JobListing{
		ID:          insertedId,
		Title:       jobInfo.Title,
		Company:     jobInfo.Company,
		Description: jobInfo.Description,
		URL:         jobInfo.URL,
	}

	return &returnJobListing
}

func (db *DB) UpdateJobListing(id string, jobInfo model.UpdateJobListingInput) *model.JobListing {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	updateJobInfo := bson.M{}

	if jobInfo.Title != nil {
		updateJobInfo["title"] = jobInfo.Title
	}

	if jobInfo.Description != nil {
		updateJobInfo["description"] = jobInfo.Description
	}

	if jobInfo.URL != nil {
		updateJobInfo["url"] = jobInfo.URL
	}

	_id, _ := primitive.ObjectIDFromHex(id)

	filter := bson.M{"_id": _id}
	update := bson.M{"$set": updateJobInfo}

	// ! What is `options.FindOneAndUpdate()`
	result := jobCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))

	var jobListing model.JobListing
	if err := result.Decode(&jobListing); err != nil {
		log.Fatal(err)
	}

	return &jobListing
}

func (db *DB) DeleteJobListing(id string) *model.DeleteJobResponse {
	jobCollection := db.client.Database("graphql-job-board").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	_, err := jobCollection.DeleteOne(ctx, filter)
	HandelError(err)

	return &model.DeleteJobResponse{DeletedJobID: id}
}
