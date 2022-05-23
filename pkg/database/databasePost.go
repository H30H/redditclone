package database

import (
	"context"
	"fmt"
	"redditclone/pkg/post"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabasePost interface {
	Insert(pst post.Post) error
	Find(id uint64, pst *post.Post) error
	GetAll(filter interface{}, opts ...*options.FindOptions) ([]*post.Post, error)
	Replace(pst post.Post) error
	Delete(id uint64) error
}

type DatabasePostMongo struct {
	database *mongo.Collection
}

func InitDatabasePost(path, databaseName, collenctionName string) (*DatabasePostMongo, error) {
	clientOptions := options.Client().ApplyURI(path)
	client, errConnect := mongo.Connect(context.TODO(), clientOptions)
	if errConnect != nil {
		return nil, fmt.Errorf("mongodb: can`t connect: %w", errConnect)
	}
	errPing := client.Ping(context.TODO(), nil)
	if errPing != nil {
		return nil, fmt.Errorf("mongodb: can`t ping: %w", errPing)
	}
	collection := client.Database(databaseName).Collection(collenctionName)
	if collection == nil {
		return nil, fmt.Errorf(`mongodb: no such collection (has "nil" collection)`)
	}
	return &DatabasePostMongo{
		database: collection,
	}, nil
}

func (d *DatabasePostMongo) Close() error {
	return d.database.Database().Client().Disconnect(context.TODO())
}

func (d *DatabasePostMongo) Insert(pst post.Post) (err error) {
	_, err = d.database.InsertOne(context.TODO(), pst)
	return
}

func (d *DatabasePostMongo) Find(id uint64, pst *post.Post) (err error) {
	return d.database.FindOne(context.TODO(), bson.M{"id": id}).Decode(pst)
}

func (d *DatabasePostMongo) GetAll(filter interface{}, opts ...*options.FindOptions) (posts []*post.Post, err error) {
	cur, err := d.database.Find(context.TODO(), filter, opts...)
	if err != nil {
		return nil, err
	}
	err = cur.All(context.TODO(), &posts)
	return
}

func (d *DatabasePostMongo) Replace(pst post.Post) (err error) {
	_, err = d.database.ReplaceOne(context.TODO(), bson.M{"id": pst.ID}, pst)
	return
}

func (d *DatabasePostMongo) Delete(id uint64) (err error) {
	_, err = d.database.DeleteOne(context.TODO(), bson.M{"id": id})
	return
}
