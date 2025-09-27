package ui

import (
	"github.com/jdelnano/spotify-tui/internal/spotify"
	spotifyPkg "github.com/zmb3/spotify/v2"
)

func NewModel(spotifyClient *spotify.Client) Model {
	return Model{
		spotifyClient:  spotifyClient,
		viewMode:       PlaylistView,
		playlists:      []spotifyPkg.SimplePlaylist{},
		searchResults:  []spotifyPkg.FullTrack{},
		playlistTracks: []spotifyPkg.PlaylistTrack{},
		playlistCursor: 0,
		libraryCursor:  -1,
		trackCursor:    0,
		searchCursor:   0,
		width:          80,
		height:         24,
	}
}

