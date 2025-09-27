package spotify

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

type Client struct {
	client *spotify.Client
	ctx    context.Context
}

func NewClient(spotifyClient *spotify.Client) *Client {
	return &Client{
		client: spotifyClient,
		ctx:    context.Background(),
	}
}

func (c *Client) GetPlaylists() ([]spotify.SimplePlaylist, error) {
	playlists, err := c.client.CurrentUsersPlaylists(c.ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}
	return playlists.Playlists, nil
}

func (c *Client) GetPlaylistByName(name string) (*spotify.SimplePlaylist, error) {
	// Search through user's playlists for the specified name
	playlists, err := c.client.CurrentUsersPlaylists(c.ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	for _, playlist := range playlists.Playlists {
		if playlist.Name == name {
			return &playlist, nil
		}
	}

	// If not found in first 50, continue searching
	for playlists.Next != "" {
		err = c.client.NextPage(c.ctx, playlists)
		if err != nil {
			return nil, err
		}

		for _, playlist := range playlists.Playlists {
			if playlist.Name == name {
				return &playlist, nil
			}
		}
	}

	return nil, nil
}

func (c *Client) GetPlaylistTracks(playlistID spotify.ID) ([]spotify.PlaylistTrack, error) {
	tracks, err := c.client.GetPlaylistTracks(c.ctx, playlistID, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	var allTracks []spotify.PlaylistTrack
	for page := 1; ; page++ {
		allTracks = append(allTracks, tracks.Tracks...)
		err = c.client.NextPage(c.ctx, tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return allTracks, nil
}

func (c *Client) SearchTracks(query string) ([]spotify.FullTrack, error) {
	results, err := c.client.Search(c.ctx, query, spotify.SearchTypeTrack, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	if results.Tracks == nil {
		return []spotify.FullTrack{}, nil
	}

	return results.Tracks.Tracks, nil
}

func (c *Client) PlayTrack(deviceID spotify.ID, trackURI spotify.URI) error {
	opts := &spotify.PlayOptions{
		DeviceID: &deviceID,
		URIs:     []spotify.URI{trackURI},
	}
	return c.client.PlayOpt(c.ctx, opts)
}

func (c *Client) Pause(opts *spotify.PlayOptions) error {
	return c.client.PauseOpt(c.ctx, opts)
}

func (c *Client) Resume(opts *spotify.PlayOptions) error {
	return c.client.PlayOpt(c.ctx, opts)
}

func (c *Client) Next() error {
	return c.client.Next(c.ctx)
}

func (c *Client) Previous() error {
	return c.client.Previous(c.ctx)
}

func (c *Client) GetCurrentlyPlaying() (*spotify.CurrentlyPlaying, error) {
	currently, err := c.client.PlayerCurrentlyPlaying(c.ctx)
	if err != nil {
		return nil, err
	}
	return currently, nil
}

func (c *Client) GetDevices() ([]spotify.PlayerDevice, error) {
	devices, err := c.client.PlayerDevices(c.ctx)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (c *Client) SetVolume(percent int) error {
	return c.client.Volume(c.ctx, percent)
}

func (c *Client) GetSavedTracks() ([]spotify.SavedTrack, error) {
	tracks, err := c.client.CurrentUsersTracks(c.ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	var allTracks []spotify.SavedTrack
	for page := 1; ; page++ {
		allTracks = append(allTracks, tracks.Tracks...)
		err = c.client.NextPage(c.ctx, tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return allTracks, nil
}

// GetSavedTracksPage fetches a specific page of saved tracks
func (c *Client) GetSavedTracksPage(limit, offset int) ([]spotify.SavedTrack, int, error) {
	tracks, err := c.client.CurrentUsersTracks(c.ctx, spotify.Limit(limit), spotify.Offset(offset))
	if err != nil {
		return nil, 0, err
	}
	return tracks.Tracks, int(tracks.Total), nil
}

// GetSavedTracksCount returns just the total count of saved tracks
func (c *Client) GetSavedTracksCount() (int, error) {
	tracks, err := c.client.CurrentUsersTracks(c.ctx, spotify.Limit(1))
	if err != nil {
		return 0, err
	}
	return int(tracks.Total), nil
}

func (c *Client) GetSavedAlbums() ([]spotify.SavedAlbum, error) {
	albums, err := c.client.CurrentUsersAlbums(c.ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	var allAlbums []spotify.SavedAlbum
	for page := 1; ; page++ {
		allAlbums = append(allAlbums, albums.Albums...)
		err = c.client.NextPage(c.ctx, albums)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return allAlbums, nil
}

// GetSavedAlbumsPage fetches a specific page of saved albums
func (c *Client) GetSavedAlbumsPage(limit, offset int) ([]spotify.SavedAlbum, int, error) {
	albums, err := c.client.CurrentUsersAlbums(c.ctx, spotify.Limit(limit), spotify.Offset(offset))
	if err != nil {
		return nil, 0, err
	}
	return albums.Albums, int(albums.Total), nil
}

// GetSavedAlbumsCount returns just the total count of saved albums
func (c *Client) GetSavedAlbumsCount() (int, error) {
	albums, err := c.client.CurrentUsersAlbums(c.ctx, spotify.Limit(1))
	if err != nil {
		return 0, err
	}
	return int(albums.Total), nil
}

func (c *Client) GetAlbumTracks(albumID spotify.ID) ([]spotify.SimpleTrack, error) {
	tracks, err := c.client.GetAlbumTracks(c.ctx, albumID, spotify.Limit(50))
	if err != nil {
		return nil, err
	}

	var allTracks []spotify.SimpleTrack
	for page := 1; ; page++ {
		allTracks = append(allTracks, tracks.Tracks...)
		err = c.client.NextPage(c.ctx, tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return allTracks, nil
}

func (c *Client) GetFollowedArtists() ([]spotify.FullArtist, error) {
	artists, err := c.client.CurrentUsersFollowedArtists(c.ctx, spotify.Limit(50))
	if err != nil {
		return nil, err
	}
	return artists.Artists, nil
}

func (c *Client) GetRecentlyPlayed() ([]spotify.RecentlyPlayedItem, error) {
	recent, err := c.client.PlayerRecentlyPlayedOpt(c.ctx, &spotify.RecentlyPlayedOptions{
		Limit: 50,
	})
	if err != nil {
		return nil, err
	}
	return recent, nil
}

func (c *Client) GetActiveDevice() (spotify.ID, error) {
	devices, err := c.client.PlayerDevices(c.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get devices: %w", err)
	}

	if len(devices) == 0 {
		return "", fmt.Errorf("no active devices found. Please open Spotify on a device")
	}

	// First try to find an actively playing device
	for _, device := range devices {
		if device.Active {
			return device.ID, nil
		}
	}

	// If no active device, use the first available device
	return devices[0].ID, nil
}

func (c *Client) FormatTrackInfo(track *spotify.FullTrack) string {
	artists := ""
	for i, artist := range track.Artists {
		if i > 0 {
			artists += ", "
		}
		artists += artist.Name
	}
	duration := fmt.Sprintf("%d:%02d", track.Duration/60000, (track.Duration/1000)%60)
	return fmt.Sprintf("%s - %s (%s)", track.Name, artists, duration)
}
