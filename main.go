package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jdelnano/spotify-tui/auth"
	"github.com/jdelnano/spotify-tui/config"
	"github.com/jdelnano/spotify-tui/spotify"
	"github.com/jdelnano/spotify-tui/ui"
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
	if cfg.Token != nil {
		// Try to use the saved token
		client, err := spotifyAuth.RefreshToken(cfg.Token)
		if err == nil {
			spotifyClient = spotify.NewClient(client)
		}
	}

	// If we don't have a client yet, authenticate
	if spotifyClient == nil {
		fmt.Println("\nAuthenticating with Spotify...")
		client, err := spotifyAuth.Authenticate()
		if err != nil {
			log.Fatal("Failed to authenticate:", err)
		}
		spotifyClient = spotify.NewClient(client)
	}

	// Create the TUI model
	model := ui.NewModel(spotifyClient)

	// Start the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
