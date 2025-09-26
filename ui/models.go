package ui

import (
	"github.com/jdelnano/spotify-tui/spotify"
	spotifyPkg "github.com/zmb3/spotify/v2"
)

type ViewMode int

const (
	PlaylistView ViewMode = iota
	SearchView
	NowPlayingView
)

type Model struct {
	spotifyClient    *spotify.Client
	viewMode         ViewMode
	playlists        []spotifyPkg.SimplePlaylist
	searchResults    []spotifyPkg.FullTrack
	playlistTracks   []spotifyPkg.PlaylistTrack
	searchQuery      string
	playlistCursor   int
	trackCursor      int
	searchCursor     int
	selectedPlaylist *spotifyPkg.SimplePlaylist
	currentlyPlaying *spotifyPkg.CurrentlyPlaying
	width            int
	height           int
	isSearching      bool
	statusMessage    string
	err              error
}

