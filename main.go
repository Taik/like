package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/zmb3/spotify"
)

const (
	listenURL   = "localhost:55510"
	callbackURL = "http://localhost:55510/callback"
)

func generateState() string {
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		log.Fatalf("Unable to generate random state: %s", err)
	}
	return hex.EncodeToString(seed)
}

func main() {
	auth := spotify.NewAuthenticator(
		callbackURL,
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserLibraryModify,
	)
	state := generateState()
	log.Print(auth.AuthURL(state))

	clientCh := make(chan *spotify.Client)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Got callback call")
		token, err := auth.Token(state, r)
		if err != nil {
			http.Error(w, "Unable to parse token", http.StatusForbidden)
			log.Fatalf("Unable to parse token: %s", err)
		}
		client := auth.NewClient(token)

		select {
		case clientCh <- &client:
			// TODO(thinh): Auto close page
			w.WriteHeader(http.StatusNoContent)

		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusGatewayTimeout)
		}
	})

	go http.ListenAndServe(listenURL, nil)

	client := <-clientCh
	playing, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Fatalf("Unable to get currently playing song: %s", err)
	}
	log.Printf("%v", playing)
}
