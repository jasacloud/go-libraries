package mongoc

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"log"
)

// TestInsert method
func (c *Connections) TestInsert() {
	var ctx = context.Background()
	var doc = bson.M{"id": "1234", "hometown": "Atlanta"}
	var result *mongo.InsertOneResult
	//c.C("test")
	result, err := c.Collection.InsertOne(ctx, doc)
	if err != nil {
		log.Println(err)
		return
	}

	if result.InsertedID != doc["_id"] {
		log.Println(result.InsertedID)
		log.Println(doc["id"])
	}
}

// TestFind method
func (c *Connections) TestFind() {
	cur, err := c.Collection.Find(context.TODO(), bson.D{})

	if err != nil {
		log.Println(err)
		return
	}
	var results []interface{}
	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var elem map[string]interface{}
		err := cur.Decode(&elem)
		if err != nil {
			log.Println(err)
			continue
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		log.Println(err)
		return
	}

	// Close the cursor once finished
	err = cur.Close(context.TODO())
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(results)
}

// EnsureIndexesTest method
func (c *Connections) EnsureIndexesTest() ([]string, error) {
	indexKeys := bson.D{
		{Key: "id", Value: 1},
	}
	indexOptions := options.Index()
	indexOptions.SetUnique(true)
	indexOptions.SetSparse(true)

	index := mongo.IndexModel{}
	index.Keys = indexKeys
	index.Options = indexOptions

	indexKeys2 := bson.D{
		{Key: "name", Value: 1},
		{Key: "type", Value: 1},
		{Key: "parent", Value: 1},
	}
	indexOptions2 := options.Index()
	indexOptions2.SetUnique(true)
	indexOptions2.SetSparse(true)

	index2 := mongo.IndexModel{}
	index2.Keys = indexKeys2
	index2.Options = indexOptions2

	indexes := []mongo.IndexModel{index, index2}

	return c.EnsureIndexes(indexes...)
}
