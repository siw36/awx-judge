package mongoConnector

import (
	"context"
	"errors"
	"strconv"
	"time"

	model "../model"

	log "github.com/Sirupsen/logrus"
	guuid "github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Config model.Config
	Client *mongo.Client
)

// Database connection
func DBConnect(connectionString string, database string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(connectionString)
	log.Info("Trying to connect to MongoDB")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Database connection failed", err)
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Connected to MongoDB")
	return client
}

func DBDisconnect(client *mongo.Client) {
	err := (client).Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("Connection to MongoDB closed")
	}
}

// Create a job template
func DBCreateTemplate(template model.Template) error {
	// Switch collection
	collection := Client.Database(Config.Mongo.Database).Collection("templates")
	// Add metadata
	now := time.Now()
	template.UpdatedAt = now
	// Create or update document
	filter := bson.D{primitive.E{Key: "id", Value: template.ID}}
	var opts = options.Replace().SetUpsert(true)
	*opts.Upsert = true
	log.Info("Writing job template to DB")
	_, err := collection.ReplaceOne(context.TODO(), filter, template, opts)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// Get a job Template
func DBGetTemplate(id int) (model.Template, error) {
	filter := bson.D{primitive.E{"id", id}}
	opts := options.FindOne().SetSort(bson.D{primitive.E{Key: "name", Value: 1}})
	var result model.Template
	collection := Client.Database(Config.Mongo.Database).Collection("templates")
	err := collection.FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			return result, errors.New("Could not find requested job template in DB with ID " + strconv.Itoa(id))
		}
		return result, err
	}
	return result, nil
}

// Create a job template
func DBRemoveTemplate(templateID int) error {
	// Switch collection
	collection := Client.Database(Config.Mongo.Database).Collection("templates")
	// Remove the job template from DB
	filter := bson.D{{"id", templateID}}
	var opts = options.Delete()
	log.Info("Deleting job template with ID ", templateID)
	_, err := collection.DeleteOne(context.TODO(), filter, opts)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func DBGetTemplateAll() ([]model.Template, error) {
	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{primitive.E{Key: "name", Value: 1}})
	var results []model.Template
	collection := Client.Database(Config.Mongo.Database).Collection("templates")
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		log.Error(err)
		return results, err
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Error(err)
		return results, err
	}
	return results, nil
}

// Create a new cart
func DBCreateCart(userID string) error {
	// Switch collection
	collection := Client.Database(Config.Mongo.Database).Collection("carts")
	// Add metadata
	now := time.Now()
	// Create or update document
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	var opts = options.Update().SetUpsert(true)
	update := bson.D{{"$setOnInsert",
		bson.D{
			{"user_id", userID},
			{"updated_at", now},
			{"created_at", now},
		},
	}}
	log.Info("Creating new cart for user ", userID, " if none is present")
	_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// Delete the users cart
func DBDeleteCart(userID string) error {
	// Switch collection
	collection := Client.Database(Config.Mongo.Database).Collection("carts")
	log.Info("Deleting cart for user ", userID)
	opts := options.Delete()
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	_, err := collection.DeleteOne(context.TODO(), filter, opts)
	if err != nil {
		log.Error(err)
		return err
	}
	DBCreateCart(userID)
	return nil
}

// Get a cart
func DBGetCart(userID string) (model.Cart, error) {
	log.Info("Getting user cart")
	var result model.Cart
	// Switch collection
	collection := Client.Database(Config.Mongo.Database).Collection("carts")
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	opts := options.FindOne()
	err := collection.FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return result, errors.New("Could not find requested cart")
		}
		log.Error(err)
	}
	return result, nil
}

// Update a cart - add request
func DBUpdateCartAdd(userID string, request model.Request) error {
	// Get the current cart
	cart, err := DBGetCart(userID)
	if err != nil {
		return err
	}
	// Append the request to the cart
	request.UserID = userID
	request.ID = guuid.New()
	request.State = "draft"
	cart.Requests = append(cart.Requests, request)
	// Write the updated cart to DB
	collection := Client.Database(Config.Mongo.Database).Collection("carts")
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{"$set", bson.D{{"requests", cart.Requests}}}}
	var opts = options.Update()
	log.Info("Updating cart for user ", userID)
	_, err = collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// Update a cart - remove request
