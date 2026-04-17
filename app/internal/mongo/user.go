package mongo

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/Kryvea/Kryvea/internal/crypto"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrDisabledUser       = errors.New("user is disabled")
	ErrInvalidCredentials = errors.New("invalid credentials")

	TimeNever = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
)

const (
	userCollection = "user"

	RoleAdmin = "admin"
	RoleUser  = "user"

	TokenExpireTime         = 9 * time.Hour
	TokenExpireTimePwdReset = 15 * time.Minute
	TokenExtendTime         = 2 * time.Hour
	TokenRefreshThreshold   = 1 * time.Hour
)

var (
	Roles = []string{RoleAdmin, RoleUser}
)

var UserProjection = bson.M{
	"password":        0,
	"password_expiry": 0,
	"token":           0,
	"token_expiry":    0,
}

type User struct {
	Model          `bson:",inline"`
	DisabledAt     time.Time    `json:"disabled_at,omitempty" bson:"disabled_at"`
	Username       string       `json:"username" bson:"username"`
	Password       []byte       `json:"-" bson:"password"`
	PasswordExpiry time.Time    `json:"-" bson:"password_expiry"`
	Token          crypto.Token `json:"-" bson:"token"`
	TokenExpiry    time.Time    `json:"-" bson:"token_expiry"`
	Role           string       `json:"role" bson:"role"`
	Customers      []Customer   `json:"customers,omitempty" bson:"customers"`
	Assessments    []Assessment `json:"assessments,omitempty" bson:"assessments"`
}

type UserIndex struct {
	driver     *Driver
	collection *mongo.Collection
}

func (d *Driver) User() *UserIndex {
	return &UserIndex{
		driver:     d,
		collection: d.database.Collection(userCollection),
	}
}

func (ui UserIndex) init() error {
	_, err := ui.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "username", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (ui *UserIndex) Insert(ctx context.Context, user *User, password string) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.UUID{}, err
	}

	user.Model = Model{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user.DisabledAt = TimeNever
	user.PasswordExpiry = TimeNever

	if user.Customers == nil {
		user.Customers = []Customer{}
	}

	if user.Assessments == nil {
		user.Assessments = []Assessment{}
	}

	hash, err := crypto.Encrypt(password)
	if err != nil {
		return uuid.UUID{}, err
	}
	user.Password = hash

	_, err = ui.collection.InsertOne(ctx, user)
	if err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (ui *UserIndex) Login(ctx context.Context, username, password string) (*User, error) {
	var user User
	err := ui.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}

	if !crypto.Compare(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	if user.DisabledAt.Before(time.Now()) {
		return nil, ErrDisabledUser
	}

	user.Token = crypto.NewToken()

	user.TokenExpiry = time.Now().Add(TokenExpireTime)
	if user.PasswordExpiry.Before(time.Now()) {
		user.TokenExpiry = time.Now().Add(TokenExpireTimePwdReset)
	}

	_, err = ui.collection.UpdateOne(ctx, bson.M{"username": username}, bson.M{
		"$set": bson.M{
			"token":        user.Token,
			"token_expiry": user.TokenExpiry,
		}})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ui *UserIndex) RefreshUserToken(ctx context.Context, user *User) error {
	user.TokenExpiry = user.TokenExpiry.Add(TokenExtendTime)

	_, err := ui.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"token_expiry": user.TokenExpiry,
		}})
	if err != nil {
		return err
	}

	return nil
}

func (ui *UserIndex) Logout(ctx context.Context, ID uuid.UUID) error {
	_, err := ui.collection.UpdateOne(ctx, bson.M{"_id": ID}, bson.M{
		"$set": bson.M{
			"token":        crypto.TokenNil,
			"token_expiry": time.Time{},
		}})
	return err
}

func (ui *UserIndex) Get(ctx context.Context, ID uuid.UUID) (*User, error) {
	filter := bson.M{"_id": ID}
	opts := options.FindOne().SetProjection(UserProjection)

	user := &User{}
	err := ui.collection.FindOne(ctx, filter, opts).Decode(user)
	if err != nil {
		return nil, err
	}

	err = ui.hydrate(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ui *UserIndex) GetByIDForHydrate(ctx context.Context, ID uuid.UUID) (*User, error) {
	filter := bson.M{"_id": ID}
	opts := options.FindOne().SetProjection(bson.M{
		"username": 1,
	})

	var user User
	err := ui.collection.FindOne(ctx, filter, opts).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ui *UserIndex) GetAll(ctx context.Context) ([]User, error) {
	filter := bson.M{}
	opts := options.Find().SetProjection(UserProjection)

	cursor, err := ui.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	users := []User{}
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	for i := range users {
		err = ui.hydrate(ctx, &users[i])
		if err != nil {
			return nil, err
		}
	}

	return users, nil
}

func (ui *UserIndex) GetAllUsernames(ctx context.Context) ([]string, error) {
	opts := options.Find().SetProjection(bson.M{
		"username": 1,
	}).SetSort(bson.M{
		"username": 1,
	})
	cursor, err := ui.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(ctx)

	usernames := make([]string, 0, cursor.RemainingBatchLength())
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			return []string{}, err
		}

		usernames = append(usernames, user.Username)
	}

	return usernames, err
}

