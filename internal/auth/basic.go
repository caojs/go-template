package auth

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type basic struct {
	db *sqlx.DB
}

func NewBasic(db *sqlx.DB) *basic {
	return &basic{
		db:db,
	}
}

func (b *basic) signUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "username or password is required", http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	if username == "" || password == "" {
		http.Error(w, "username or password is required", http.StatusBadRequest)
		return
	}

	var userID string
	if err := b.db.QueryRow("select user_id from user_accounts where username=$1", username).Scan(&userID); err == nil {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	} else {
		if err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := b.db.QueryRow("insert into users(first_name) values($1) returning id", "").Scan(&userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := b.db.Exec("insert into user_accounts (user_id, username, password) values($1, $2, $3)", userID, username, hashPass); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}


	cookie, _ := createCookie(userID)
	http.SetCookie(w, cookie)
	w.Write([]byte(""))
}

func (b *basic) signIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "username or password is required", http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	if username == "" || password == "" {
		http.Error(w, "username or password is required", http.StatusBadRequest)
		return
	}

	var userID string
	var hashPass string
	if err := b.db.QueryRow("select user_id, password from user_accounts where username=$1", username).Scan(&userID, &hashPass); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User does not exist", http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(password)); err != nil {
		http.Error(w, "Password not match", http.StatusBadRequest)
		return
	}

	cookie, err := createCookie(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	w.Write([]byte(""))
}

func (b *basic) logout(w http.ResponseWriter, _ *http.Request) {
	cookie := expireCookie()
	http.SetCookie(w, cookie)
	w.Write([]byte(""))
}
