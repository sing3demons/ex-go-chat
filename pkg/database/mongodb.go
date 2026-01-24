package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB wraps the MongoDB client
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Connect establishes a connection to MongoDB
func Connect(ctx context.Context, uri, database string, timeout time.Duration) (*MongoDB, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(database)

	return &MongoDB{
		Client:   client,
		Database: db,
	}, nil
}

// Disconnect closes the MongoDB connection
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}

// CreateIndexes creates all necessary indexes for the collections
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Users collection indexes
	usersCollection := m.Database.Collection("users")
	_, err := usersCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Rooms collection indexes
	roomsCollection := m.Database.Collection("rooms")
	_, err = roomsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "members", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "type", Value: 1},
				{Key: "members", Value: 1},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create rooms indexes: %w", err)
	}

	// Messages collection indexes
	messagesCollection := m.Database.Collection("messages")
	_, err = messagesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "roomId", Value: 1}, {Key: "createdAt", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "senderId", Value: 1}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create messages indexes: %w", err)
	}

	// Message Status collection indexes (for separated status tracking)
	messageStatusCollection := m.Database.Collection("messageStatus")
	_, err = messageStatusCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "messageId", Value: 1},
				{Key: "userId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "messageId", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "updatedAt", Value: -1},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create message status indexes: %w", err)
	}

	// Notifications collection indexes
	notificationsCollection := m.Database.Collection("notifications")
	_, err = notificationsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "read", Value: 1}, {Key: "createdAt", Value: -1}},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create notifications indexes: %w", err)
	}

	return nil
}
