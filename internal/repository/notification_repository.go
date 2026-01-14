package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"realtime-chat-system/internal/models"
)

// NotificationRepository handles notification data operations
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	FindByID(ctx context.Context, id string) (*models.Notification, error)
	FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
	FindPendingByUserID(ctx context.Context, userID string) ([]*models.Notification, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	CountUnread(ctx context.Context, userID string) (int64, error)
}

type notificationRepository struct {
	collection *mongo.Collection
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *mongo.Database) NotificationRepository {
	return &notificationRepository{
		collection: db.Collection("notifications"),
	}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	if notification.ID == "" {
		notification.ID = primitive.NewObjectID().Hex()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, notification)
	return err
}

// FindByID finds a notification by ID
func (r *notificationRepository) FindByID(ctx context.Context, id string) (*models.Notification, error) {
	var notification models.Notification
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&notification)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

// FindByUserID finds notifications for a user with pagination
func (r *notificationRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{"userId": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

// FindPendingByUserID finds unread notifications for a user
func (r *notificationRepository) FindPendingByUserID(ctx context.Context, userID string) ([]*models.Notification, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{
		"userId": userID,
		"read":   false,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"read": true}},
	)
	return err
}

// MarkAllAsRead marks all notifications for a user as read
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"userId": userID, "read": false},
		bson.M{"$set": bson.M{"read": true}},
	)
	return err
}

// CountUnread counts unread notifications for a user
func (r *notificationRepository) CountUnread(ctx context.Context, userID string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"userId": userID,
		"read":   false,
	})
}
