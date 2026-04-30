package db

import (
	"context"
	"errors"
	"time"

	"github.com/Kryvea/Kryvea/internal/crypto"
	"github.com/Kryvea/Kryvea/internal/model"
	"github.com/Kryvea/Kryvea/internal/store"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserIndex struct{ driver *Driver }

func (ui *UserIndex) Insert(ctx context.Context, user *model.User, password string) (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}
	hash, err := crypto.Encrypt(password)
	if err != nil {
		return uuid.Nil, err
	}

	role := user.Role
	if role == "" {
		role = model.RoleUser
	}

	row := &dbUser{
		ID:             id,
		DisabledAt:     timePtrOrTimeNever(user.DisabledAt),
		Username:       user.Username,
		Password:       hash,
		PasswordExpiry: timeOrTimeNever(user.PasswordExpiry),
		Role:           role,
	}
	if _, err := idbFrom(ctx, ui.driver.db).NewInsert().Model(row).Exec(ctx); err != nil {
		return uuid.Nil, mapErr(err)
	}

	if err := ui.insertUserCustomers(ctx, id, user.Customers); err != nil {
		return uuid.Nil, err
	}
	user.ID = id
	return id, nil
}

func (ui *UserIndex) insertUserCustomers(ctx context.Context, userID uuid.UUID, customers []model.Customer) error {
	rows := make([]dbUserCustomer, 0, len(customers))
	for _, c := range customers {
		if c.ID == uuid.Nil {
			continue
		}
		rows = append(rows, dbUserCustomer{UserID: userID, CustomerID: c.ID})
	}
	if len(rows) == 0 {
		return nil
	}
	_, err := idbFrom(ctx, ui.driver.db).NewInsert().
		Model(&rows).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) Login(ctx context.Context, username, password string) (*model.User, error) {
	var row dbUser
	err := idbFrom(ctx, ui.driver.db).NewSelect().
		Model(&row).
		Where("username = ?", username).
		Scan(ctx)
	if err != nil {
		if errors.Is(mapErr(err), store.ErrNotFound) {
			return nil, store.ErrInvalidCredentials
		}
		return nil, mapErr(err)
	}
	if !crypto.Compare(password, row.Password) {
		return nil, store.ErrInvalidCredentials
	}
	if row.DisabledAt.Before(time.Now()) {
		return nil, store.ErrDisabledUser
	}

	u := row.toModel()
	u.Token = crypto.NewToken()
	u.TokenExpiry = time.Now().Add(model.TokenExpireTime)
	if !u.PasswordExpiry.IsZero() && u.PasswordExpiry.Before(time.Now()) {
		u.TokenExpiry = time.Now().Add(model.TokenExpireTimePwdReset)
	}

	if _, err := idbFrom(ctx, ui.driver.db).NewUpdate().
		Model((*dbUser)(nil)).
		Set("token = ?", []byte(u.Token)).
		Set("token_expiry = ?", u.TokenExpiry).
		Where("id = ?", u.ID).
		Exec(ctx); err != nil {
		return nil, mapErr(err)
	}
	return &u, nil
}

func (ui *UserIndex) RefreshUserToken(ctx context.Context, user *model.User) error {
	user.TokenExpiry = user.TokenExpiry.Add(model.TokenExtendTime)
	_, err := idbFrom(ctx, ui.driver.db).NewUpdate().
		Model((*dbUser)(nil)).
		Set("token_expiry = ?", user.TokenExpiry).
		Where("id = ?", user.ID).
		Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) Logout(ctx context.Context, id uuid.UUID) error {
	_, err := idbFrom(ctx, ui.driver.db).NewUpdate().
		Model((*dbUser)(nil)).
		Set("token = NULL").
		Set("token_expiry = NULL").
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) Get(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u, err := ui.fetchOneWithRelations(ctx, "u.id = ?", id)
	if err != nil {
		return nil, err
	}
	u.Password = nil
	u.Token = nil
	return u, nil
}

