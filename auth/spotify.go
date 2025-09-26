package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const redirectURI = "http://127.0.0.1:8080/callback"

type SpotifyAuth struct {
	auth  *spotifyauth.Authenticator
	ch    chan *spotify.Client
	state string
}

func NewSpotifyAuth(clientID, clientSecret string) *SpotifyAuth {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserReadEmail,
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadRecentlyPlayed,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserLibraryModify,
			spotifyauth.ScopeUserTopRead,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistReadCollaborative,
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopePlaylistModifyPrivate,
			spotifyauth.ScopeStreaming,
		),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)

	return &SpotifyAuth{
		auth:  auth,
		ch:    make(chan *spotify.Client),
		state: "abc123",
	}
}

func (s *SpotifyAuth) GetAuthURL() string {
	return s.auth.AuthURL(s.state)
}

func (s *SpotifyAuth) Authenticate() (*spotify.Client, error) {
	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token, err := s.auth.Token(r.Context(), s.state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != s.state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, s.state)
		}

		client := spotify.New(s.auth.Client(r.Context(), token))
		fmt.Fprintf(w, "Login Completed! You can close this window and return to the terminal.")
		s.ch <- client

		go func() {
			time.Sleep(1 * time.Second)
			server.Shutdown(context.Background())
		}()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	fmt.Printf("Please log in to Spotify by visiting the following page:\n%s\n", s.GetAuthURL())

	client := <-s.ch

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		return nil, err
	}
	fmt.Println("You are logged in as:", user.ID)

	return client, nil
}

func (s *SpotifyAuth) RefreshToken(token *oauth2.Token) (*spotify.Client, error) {
	client := spotify.New(s.auth.Client(context.Background(), token))
	return client, nil
}