func (ui *UserIndex) GetByToken(ctx context.Context, token crypto.Token) (*User, error) {
	opts := options.FindOne().SetProjection(bson.M{
		"password": 0,
	})

	var user User
	if err := ui.collection.FindOne(ctx, bson.M{"token": token}, opts).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (ui *UserIndex) GetByUsername(ctx context.Context, username string) (*User, error) {
	opts := options.FindOne().SetProjection(bson.M{
		"password": 0,
	})

	var user User
	if err := ui.collection.FindOne(ctx, bson.M{"username": username}, opts).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (ui *UserIndex) Search(ctx context.Context, query string) ([]User, error) {
	filter := bson.M{
		"username": bson.M{"$regex": bson.Regex{Pattern: regexp.QuoteMeta(query), Options: "i"}},
	}
	opts := options.Find().SetProjection(UserProjection)

	cursor, err := ui.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	users := []User{}
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Update modifies an existing user
//
// Requires transactional context to ensure data integrity
func (ui *UserIndex) Update(ctx context.Context, ID uuid.UUID, user *User) error {
	filter := bson.M{"_id": ID}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if !user.DisabledAt.IsZero() {
		update["$set"].(bson.M)["disabled_at"] = user.DisabledAt
	}
	if user.Username != "" {
		update["$set"].(bson.M)["username"] = user.Username
	}
	if user.Role != "" {
		update["$set"].(bson.M)["role"] = user.Role
	}
	if user.Customers != nil {
		update["$set"].(bson.M)["customers"] = user.Customers
	}

	_, err := ui.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	count, err := ui.collection.CountDocuments(
		ctx,
		bson.M{"role": RoleAdmin},
	)
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrAdminUserRequired
	}

	return err
}

// UpdateMe modifies the current user's own information
//
// Requires transactional context to ensure data integrity
func (ui *UserIndex) UpdateMe(ctx context.Context, userID uuid.UUID, newUser *User, password string) error {
	filter := bson.M{"_id": userID}

	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if newUser.Username != "" {
		update["$set"].(bson.M)["username"] = newUser.Username
	}

	if password != "" {
		hash, err := crypto.Encrypt(password)
		if err != nil {
			return err
		}
		update["$set"].(bson.M)["password"] = hash
	}

	_, err := ui.collection.UpdateOne(ctx, filter, update)
	return err
}

func (ui *UserIndex) UpdateOwnedAssessment(ctx context.Context, userID, assessmentID uuid.UUID, addToOwned bool) error {
	filter := bson.M{"_id": userID}

	op := "$pull"
	if addToOwned {
		op = "$addToSet"
	}

	update := bson.M{
		op: bson.M{
			"assessments": bson.M{
				"_id": assessmentID,
			},
		},
	}

	_, err := ui.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete removes a user and its references from vulnerabilities
//
// Requires transactional context to ensure data integrity
func (ui *UserIndex) Delete(ctx context.Context, ID uuid.UUID) error {
	// remove user from vulnerability
	err := ui.driver.Vulnerability().RemoveUserReference(ctx, ID)
	if err != nil {
		return err
	}

	// delete user
	_, err = ui.collection.DeleteOne(ctx, bson.M{"_id": ID})
	return err
}

func (ui *UserIndex) ResetUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	hash, err := crypto.Encrypt(newPassword)
	if err != nil {
		return err
	}

	_, err = ui.collection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$set": bson.M{
			"updated_at":      time.Now(),
			"password":        hash,
			"password_expiry": time.Now(),
		}})
	if err != nil {
		return err
	}

	return nil
}

func (ui *UserIndex) ResetPassword(ctx context.Context, user *User, password string) error {
	hash, err := crypto.Encrypt(password)
	if err != nil {
		return err
	}

	user.UpdatedAt = time.Now()
	user.PasswordExpiry = TimeNever

	user.Token = crypto.NewToken()
	user.TokenExpiry = time.Now().Add(TokenExpireTime)

	_, err = ui.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"updated_at":      user.UpdatedAt,
			"password":        hash,
			"password_expiry": user.PasswordExpiry,
			"token":           user.Token,
			"token_expiry":    user.TokenExpiry,
		}})
	return err
}

func (ui *UserIndex) ValidatePassword(ctx context.Context, ID uuid.UUID, currentPassword string) error {
	opts := options.FindOne().SetProjection(bson.M{
		"password": 1,
	})

	var user User
	err := ui.collection.FindOne(ctx, bson.M{"_id": ID}, opts).Decode(&user)
	if err != nil {
		return err
	}

	if !crypto.Compare(currentPassword, user.Password) {
		return ErrInvalidCredentials
	}

	return nil
}

func (u *User) CanAccessCustomer(customer uuid.UUID) bool {
	if u.Role == RoleAdmin {
		return true
	}

	for _, allowedCustomer := range u.Customers {
		if allowedCustomer.ID == customer {
			return true
		}
	}
	return false
}

func IsValidRole(role string) bool {
	if role == "" {
		return false
	}

	for _, r := range Roles {
		if r == role {
			return true
		}
	}

	return false
}

// hydrate fills in the nested fields for a User
func (ui *UserIndex) hydrate(ctx context.Context, user *User) error {
	// customers are optional
	if len(user.Customers) > 0 {
		customers, err := ui.driver.Customer().GetManyForHydrate(ctx, user.Customers)
		if err != nil {
			return err
		}

		user.Customers = customers
	}

	// assessments are optional
	if len(user.Assessments) > 0 {
		assessment, err := ui.driver.Assessment().GetManyForHydrate(ctx, user.Assessments)
		if err != nil {
			return err
		}

		user.Assessments = assessment
	}

	return nil
}
