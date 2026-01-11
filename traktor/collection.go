package traktor

import (
	"fmt"
	"sort"
)

var tc *TraktorCollection
var tcLoaded bool = false

// GetPlaylists returns a list of playlists.
func GetPlaylists() []Playlist {
	if !tcLoaded {
		tc, _ = ParseCollection()
		tcLoaded = true
	}
	return tc.Playlists
}

// GetSortedPlaylistNames returns a sorted list of playlist names.
func GetSortedPlaylistNames() []string {
	pl := GetPlaylists()
	names := make([]string, len(pl))
	for i, playlist := range pl {
		names[i] = playlist.Name
	}
	sort.Strings(names)
	return names
}

// LoadCollection loads the Traktor collection.
func LoadCollection() {
	if !tcLoaded {
		tc, _ = ParseCollection()
		tcLoaded = true
	}
	fmt.Printf("Traktor collection loaded. Number of playlists: %d\n", len(tc.Playlists))
}

func GetPlaylistByName(name string) *Playlist {
	if !tcLoaded {
		tc, _ = ParseCollection()
		tcLoaded = true
	}
	return tc.GetPlaylistByName(name)
}
