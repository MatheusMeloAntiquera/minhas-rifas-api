package user

import (
	"context"

	"github.com/matheusantiquera/minhas-rifas/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	Update(ctx context.Context, id int, user domain.User) (domain.User, error)
	Delete(ctx context.Context, id int) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type repository struct {
	collection *mongo.Collection
	counters   *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	coll := db.Collection("users")

	coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &repository{
		collection: coll,
		counters:   db.Collection("counters"),
	}
}

func (r *repository) nextID(ctx context.Context) (int, error) {
	filter := bson.M{"_id": "users"}
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

func (r *repository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	id, err := r.nextID(ctx)
	if err != nil {
		return domain.User{}, err
	}

	user.ID = id

	_, err = r.collection.InsertOne(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *repository) Update(ctx context.Context, id int, user domain.User) (domain.User, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"name":       user.Name,
		"email":      user.Email,
		"updated_at": user.UpdatedAt,
	}}

	result := r.collection.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var updated domain.User
	if err := result.Decode(&updated); err != nil {
		return domain.User{}, err
	}

	return updated, nil
}

func (r *repository) Delete(ctx context.Context, id int) error {
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	filter := bson.M{"email": email}

	var user domain.User
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