func DBUpdateCartRemove(userID string, requestID guuid.UUID) error {
	// Get the current cart
	cart, err := DBGetCart(userID)
	if err != nil {
		return err
	}
	// Create empty request for overwriting
	var dummyRequest model.Request
	// Remove the request from the cart
	for i, request := range cart.Requests {
		if request.ID == requestID {
			copy(cart.Requests[i:], cart.Requests[i+1:])
			cart.Requests[len(cart.Requests)-1] = dummyRequest
			cart.Requests = cart.Requests[:len(cart.Requests)-1]
		}
	}
	// Write the updated cart to DB
	collection := Client.Database(Config.Mongo.Database).Collection("carts")
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	update := bson.D{{"$set", bson.D{{"requests", cart.Requests}}}}
	var opts = options.Update()
	log.Info("Updating cart for user ", userID)
	_, err = collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func DBUpdateCartEdit(userID string, request model.Request) error {
	collection := Client.Database(Config.Mongo.Database).Collection("carts")
	opts := options.Update().SetUpsert(true)
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}, primitive.E{Key: "requests.id", Value: request.ID}}
	update := bson.D{{"$set", bson.D{{"requests.$", request}}}}

	result, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Error(err)
		return err
	}

	if result.MatchedCount != 0 {
		log.Info("Updated cart item for user ", userID)
		return nil
	}
	if result.UpsertedCount != 0 {
		log.Info("Added new item to cart for user ", userID)
		return nil
	}
	return nil
}

// Create a new request
func DBCreateRequest(userID string, request model.Request) error {
	collection := Client.Database(Config.Mongo.Database).Collection("requests")
	// Add meta data
	request.State = "pending"
	now := time.Now()
	request.UpdatedAt = now
	request.CreatedAt = now
	// Write to DB
	_, err := collection.InsertOne(context.TODO(), request)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info("Created a new request")
	// Send slack notification
	return nil
}

func DBCartToRequest(userID string) error {
	// Get the current cart
	cart, err := DBGetCart(userID)
	if err != nil {
		return err
	}
	for _, request := range cart.Requests {
		err = DBCreateRequest(userID, request)
		if err != nil {
			return err
		}
	}
	// Delete the users cart
	DBDeleteCart(userID)
	return nil
}

// Get requests
func DBGetRequests(userID string) ([]model.Request, error) {
	log.Info("Getting user requests")
	var results []model.Request
	// Switch collection
	collection := Client.Database(Config.Mongo.Database).Collection("requests")
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	opts := options.Find().SetSort(bson.D{{Key: "updated_at", Value: 1}})
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if cursor == nil {
		log.Error("Query did not return a cursor: ", err)
		cursor.Close(context.TODO())
		return nil, err
	}
	if err != nil {
		log.Info("Did not find any requests: ", err)
		cursor.Close(context.TODO())
		return nil, err
	}
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Error(err)
		cursor.Close(context.TODO())
		return nil, err
	}
	cursor.Close(context.TODO())
	return results, nil
}

// Get a request
func DBGetRequest(userID string, requestID guuid.UUID) (model.Request, error) {
	log.Info("Getting user request")
	var result model.Request
	// Switch collection
	collection := Client.Database(Config.Mongo.Database).Collection("requests")
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}, primitive.E{Key: "id", Value: requestID}}
	opts := options.FindOne()
	err := collection.FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in the collection
		if err == mongo.ErrNoDocuments {
			log.Error(err)
			return result, errors.New("Could not find request")
		}
		log.Error(err)
	}
	return result, nil
}

func DBUpdateRequest(userID string, request model.Request) error {
	collection := Client.Database(Config.Mongo.Database).Collection("requests")
	opts := options.Update().SetUpsert(true)
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}, primitive.E{Key: "id", Value: request.ID}}
	update := bson.D{{"$set", bson.D{{"requests.$", request}}}}

	result, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Error(err)
		return err
	}

	if result.MatchedCount != 0 {
		log.Info("Updated request for user ", userID)
		return nil
	}
	if result.UpsertedCount != 0 {
		log.Info("Added request to cart for user ", userID)
		return nil
	}
	return nil
}
