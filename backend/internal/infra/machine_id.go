package infra

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sony/sonyflake/v2"
)

const (
	leaseDuration   = 60 * time.Second
	refreshInterval = 20 * time.Second
)

type MachineIDLease struct {
	ID      int
	OwnerID string
	db      *sql.DB
	cancel  context.CancelFunc
	lost    chan error
}

func InitSF(ctx context.Context, db *sql.DB) (*sonyflake.Sonyflake, *MachineIDLease, error) {
	lease, err := AcquireMachineID(ctx, db)
	if err != nil {
		return nil, nil, err
	}

	sf, err := sonyflake.New(sonyflake.Settings{
		MachineID: func() (int, error) {
			return lease.ID, nil
		},
	})
	if err != nil {
		_ = lease.Release(context.Background())
		return nil, nil, err
	}

	lease.StartKeepAlive(ctx)

	return sf, lease, nil
}

func AcquireMachineID(ctx context.Context, db *sql.DB) (*MachineIDLease, error) {
	ownerID, err := randomOwnerID()
	if err != nil {
		return nil, fmt.Errorf("generate owner ID: %w", err)
	}

	// Random starting point prevents every replica fighting over ID 0.
	var randomBytes [2]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return nil, fmt.Errorf("generate starting machine ID: %w", err)
	}

	start := int(randomBytes[0])<<8 | int(randomBytes[1])

	for offset := 0; offset < 65536; offset++ {
		machineID := (start + offset) % 65536

		result, err := db.ExecContext(ctx, `
			INSERT INTO snowflake_machine_ids (
				machine_id,
				owner_id,
				expires_at
			)
			VALUES ($1, $2, NOW() + $3 * INTERVAL '1 second')
			ON CONFLICT (machine_id) DO UPDATE
			SET owner_id = EXCLUDED.owner_id,
			    expires_at = EXCLUDED.expires_at
			WHERE snowflake_machine_ids.expires_at <= NOW()
		`,
			machineID,
			ownerID,
			int(leaseDuration.Seconds()),
		)
		if err != nil {
			return nil, fmt.Errorf("claim machine ID %d: %w", machineID, err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("check machine ID claim: %w", err)
		}

		if rows == 1 {
			return &MachineIDLease{
				ID:      machineID,
				OwnerID: ownerID,
				db:      db,
				lost:    make(chan error, 1),
			}, nil
		}
	}

	return nil, errors.New("no snowflake machine IDs available")
}

func (l *MachineIDLease) StartKeepAlive(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	l.cancel = cancel

	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				err := l.refresh(ctx)
				if err != nil {
					select {
					case l.lost <- err:
					default:
					}

					return
				}
			}
		}
	}()
}

func (l *MachineIDLease) refresh(ctx context.Context) error {
	result, err := l.db.ExecContext(ctx, `
		UPDATE snowflake_machine_ids
		SET expires_at = NOW() + $3 * INTERVAL '1 second'
		WHERE machine_id = $1
		  AND owner_id = $2
		  AND expires_at > NOW()
	`,
		l.ID,
		l.OwnerID,
		int(leaseDuration.Seconds()),
	)
	if err != nil {
		return fmt.Errorf("refresh machine ID lease: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check lease refresh: %w", err)
	}

	if rows != 1 {
		return errors.New("snowflake machine ID lease was lost")
	}

	return nil
}

func (l *MachineIDLease) Release(ctx context.Context) error {
	if l.cancel != nil {
		l.cancel()
	}

	_, err := l.db.ExecContext(ctx, `
		DELETE FROM snowflake_machine_ids
		WHERE machine_id = $1
		  AND owner_id = $2
	`, l.ID, l.OwnerID)

	return err
}

func (l *MachineIDLease) Lost() <-chan error {
	return l.lost
}

func randomOwnerID() (string, error) {
	var value [16]byte

	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(value[:]), nil
}
