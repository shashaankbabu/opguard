package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/pusher/pusher-http-go"
)

func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {
	userID := repo.App.Session.GetInt(r.Context(), "userID")

	u, _ := repo.DB.GetUserById(userID)
	params, _ := io.ReadAll(r.Body)
	presenceData := pusher.MemberData{
		UserID: strconv.Itoa(userID),
		UserInfo: map[string]string{
			"name": u.FirstName,
			"id":   strconv.Itoa(userID),
		},
	}

	response, err := app.WsClient.AuthenticatePresenceChannel(params, presenceData)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Typee", "application/json")
	_, _ = w.Write(response)
}

// SendPrivateMessage is Sample code for sending to private channel
func (repo *DBRepo) SendPrivateMessage(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	id := r.URL.Query().Get("id")

	data := make(map[string]string)

	data["message"] = msg
	_ = repo.App.WsClient.Trigger(fmt.Sprintf("private-channel-%s", id), "private-message", data)
}
