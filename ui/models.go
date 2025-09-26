package ui

import (
	"github.com/jdelnano/spotify-tui/internal/spotify"
	spotifyPkg "github.com/zmb3/spotify/v2"
)

type ViewMode int

const (
	PlaylistView ViewMode = iota
	SearchView
	NowPlayingView
)

type LibraryCategory struct {
	Name       string
	Icon       string
	Type       string
	ItemCount  int
	PlaylistID spotifyPkg.ID
}

type Model struct {
	spotifyClient       *spotify.Client
	viewMode            ViewMode
	playlists           []spotifyPkg.SimplePlaylist
	searchResults       []spotifyPkg.FullTrack
	playlistTracks      []spotifyPkg.PlaylistTrack
	searchQuery         string
	playlistCursor      int
	trackCursor         int
	searchCursor        int
	libraryCursor       int
	selectedPlaylist    *spotifyPkg.SimplePlaylist
	selectedLibraryItem *LibraryCategory
	currentlyPlaying    *spotifyPkg.CurrentlyPlaying
	libraryCategories   []LibraryCategory
	savedTracks         []spotifyPkg.SavedTrack
	savedAlbums         []spotifyPkg.SavedAlbum
	followedArtists     []spotifyPkg.FullArtist
	recentlyPlayed      []*spotifyPkg.RecentlyPlayedItem
	width               int
	height              int
	isSearching         bool
	statusMessage       string
	err                 error
}
