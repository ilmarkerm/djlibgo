package traktor

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// NML represents the root element of the Traktor collection.nml file
type NML struct {
	XMLName    xml.Name   `xml:"NML"`
	Version    string     `xml:"VERSION,attr"`
	Collection Collection `xml:"COLLECTION"`
	Playlists  Playlists  `xml:"PLAYLISTS"`
}

// Collection contains all tracks in the library
type Collection struct {
	Entries int     `xml:"ENTRIES,attr"`
	Tracks  []Entry `xml:"ENTRY"`
}

// Entry represents a single track in the collection
type Entry struct {
	Artist       string     `xml:"ARTIST,attr"`
	Title        string     `xml:"TITLE,attr"`
	AudioID      string     `xml:"AUDIO_ID,attr"`
	ModifiedDate string     `xml:"MODIFIED_DATE,attr"`
	ModifiedTime string     `xml:"MODIFIED_TIME,attr"`
	Location     Location   `xml:"LOCATION"`
	Album        Album      `xml:"ALBUM"`
	Info         Info       `xml:"INFO"`
	Tempo        Tempo      `xml:"TEMPO"`
	Loudness     Loudness   `xml:"LOUDNESS"`
	MusicalKey   MusicalKey `xml:"MUSICAL_KEY"`
	CuePoints    []CuePoint `xml:"CUE_V2"`
	LoopInfo     LoopInfo   `xml:"LOOPINFO"`
	PrimaryKey   string     `xml:"-"` // Computed field for playlist references
}

// Location contains file path information
type Location struct {
	Dir      string `xml:"DIR,attr"`
	File     string `xml:"FILE,attr"`
	Volume   string `xml:"VOLUME,attr"`
	VolumeID string `xml:"VOLUMEID,attr"`
}

// Album contains album metadata
type Album struct {
	Title    string `xml:"TITLE,attr"`
	Track    int    `xml:"TRACK,attr"`
	OfTracks int    `xml:"OF_TRACKS,attr"`
}

// Info contains additional track information
type Info struct {
	Bitrate       int     `xml:"BITRATE,attr"`
	Genre         string  `xml:"GENRE,attr"`
	Label         string  `xml:"LABEL,attr"`
	Comment       string  `xml:"COMMENT,attr"`
	Comment2      string  `xml:"COMMENT2,attr"`
	CoverArtID    string  `xml:"COVERARTID,attr"`
	Key           string  `xml:"KEY,attr"`
	PlayCount     int     `xml:"PLAYCOUNT,attr"`
	PlayTime      int     `xml:"PLAYTIME,attr"`
	PlayTimeFloat float64 `xml:"PLAYTIME_FLOAT,attr"`
	ImportDate    string  `xml:"IMPORT_DATE,attr"`
	LastPlayed    string  `xml:"LAST_PLAYED,attr"`
	Ranking       int     `xml:"RANKING,attr"`
	ReleaseDate   string  `xml:"RELEASE_DATE,attr"`
	Remixer       string  `xml:"REMIXER,attr"`
	Producer      string  `xml:"PRODUCER,attr"`
	Mix           string  `xml:"MIX,attr"`
	FileSize      int     `xml:"FILESIZE,attr"`
	Flags         int     `xml:"FLAGS,attr"`
}

// Tempo contains BPM information
type Tempo struct {
	Bpm        float64 `xml:"BPM,attr"`
	BpmQuality float64 `xml:"BPM_QUALITY,attr"`
}

// Loudness contains loudness analysis data
type Loudness struct {
	PeakDb      float64 `xml:"PEAK_DB,attr"`
	PerceivedDb float64 `xml:"PERCEIVED_DB,attr"`
	AnalyzedDb  float64 `xml:"ANALYZED_DB,attr"`
}

// MusicalKey contains key detection information
type MusicalKey struct {
	Value int `xml:"VALUE,attr"`
}

// CuePoint represents a cue point or loop marker
type CuePoint struct {
	Name    string  `xml:"NAME,attr"`
	Type    int     `xml:"TYPE,attr"`
	Start   float64 `xml:"START,attr"`
	Len     float64 `xml:"LEN,attr"`
	Repeats int     `xml:"REPEATS,attr"`
	HotCue  int     `xml:"HOTCUE,attr"`
}

// LoopInfo contains loop information
type LoopInfo struct {
	LoopStart float64 `xml:"LOOP_START,attr"`
	LoopEnd   float64 `xml:"LOOP_END,attr"`
}

// Playlists contains the playlist structure
type Playlists struct {
	Node Node `xml:"NODE"`
}

// Node represents a folder or playlist in the playlist tree
type Node struct {
	Type     string        `xml:"TYPE,attr"`
	Name     string        `xml:"NAME,attr"`
	Count    int           `xml:"COUNT,attr"`
	Subnodes []Node        `xml:"SUBNODES>NODE"`
	Playlist *PlaylistData `xml:"PLAYLIST"`
}

