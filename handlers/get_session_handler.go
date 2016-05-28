package handlers

import (
	"crypto/rsa"
	"database/sql"
	"net/http"
	"time"

	"pixur.org/pixur/schema"
	"pixur.org/pixur/tasks"
)

type GetSessionHandler struct {
	// embeds
	http.Handler

	// deps
	DB         *sql.DB
	Runner     *tasks.TaskRunner
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func (h *GetSessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
		return
	}

	var task = &tasks.AuthUserTask{
		DB:     h.DB,
		Now:    time.Now,
		Email:  r.FormValue("ident"),
		Secret: r.FormValue("secret"),
	}
	runner := new(tasks.TaskRunner)
	if err := runner.Run(task); err != nil {
		returnTaskError(w, err)
		return
	}

	enc := JwtEncoder{
		PrivateKey: h.PrivateKey,
		Now:        time.Now,
		Expiration: time.Hour * 24 * 365 * 10,
	}
	jwt, err := enc.Encode(&JwtPayload{
		Subject: schema.Varint(task.User.UserId).Encode(),
	})
	if err != nil {
		returnTaskError(w, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    string(jwt),
		Path:     "/api/",
		Expires:  time.Now().Add(enc.Expiration),
		Secure:   true,
		HttpOnly: true,
	})

	resp := GetSessionResponse{
		UserId: schema.Varint(task.User.UserId).Encode(),
	}

	returnProtoJSON(w, r, &resp)
}

func init() {
	register(func(mux *http.ServeMux, c *ServerConfig) {
		mux.Handle("/api/getSession", &GetSessionHandler{
			DB:         c.DB,
			PrivateKey: c.PrivateKey,
			PublicKey:  c.PublicKey,
		})
	})
}
