package auth

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth"
)

func oauthSave(db *sqlx.DB) SaveFunc {
	return func(user goth.User) (s string, err error) {
		row := db.QueryRow("select users.id from users join user_providers on users.id = user_providers.user_id where user_providers.provider_user_id = $1", user.UserID)

		var userID string
		if err := row.Scan(&userID); err != nil {
			if err != sql.ErrNoRows {
				return "", err
			}

			// New user
			err := db.QueryRow(`insert into users (email, first_name, last_name) values ($1, $2, $3) returning id`, user.Email, user.FirstName, user.LastName).Scan(&userID)
			if err != nil {
				return "", err
			}

			// Add oauth
			_, err = db.Exec("insert into user_providers (user_id, provider, provider_user_id) values ($1, $2, $3)", userID, user.Provider, user.UserID)
			if err != nil {
				return "", err
			}
		}

		return userID, nil
	}
}

