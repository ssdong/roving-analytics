package user

import (
	"time"
)

type User struct {
	ID                   int64     `db:"id"`
	Email                string    `db:"email"`
	PasswordHash         string    `db:"password_hash"`
	Password             string    `db:"-"`
	PasswordConfirmation string    `db:"-"`
	Name                 string    `db:"name"`
	LastSeen             time.Time `db:"last_seen"`
	TrialExpiryDate      time.Time `db:"trial_expiry_date"`
	Theme                string    `db:"theme"`
	EmailVerified        bool      `db:"email_verified"`
	// Assuming GracePeriod is a simple struct, similar to ImportedData above
	// GracePeriod         GracePeriod
	// Skipping associations for simplicity; in a real-world Go application,
	// associations are typically managed separately.
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