// PlaylistData contains the actual playlist entries
type PlaylistData struct {
	Entries int            `xml:"ENTRIES,attr"`
	Type    string         `xml:"TYPE,attr"`
	UUID    string         `xml:"UUID,attr"`
	Items   []PlaylistItem `xml:"ENTRY"`
}

// PlaylistItem represents a track reference in a playlist
type PlaylistItem struct {
	PrimaryKey PrimaryKey `xml:"PRIMARYKEY"`
}

// PrimaryKey is the unique identifier for a track
type PrimaryKey struct {
	Type string `xml:"TYPE,attr"`
	Key  string `xml:"KEY,attr"`
}

// Track represents a simplified track for external use
type Track struct {
	Artist      string
	Title       string
	Album       string
	Genre       string
	Label       string
	Comment     string
	Remixer     string
	Producer    string
	BPM         float64
	Key         string
	MusicalKey  int
	Rating      int
	PlayCount   int
	Duration    float64
	Bitrate     int
	FileSize    int
	FilePath    string
	FileName    string
	Volume      string
	ImportDate  string
	LastPlayed  string
	ReleaseDate string
	PeakDb      float64
	PerceivedDb float64
	CuePoints   []CuePoint
	PrimaryKey  string
}

// Playlist represents a simplified playlist for external use
type Playlist struct {
	Name      string
	Path      string
	TrackKeys []string
	Tracks    []*Track
}

// TraktorCollection holds the parsed collection data
type TraktorCollection struct {
	Version   string
	Tracks    []Track
	Playlists []Playlist
	trackMap  map[string]*Track
}

// IsAvailable checks if Traktor is installed and collection exists
func IsAvailable() bool {
	return collectionLocation() != ""
}

