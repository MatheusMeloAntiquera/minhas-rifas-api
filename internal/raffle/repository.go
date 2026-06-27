package raffle

import (
	"context"

	"github.com/matheusantiquera/minhas-rifas/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Create(ctx context.Context, raffle domain.Raffle) (domain.Raffle, error)
}

type repository struct {
	collection *mongo.Collection
	counters   *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		collection: db.Collection("raffles"),
		counters:   db.Collection("counters"),
	}
}

func (r *repository) nextID(ctx context.Context) (int, error) {
	filter := bson.M{"_id": "raffles"}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int `bson:"seq"`
	}

	err := r.counters.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.Seq, nil
}

func (r *repository) Create(ctx context.Context, raffle domain.Raffle) (domain.Raffle, error) {
	id, err := r.nextID(ctx)
	if err != nil {
		return domain.Raffle{}, err
	}

	raffle.ID = id

	_, err = r.collection.InsertOne(ctx, raffle)
	if err != nil {
		return domain.Raffle{}, err
	}

	return raffle, nil
}
