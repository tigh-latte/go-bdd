package clients

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

type MongoOptions struct {
	URI string
}

// InitMongo initializes the mongo client, checks that it is accessible, and sets the client
func InitMongo(opts *MongoOptions) error {
	if opts == nil {
		return nil
	}

	mOpts := options.Client().ApplyURI(opts.URI)
	cli, err := mongo.Connect(context.TODO(), mOpts)
	if err != nil {
		return err
	}

	if err := cli.Ping(context.TODO(), nil); err != nil {
		return err
	}

	MongoClient = cli
	return nil
}
