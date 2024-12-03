package repository

import (
	"context"
	"fmt"
	"go-graphql-user-svc/internal/model"
	"go-graphql-user-svc/util"

	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IUserRepository interface {
	Getall(ctx context.Context) *[]model.User
	Create(ctx context.Context, user model.User) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, id string, user model.User) (*model.User, error)
	Delete(ctx context.Context, id string) error
}

type UserRepository struct {
	Collection *mongo.Collection
}

// NewUserRepository creates a new repository instance for user-related database operations
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		Collection: db.Collection("users"),
	}
}

// GetAll retrieves all user
func (r *UserRepository) Getall(ctx context.Context) *[]model.User {
	var users []model.User
	var filter = bson.M{}
	cur, err := r.Collection.Find(ctx, filter)
	if err != nil {
		log.Println(err)
	}
	err = cur.All(ctx, &users)
	log.Println(users)
	if err != nil {
		log.Println(err)
	}
	return &users
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user model.User) (*model.User, error) {
	timeNow := util.TimeNow()
	user.CreatedAt = timeNow
	user.UpdatedAt = timeNow
	result, err := r.Collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("could not insert user: %v", err)
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = model.MOID(oid.Hex())
	}
	return &user, nil
}

// FindByID retrieves a user by its ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	oid, _ := primitive.ObjectIDFromHex(id)
	err := r.Collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}
	return &user, nil
}

// Update modifies a user's details by ID
func (r *UserRepository) Update(ctx context.Context, id string, user model.User) (*model.User, error) {
	tx, err := r.Collection.Database().Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer tx.EndSession(context.TODO())

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		oid, _ := primitive.ObjectIDFromHex(id)
		update := bson.M{
			"$set": bson.M{
				"name":       user.Name,
				"email":      user.Email,
				"role":       user.Role,
				"password":   user.Password,
				"updated_at": util.TimeNow(),
			},
		}
		res, errUpdate := r.Collection.UpdateOne(sessCtx, bson.M{"_id": oid}, update)
		if errUpdate != nil || res.ModifiedCount == 0 {
			tx.AbortTransaction(sessCtx)
			return nil, fmt.Errorf("could not update user: %v", errUpdate)
		}

		if res.MatchedCount > 0 {
			user.ID = model.MOID(oid.Hex())
		}

		if res.ModifiedCount == 1 {
			errUpdate = tx.CommitTransaction(sessCtx)
			if errUpdate != nil {
				log.Println(errUpdate)
				return nil, errUpdate
			}
		}

		return res, nil
	}

	_, err = tx.WithTransaction(context.TODO(), callback)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("could not update user: %v", err)
	}

	return &user, nil
}

// Delete removes a user from the database by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := r.Collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("could not delete user: %v", err)
	}
	return nil
}