func (ui *UserIndex) GetAll(ctx context.Context) ([]model.User, error) {
	var rows []dbUser
	if err := ui.selectWithRelations(ctx, &rows).
		OrderExpr("username").
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	users := make([]model.User, len(rows))
	for i := range rows {
		u := rowToUserWithRelations(&rows[i])
		u.Password = nil
		u.Token = nil
		users[i] = u
	}
	return users, nil
}

func (ui *UserIndex) GetAllUsernames(ctx context.Context) ([]string, error) {
	var names []string
	if err := idbFrom(ctx, ui.driver.db).NewSelect().
		Model((*dbUser)(nil)).
		Column("username").
		OrderExpr("username").
		Scan(ctx, &names); err != nil {
		return nil, mapErr(err)
	}
	return names, nil
}

func (ui *UserIndex) GetByToken(ctx context.Context, token crypto.Token) (*model.User, error) {
	u, err := ui.fetchOneWithRelations(ctx, "token = ?", []byte(token))
	if err != nil {
		return nil, err
	}
	u.Password = nil
	return u, nil
}

func (ui *UserIndex) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	u, err := ui.fetchOneWithRelations(ctx, "username = ?", username)
	if err != nil {
		return nil, err
	}
	u.Password = nil
	return u, nil
}

func (ui *UserIndex) Update(ctx context.Context, id uuid.UUID, user *model.User) error {
	idb := idbFrom(ctx, ui.driver.db)
	q := idb.NewUpdate().Model((*dbUser)(nil)).Where("id = ?", id)
	dirty := false
	if !user.DisabledAt.IsZero() {
		q = q.Set("disabled_at = ?", user.DisabledAt)
		dirty = true
	}
	if user.Username != "" {
		q = q.Set("username = ?", user.Username)
		dirty = true
	}
	if user.Role != "" {
		q = q.Set("role = ?", user.Role)
		dirty = true
	}
	if dirty {
		if _, err := q.Exec(ctx); err != nil {
			return mapErr(err)
		}
	}

	if user.Customers != nil {
		if _, err := idb.NewDelete().
			Model((*dbUserCustomer)(nil)).
			Where("user_id = ?", id).
			Exec(ctx); err != nil {
			return mapErr(err)
		}
		if err := ui.insertUserCustomers(ctx, id, user.Customers); err != nil {
			return err
		}
	}

	count, err := idb.NewSelect().
		Model((*dbUser)(nil)).
		Where("role = ?", model.RoleAdmin).
		Count(ctx)
	if err != nil {
		return mapErr(err)
	}
	if count == 0 {
		return store.ErrAdminUserRequired
	}
	return nil
}

func (ui *UserIndex) UpdateMe(ctx context.Context, id uuid.UUID, user *model.User, password string) error {
	q := idbFrom(ctx, ui.driver.db).NewUpdate().Model((*dbUser)(nil)).Where("id = ?", id)
	dirty := false
	if user.Username != "" {
		q = q.Set("username = ?", user.Username)
		dirty = true
	}
	if password != "" {
		hash, err := crypto.Encrypt(password)
		if err != nil {
			return err
		}
		q = q.Set("password = ?", hash)
		dirty = true
	}
	if !dirty {
		return nil
	}
	_, err := q.Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) UpdateOwnedAssessment(ctx context.Context, userID, assessmentID uuid.UUID, addToOwned bool) error {
	idb := idbFrom(ctx, ui.driver.db)
	if addToOwned {
		_, err := idb.NewInsert().
			Model(&dbUserAssessment{UserID: userID, AssessmentID: assessmentID}).
			On("CONFLICT DO NOTHING").
			Exec(ctx)
		return mapErr(err)
	}
	_, err := idb.NewDelete().
		Model((*dbUserAssessment)(nil)).
		Where("user_id = ?", userID).
		Where("assessment_id = ?", assessmentID).
		Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) AssignCustomer(ctx context.Context, userID, customerID uuid.UUID) error {
	_, err := idbFrom(ctx, ui.driver.db).NewInsert().
		Model(&dbUserCustomer{UserID: userID, CustomerID: customerID}).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) Delete(ctx context.Context, id uuid.UUID) error {
	idb := idbFrom(ctx, ui.driver.db)
	if _, err := idb.NewDelete().
		Model((*dbUser)(nil)).
		Where("id = ?", id).
		Exec(ctx); err != nil {
		return mapErr(err)
	}
	count, err := idb.NewSelect().
		Model((*dbUser)(nil)).
		Where("role = ?", model.RoleAdmin).
		Count(ctx)
	if err != nil {
		return mapErr(err)
	}
	if count == 0 {
		return store.ErrAdminUserRequired
	}
	return nil
}

