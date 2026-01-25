package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"realtime-chat-system/pkg/errors"
	"realtime-chat-system/pkg/logAction"
	"realtime-chat-system/pkg/logger"
	"realtime-chat-system/pkg/mlog"
)

// MessageStatus represents delivery and read status for a message
type MessageStatus struct {
	ID          string     `bson:"_id,omitempty"`
	MessageID   string     `bson:"messageId"`
	UserID      string     `bson:"userId"`
	Delivered   bool       `bson:"delivered"`
	DeliveredAt *time.Time `bson:"deliveredAt,omitempty"`
	Read        bool       `bson:"read"`
	ReadAt      *time.Time `bson:"readAt,omitempty"`
	CreatedAt   time.Time  `bson:"createdAt"`
	UpdatedAt   time.Time  `bson:"updatedAt"`
}

// MessageStatusRepository handles message status operations
type MessageStatusRepository interface {
	UpsertStatus(ctx context.Context, messageID, userID string, status *MessageStatus) error
	BulkUpsertStatus(ctx context.Context, statuses []*MessageStatus) error
	FindStatusForMessage(ctx context.Context, messageID string) ([]*MessageStatus, error)
	FindStatusForUser(ctx context.Context, messageIDs []string, userID string) ([]*MessageStatus, error)
	FindStatusByMessageAndUser(ctx context.Context, messageID, userID string) (*MessageStatus, error)
	DeleteStatusesForMessage(ctx context.Context, messageID string) error
}

type messageStatusRepository struct {
	collection *mongo.Collection
}

// NewMessageStatusRepository creates a new message status repository
func NewMessageStatusRepository(db *mongo.Database) MessageStatusRepository {
	return &messageStatusRepository{
		collection: db.Collection("messageStatus"),
	}
}

// UpsertStatus upserts a message status (insert or update)
func (r *messageStatusRepository) UpsertStatus(ctx context.Context, messageID, userID string, status *MessageStatus) error {
	dependency := "mongo"
	log := mlog.L(ctx)
	start := time.Now()

	if status.ID == "" {
		status.ID = primitive.NewObjectID().Hex()
	}
	status.UpdatedAt = time.Now()
	if status.CreatedAt.IsZero() {
		status.CreatedAt = time.Now()
	}

	filter := bson.M{
		"messageId": messageID,
		"userId":    userID,
	}

	update := bson.M{
		"$set": status,
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_UPDATE, "mongo upsert message status"), map[string]any{
		"filter": filter,
		"update": update,
	})
	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_UPDATE, "mongo upsert message status"), map[string]any{
			"error": err.Error(),
		})
		return errors.ErrDatabase(err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_UPDATE, "mongo upsert message status"), map[string]any{
		"result": result,
	})

	return nil
}

// BulkUpsertStatus performs bulk upsert for multiple statuses
func (r *messageStatusRepository) BulkUpsertStatus(ctx context.Context, statuses []*MessageStatus) error {
	log := mlog.L(ctx)
	dependency := "mongo"
	start := time.Now()
	if len(statuses) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, len(statuses))
	now := time.Now()

	for i, status := range statuses {
		if status.ID == "" {
			status.ID = primitive.NewObjectID().Hex()
		}
		status.UpdatedAt = now
		if status.CreatedAt.IsZero() {
			status.CreatedAt = now
		}

		filter := bson.M{
			"messageId": status.MessageID,
			"userId":    status.UserID,
		}

		update := bson.M{
			"$set": status,
		}

		models[i] = mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
	}

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_BULK_UPDATE, "mongo bulk upsert message statuses"), map[string]any{
		"modelsCount": len(models),
	})

	opts := options.BulkWrite().SetOrdered(false) // Parallel writes
	_, err := r.collection.BulkWrite(ctx, models, opts)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_BULK_UPDATE, "mongo bulk upsert message statuses"), map[string]any{
			"error": err.Error(),
		})
		return errors.ErrDatabase(err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_BULK_UPDATE, "mongo bulk upsert message statuses"), map[string]any{
		"modelsCount": len(models),
	})

	return nil
}

// FindStatusForMessage finds all statuses for a message
func (r *messageStatusRepository) FindStatusForMessage(ctx context.Context, messageID string) ([]*MessageStatus, error) {
	log := mlog.L(ctx)
	dependency := "mongo"
	start := time.Now()
	filter := bson.M{"messageId": messageID}

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "mongo find statuses for message"), map[string]any{
		"filter": filter,
	})
	cursor, err := r.collection.Find(ctx, filter)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "mongo find statuses for message"), map[string]any{
			"error": err.Error(),
		})
		return nil, errors.ErrDatabase(err)
	}
	defer cursor.Close(ctx)

	var statuses []*MessageStatus
	if err := cursor.All(ctx, &statuses); err != nil {
		return nil, errors.ErrDatabase(err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "mongo find statuses for message"), map[string]any{
		"count": len(statuses),
	})

	return statuses, nil
}

// FindStatusForUser finds statuses for a user across multiple messages
func (r *messageStatusRepository) FindStatusForUser(ctx context.Context, messageIDs []string, userID string) ([]*MessageStatus, error) {
	log := mlog.L(ctx)
	dependency := "mongo"
	start := time.Now()

	filter := bson.M{
		"messageId": bson.M{"$in": messageIDs},
		"userId":    userID,
	}

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "mongo find statuses for user"), map[string]any{
		"filter": filter,
	})

	cursor, err := r.collection.Find(ctx, filter)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "mongo find statuses for user"), map[string]any{
			"error": err.Error(),
		})
		return nil, errors.ErrDatabase(err)
	}
	defer cursor.Close(ctx)

	var statuses []*MessageStatus
	if err := cursor.All(ctx, &statuses); err != nil {
		return nil, errors.ErrDatabase(err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "mongo find statuses for user"), map[string]any{
		"count": len(statuses),
	})

	return statuses, nil
}

// FindStatusByMessageAndUser finds a specific status
func (r *messageStatusRepository) FindStatusByMessageAndUser(ctx context.Context, messageID, userID string) (*MessageStatus, error) {
	log := mlog.L(ctx)
	dependency := "mongo"
	start := time.Now()
	filter := bson.M{
		"messageId": messageID,
		"userId":    userID,
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "mongo find status by message and user"), map[string]any{
		"filter": filter,
	})

	var status MessageStatus
	err := r.collection.FindOne(ctx, filter).Decode(&status)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "mongo find status by message and user"), map[string]any{
			"error": err.Error(),
		})
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, errors.ErrDatabase(err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "mongo find status by message and user"), map[string]any{
		"result": status,
	})

	return &status, nil
}

// DeleteStatusesForMessage deletes all statuses for a message
func (r *messageStatusRepository) DeleteStatusesForMessage(ctx context.Context, messageID string) error {
	dependency := "mongo"
	log := mlog.L(ctx)
	start := time.Now()
	filter := bson.M{"messageId": messageID}

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_DELETE, "mongo delete statuses for message"), map[string]any{
		"filter": filter,
	})
	result, err := r.collection.DeleteMany(ctx, filter)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "mongo delete statuses for message"), map[string]any{
			"error": err.Error(),
		})
		return errors.ErrDatabase(err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "mongo delete statuses for message"), map[string]any{
		"result": result,
	})

	return nil
}
