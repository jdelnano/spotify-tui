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
	auth   *spotifyauth.Authenticator
	ch     chan *spotify.Client
	state  string
	config *oauth2.Config
}

func NewSpotifyAuth(clientID, clientSecret string) *SpotifyAuth {
	scopes := []string{
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
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(scopes...),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)

	// Create oauth2 config for token refresh
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}

	return &SpotifyAuth{
		auth:   auth,
		ch:     make(chan *spotify.Client),
		state:  "abc123",
		config: config,
	}
}

func (s *SpotifyAuth) GetAuthURL() string {
	return s.auth.AuthURL(s.state)
}

func (s *SpotifyAuth) Authenticate() (*spotify.Client, *oauth2.Token, error) {
	server := &http.Server{Addr: ":8080"}
	tokenCh := make(chan *oauth2.Token)

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
		tokenCh <- token

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
	token := <-tokenCh

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("You are logged in as:", user.ID)

	return client, token, nil
}

func (s *SpotifyAuth) ClientFromToken(token *oauth2.Token) (*spotify.Client, *oauth2.Token, error) {
	ctx := context.Background()

	// Create a token source that can refresh the token if needed
	src := s.config.TokenSource(ctx, token)

	// Get the token (will refresh if expired)
	newToken, err := src.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get/refresh token: %w", err)
	}

	// Create HTTP client with the token
	httpClient := oauth2.NewClient(ctx, src)
	client := spotify.New(httpClient)

	// Verify the token works by making a test API call
	_, err = client.CurrentUser(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("token validation failed: %w", err)
	}

	return client, newToken, nil
}

func (s *SpotifyAuth) RefreshToken(token *oauth2.Token) (*spotify.Client, error) {
	client := spotify.New(s.auth.Client(context.Background(), token))
	return client, nil
}