func (ui *UserIndex) ResetUserPassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hash, err := crypto.Encrypt(newPassword)
	if err != nil {
		return err
	}
	_, err = idbFrom(ctx, ui.driver.db).NewUpdate().
		Model((*dbUser)(nil)).
		Set("password = ?", hash).
		Set("password_expiry = now()").
		Where("id = ?", id).
		Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) ResetPassword(ctx context.Context, user *model.User, password string) error {
	hash, err := crypto.Encrypt(password)
	if err != nil {
		return err
	}
	user.PasswordExpiry = model.TimeNever
	user.Token = crypto.NewToken()
	user.TokenExpiry = time.Now().Add(model.TokenExpireTime)

	_, err = idbFrom(ctx, ui.driver.db).NewUpdate().
		Model((*dbUser)(nil)).
		Set("password = ?", hash).
		Set("password_expiry = ?", user.PasswordExpiry).
		Set("token = ?", []byte(user.Token)).
		Set("token_expiry = ?", user.TokenExpiry).
		Where("id = ?", user.ID).
		Exec(ctx)
	return mapErr(err)
}

func (ui *UserIndex) ValidatePassword(ctx context.Context, id uuid.UUID, currentPassword string) error {
	var row dbUser
	err := idbFrom(ctx, ui.driver.db).NewSelect().
		Model(&row).
		Column("password").
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(mapErr(err), store.ErrNotFound) {
			return store.ErrInvalidCredentials
		}
		return mapErr(err)
	}
	if !crypto.Compare(currentPassword, row.Password) {
		return store.ErrInvalidCredentials
	}
	return nil
}

func (ui *UserIndex) selectWithRelations(ctx context.Context, dest any) *bun.SelectQuery {
	return idbFrom(ctx, ui.driver.db).NewSelect().
		Model(dest).
		Relation("Customers", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("name")
		}).
		Relation("Assessments", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Order("created_at DESC")
		})
}

func (ui *UserIndex) fetchOneWithRelations(ctx context.Context, where string, args ...any) (*model.User, error) {
	var row dbUser
	if err := ui.selectWithRelations(ctx, &row).
		Where(where, args...).
		Scan(ctx); err != nil {
		return nil, mapErr(err)
	}
	u := rowToUserWithRelations(&row)
	return &u, nil
}

func rowToUserWithRelations(r *dbUser) model.User {
	u := r.toModel()
	u.Customers = make([]model.Customer, len(r.Customers))
	for i := range r.Customers {
		u.Customers[i] = r.Customers[i].toModel()
	}
	u.Assessments = make([]model.Assessment, len(r.Assessments))
	for i := range r.Assessments {
		u.Assessments[i] = r.Assessments[i].toModel()
	}
	return u
}

func timeOrTimeNever(t time.Time) time.Time {
	if t.IsZero() {
		return model.TimeNever
	}
	return t
}

func timePtrOrTimeNever(t time.Time) *time.Time {
	if t.IsZero() {
		return &model.TimeNever
	}
	return &t
}
