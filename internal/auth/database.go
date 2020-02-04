package auth

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth"
)

func oauthSave(db *sqlx.DB) SaveFunc {
	return func(user goth.User) (s string, err error) {
		row := db.QueryRow("select users.id from users join oauth on users.id = oauth.user_id where oauth.provider_user_id = $1", user.UserID)

		var userID string
		if err := row.Scan(&userID); err != nil {
			if err != sql.ErrNoRows {
				return "", err
			}

			// Insert
			//db.Exec("insert into oauth(user_id, provider, provider_user_id) values()")
		}

		return userID, nil
	}
}

