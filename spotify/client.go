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