// CollectionLocation returns the path to the Traktor collection.nml file
func collectionLocation() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Check common Traktor versions
	versions := []string{
		"Traktor 4.4.1",
		"Traktor 4.4.0",
	}

	for _, version := range versions {
		path := filepath.Join(homeDir, "Documents", "Native Instruments", version, "collection.nml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// ParseCollection parses the Traktor collection.nml file from the default location
func ParseCollection() (*TraktorCollection, error) {
	location := collectionLocation()
	if location == "" {
		return nil, os.ErrNotExist
	}
	return ParseCollectionFromPath(location)
}

// ParseCollectionFromPath parses a Traktor collection.nml file from a specific path
func ParseCollectionFromPath(path string) (*TraktorCollection, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var nml NML
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&nml); err != nil {
		return nil, err
	}

	collection := &TraktorCollection{
		Version:  nml.Version,
		trackMap: make(map[string]*Track),
	}

	// Parse tracks
	collection.Tracks = make([]Track, 0, len(nml.Collection.Tracks))
	for _, entry := range nml.Collection.Tracks {
		track := convertEntryToTrack(entry)
		collection.Tracks = append(collection.Tracks, track)
		collection.trackMap[track.PrimaryKey] = &collection.Tracks[len(collection.Tracks)-1]
	}

	// Parse playlists
	collection.Playlists = extractPlaylists(nml.Playlists.Node, "", collection.trackMap)

	return collection, nil
}

// convertEntryToTrack converts an NML Entry to a simplified Track
func convertEntryToTrack(entry Entry) Track {
	// Build the primary key (used to reference tracks in playlists)
	primaryKey := buildPrimaryKey(entry.Location)

	// Build full file path
	filePath := buildFilePath(entry.Location)

	return Track{
		Artist:      entry.Artist,
		Title:       entry.Title,
		Album:       entry.Album.Title,
		Genre:       entry.Info.Genre,
		Label:       entry.Info.Label,
		Comment:     entry.Info.Comment,
		Remixer:     entry.Info.Remixer,
		Producer:    entry.Info.Producer,
		BPM:         entry.Tempo.Bpm,
		Key:         entry.Info.Key,
		MusicalKey:  entry.MusicalKey.Value,
		Rating:      entry.Info.Ranking,
		PlayCount:   entry.Info.PlayCount,
		Duration:    entry.Info.PlayTimeFloat,
		Bitrate:     entry.Info.Bitrate,
		FileSize:    entry.Info.FileSize,
		FilePath:    filePath,
		FileName:    entry.Location.File,
		Volume:      entry.Location.Volume,
		ImportDate:  entry.Info.ImportDate,
		LastPlayed:  entry.Info.LastPlayed,
		ReleaseDate: entry.Info.ReleaseDate,
		PeakDb:      entry.Loudness.PeakDb,
		PerceivedDb: entry.Loudness.PerceivedDb,
		CuePoints:   entry.CuePoints,
		PrimaryKey:  primaryKey,
	}
}

// buildPrimaryKey builds the primary key string used in playlist references
func buildPrimaryKey(loc Location) string {
	// Traktor uses VOLUME/:DIR/:FILE format for primary keys
	return loc.Volume + loc.Dir + loc.File
}

// buildFilePath converts Traktor's path format to a native file path
func buildFilePath(loc Location) string {
	// Traktor stores paths with /: as directory separators
	dir := strings.ReplaceAll(loc.Dir, "/:", string(os.PathSeparator))
	dir = strings.TrimPrefix(dir, string(os.PathSeparator))

	// On macOS, Volume is typically the volume name
	// We need to construct the full path
	if loc.Volume != "" {
		if loc.Volume == "Macintosh HD" || loc.Volume == ":" {
			return filepath.Join("/", dir, loc.File)
		}
		return filepath.Join("/Volumes", loc.Volume, dir, loc.File)
	}

	return filepath.Join(dir, loc.File)
}

// extractPlaylists recursively extracts playlists from the node tree
func extractPlaylists(node Node, parentPath string, trackMap map[string]*Track) []Playlist {
	var playlists []Playlist

	currentPath := parentPath
	if node.Name != "" && node.Name != "$ROOT" {
		if currentPath == "" {
			currentPath = node.Name
		} else {
			currentPath = currentPath + "/" + node.Name
		}
	}

	// If this node is a playlist (has playlist data)
	if node.Type == "PLAYLIST" && node.Playlist != nil {
		playlist := Playlist{
			Name:      node.Name,
			Path:      currentPath,
			TrackKeys: make([]string, 0, len(node.Playlist.Items)),
			Tracks:    make([]*Track, 0, len(node.Playlist.Items)),
		}

		for _, item := range node.Playlist.Items {
			key := item.PrimaryKey.Key
			playlist.TrackKeys = append(playlist.TrackKeys, key)

			// Look up the track in our map
			if track, exists := trackMap[key]; exists {
				playlist.Tracks = append(playlist.Tracks, track)
			}
		}

		playlists = append(playlists, playlist)
	}

	// Recursively process subnodes
	for _, subnode := range node.Subnodes {
		subPlaylists := extractPlaylists(subnode, currentPath, trackMap)
		playlists = append(playlists, subPlaylists...)
	}

	return playlists
}

// GetTrackByKey retrieves a track by its primary key
func (c *TraktorCollection) GetTrackByKey(key string) *Track {
	return c.trackMap[key]
}

// GetPlaylistByName finds a playlist by name
func (c *TraktorCollection) GetPlaylistByName(name string) *Playlist {
	for i := range c.Playlists {
		if c.Playlists[i].Name == name {
			return &c.Playlists[i]
		}
	}
	return nil
}

// GetPlaylistByPath finds a playlist by its full path
func (c *TraktorCollection) GetPlaylistByPath(path string) *Playlist {
	for i := range c.Playlists {
		if c.Playlists[i].Path == path {
			return &c.Playlists[i]
		}
	}
	return nil
}

// SearchTracks searches for tracks matching the query in artist, title, or album
func (c *TraktorCollection) SearchTracks(query string) []Track {
	query = strings.ToLower(query)
	var results []Track

	for _, track := range c.Tracks {
		if strings.Contains(strings.ToLower(track.Artist), query) ||
			strings.Contains(strings.ToLower(track.Title), query) ||
			strings.Contains(strings.ToLower(track.Album), query) {
			results = append(results, track)
		}
	}

	return results
}

// GetTracksByBPMRange returns tracks within a BPM range
func (c *TraktorCollection) GetTracksByBPMRange(minBPM, maxBPM float64) []Track {
	var results []Track

	for _, track := range c.Tracks {
		if track.BPM >= minBPM && track.BPM <= maxBPM {
			results = append(results, track)
		}
	}

	return results
}

// GetTracksByKey returns tracks with a specific musical key
func (c *TraktorCollection) GetTracksByKey(key string) []Track {
	key = strings.ToLower(key)
	var results []Track

	for _, track := range c.Tracks {
		if strings.ToLower(track.Key) == key {
			results = append(results, track)
		}
	}

	return results
}

// KeyValueToString converts the numeric musical key value to a string representation
func KeyValueToString(value int) string {
	// Traktor uses Open Key notation internally (0-23)
	// Even numbers are major, odd numbers are minor
	keys := []string{
		"1d", "1m", "2d", "2m", "3d", "3m",
		"4d", "4m", "5d", "5m", "6d", "6m",
		"7d", "7m", "8d", "8m", "9d", "9m",
		"10d", "10m", "11d", "11m", "12d", "12m",
	}

	if value >= 0 && value < len(keys) {
		return keys[value]
	}
	return ""
}

// FormatDuration formats duration in seconds to MM:SS format
func FormatDuration(seconds float64) string {
	totalSeconds := int(seconds)
	minutes := totalSeconds / 60
	secs := totalSeconds % 60
	return strconv.Itoa(minutes) + ":" + strconv.Itoa(secs)
}

// CuePointTypeToString converts cue point type to a human-readable string
func CuePointTypeToString(cueType int) string {
	switch cueType {
	case 0:
		return "Cue"
	case 1:
		return "Fade In"
	case 2:
		return "Fade Out"
	case 3:
		return "Load"
	case 4:
		return "Grid"
	case 5:
		return "Loop"
	default:
		return "Unknown"
	}
}
