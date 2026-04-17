package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

type Session struct {
	id      uuid.UUID
	session *mongo.Session
	ctx     context.Context
	opts    *options.TransactionOptionsBuilder
	driver  *Driver
	lock    string
}

func (d *Driver) NewSession() (*Session, error) {
	return d.NewSessionWithLock("")
}

func (d *Driver) NewSessionWithLock(lock string) (*Session, error) {
	session, err := d.client.StartSession()
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority()).
		SetReadPreference(readpref.Primary())

	return &Session{
		id:      id,
		session: session,
		ctx:     context.Background(),
		opts:    txnOpts,
		driver:  d,
		lock:    lock,
	}, nil
}

func (s *Session) End() {
	defer s.session.EndSession(s.ctx)

	if s.lock != "" {
		err := s.driver.Lock().Unlock(
			s.ctx,
			s.lock,
		)
		if err != nil {
			s.driver.logger.Error().Msgf("cannot unlock lock: %v", err)
		}
	}
}

func (s *Session) WithTransaction(
	fn func(context.Context) (any, error),
) (any, error) {
	wrapper := func(ctx context.Context) (any, error) {
		if s.lock != "" {
			for i := range 5 {
				err := s.driver.Lock().Lock(
					s.ctx,
					s.lock,
					s.id,
				)
				if err == nil {
					break
				}

				// handle retry
				if errors.Is(err, ErrLocked) && i < 4 {
					time.Sleep(100 * time.Millisecond)
					continue
				}

				return nil, fmt.Errorf("lock: %w", err)
			}
		}

		return fn(ctx)
	}
	return s.session.WithTransaction(s.ctx, wrapper, s.opts)
}
