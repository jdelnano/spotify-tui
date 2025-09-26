package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jdelnano/spotify-tui/internal/auth"
	"github.com/jdelnano/spotify-tui/internal/config"
	"github.com/jdelnano/spotify-tui/internal/spotify"
	"github.com/jdelnano/spotify-tui/internal/ui"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Check if we need to set up Spotify credentials
	if !cfg.IsConfigured() {
		fmt.Println("Welcome to Spotify TUI Player!")
		fmt.Println("\nTo use this app, you need to create a Spotify App:")
		fmt.Println("1. Go to https://developer.spotify.com/dashboard")
		fmt.Println("2. Create a new app")
		fmt.Println("3. Add http://127.0.0.1:8080/callback as a redirect URI")
		fmt.Println("4. Copy your Client ID and Client Secret")

		if err := cfg.PromptForCredentials(); err != nil {
			log.Fatal("Failed to save credentials:", err)
		}
	}

	// Set up Spotify authentication
	spotifyAuth := auth.NewSpotifyAuth(cfg.ClientID, cfg.ClientSecret)

	// Check if we have a saved token
	var spotifyClient *spotify.Client
	tokenUpdated := false

	if cfg.Token != nil {
		// Try to use the saved token (will automatically refresh if expired)
		client, newToken, err := spotifyAuth.ClientFromToken(cfg.Token)
		if err == nil {
			spotifyClient = spotify.NewClient(client)
			// Update the token in config if it was refreshed
			if newToken.AccessToken != cfg.Token.AccessToken {
				cfg.Token = newToken
				tokenUpdated = true
			}
		} else {
			fmt.Printf("Cached token invalid or expired: %v\n", err)
		}
	}

	// If we don't have a client yet, authenticate from scratch
	if spotifyClient == nil {
		fmt.Println("\nAuthenticating with Spotify...")
		client, token, err := spotifyAuth.Authenticate()
		if err != nil {
			log.Fatal("Failed to authenticate:", err)
		}
		spotifyClient = spotify.NewClient(client)
		cfg.Token = token
		tokenUpdated = true
	}

	// Save the updated token to config if it changed
	if tokenUpdated {
		if err := cfg.Save(); err != nil {
			log.Printf("Warning: Failed to save token to config: %v", err)
		}
	}

	// Create the TUI model
	model := ui.NewModel(spotifyClient)

	// Start the TUI
	p := tea.NewProgram(&model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
