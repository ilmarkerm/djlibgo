package windows

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ilmarkerm/djlibgo/traktor"
)

// TreeNodeUID represents a unique identifier for tree nodes
type TreeNodeUID string

// AppState holds the application state
type AppState struct {
	selectedPath string
	fileTable    *widget.Table
	files        []FileItem
	tree         *widget.Tree
	treeData     map[TreeNodeUID][]TreeNodeUID
}

// FileItem represents a file in the file list
type FileItem struct {
	Artist string
	Title  string
	Label  string
	Year   int
	Path   string
	Size   int64
}

// NewAppState creates a new application state
func NewAppState() *AppState {
	return &AppState{
		treeData: make(map[TreeNodeUID][]TreeNodeUID),
		//treePaths: make(map[TreeNodeUID]string),
		files: []FileItem{},
	}
}

// getMusicTreeRoot returns the Music tree root entries
func (s *AppState) getMusicTreeRoot() []TreeNodeUID {
	var roots []TreeNodeUID

	// Add special entries accounts
	for _, name := range []string{"plex", "bandcamp", "traktor", "rekordbox"} {
		if name == "traktor" && traktor.IsAvailable() {
			uid := TreeNodeUID(traktor.Prefix)
			//s.treePaths[uid] = traktorPrefix
			roots = append(roots, uid)
		} else {
			uid := TreeNodeUID(fmt.Sprintf("special://%s", name))
			//s.treePaths[uid] = fmt.Sprintf("special://%s", name)
			roots = append(roots, uid)
		}
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		uid := TreeNodeUID(homeDir)
		roots = append(roots, uid)
	}

	s.treeData[""] = roots
	return roots
}

// loadChildren loads children for a given tree node
func (s *AppState) loadChildren(uid TreeNodeUID) []TreeNodeUID {
	if children, exists := s.treeData[uid]; exists {
		return children
	}

	path := string(uid)
	if path == "" {
		return nil
	}

	var children []TreeNodeUID

	if strings.HasPrefix(path, traktor.Prefix) && traktor.IsAvailable() {
		if path == traktor.Prefix {
			children = append(children, TreeNodeUID(traktor.PlaylistPrefix))
			children = append(children, TreeNodeUID(traktor.CollectionPrefix))
		} else if path == traktor.PlaylistPrefix {
			for _, entry := range traktor.GetSortedPlaylistNames() {
				if strings.HasPrefix(entry, "_") {
					continue
				}
				children = append(children, TreeNodeUID(fmt.Sprintf("%s/%s", traktor.PlaylistPrefix, entry)))
			}
		}
	} else {
		// File handling
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		var dirNames []string

		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				dirNames = append(dirNames, entry.Name())
			}
		}

		sort.Strings(dirNames)

		for _, name := range dirNames {
			childPath := filepath.Join(path, name)
			childUID := TreeNodeUID(childPath)
			children = append(children, childUID)
		}
	}

	s.treeData[uid] = children
	return children
}

// getNodeLabel returns the display label for a tree node
func (s *AppState) getNodeLabel(uid TreeNodeUID) string {
	path := string(uid)
	if path == "" {
		return path
	}

	// Handle special root nodes
	if strings.HasPrefix(path, traktor.Prefix) {
		switch path {
		case traktor.Prefix:
			return "Traktor"
		case traktor.PlaylistPrefix:
			return "Playlists"
		case traktor.CollectionPrefix:
			return "Collection"
		default:
			parts := strings.Split(path, "/")
			return parts[len(parts)-1]
		}
	} else if strings.HasPrefix(path, "special://") {
		return strings.TrimPrefix(path, "special://")
	} else {
		switch uid {
		case "root":
			return "/"
		default:
			return filepath.Base(path)
		}
	}
}

// loadFilesForPath loads files for the given directory path
func (s *AppState) loadFilesForPath(dirPath string) {
	s.files = []FileItem{}

	if dirPath == "" {
		if s.fileTable != nil {
			s.fileTable.Refresh()
		}
		return
	}

	if strings.HasPrefix(dirPath, traktor.PlaylistPrefix) {
		parts := strings.Split(dirPath, "/")
		pl := traktor.GetPlaylistByName(parts[len(parts)-1])
		if pl != nil && pl.Tracks != nil {
			for _, track := range pl.Tracks {
				item := FileItem{
					Artist: track.Artist,
					Title:  track.Title,
					Label:  track.Label,
					Path:   track.FilePath,
					Size:   int64(track.FileSize),
				}
				s.files = append(s.files, item)
			}
		}
	} else {
		// Load filesystem files
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") || entry.IsDir() {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			item := FileItem{
				Artist: entry.Name(),
				Title:  "ttt",
				Label:  "lll",
				Path:   filepath.Join(dirPath, entry.Name()),
				Size:   info.Size(),
			}
			s.files = append(s.files, item)
		}
	}

	// Sort: directories first, then files, both alphabetically
	sort.Slice(s.files, func(i, j int) bool {
		return strings.ToLower(s.files[i].Path) < strings.ToLower(s.files[j].Path)
	})

	if s.fileTable != nil {
		s.fileTable.Refresh()
	}
}

// formatSize formats file size in human-readable format
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return formatFloat(float64(size)/float64(GB)) + " GB"
	case size >= MB:
		return formatFloat(float64(size)/float64(MB)) + " MB"
	case size >= KB:
		return formatFloat(float64(size)/float64(KB)) + " KB"
	default:
		return formatInt(size) + " B"
	}
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return formatInt(int64(f))
	}
	// Simple formatting without fmt
	intPart := int64(f)
	decPart := int64((f - float64(intPart)) * 10)
	if decPart < 0 {
		decPart = -decPart
	}
	return formatInt(intPart) + "." + formatInt(decPart)
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}

func MainWindow() {
	myApp := app.New()
	myApp.Settings().SetTheme(&myTheme{})

	window := myApp.NewWindow("DJ Library")
	window.Resize(fyne.NewSize(1200, 800))

	state := NewAppState()
	state.getMusicTreeRoot()

	// Create the directory tree
	tree := widget.NewTree(
		// ChildUIDs - returns children for a node
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			children := state.loadChildren(TreeNodeUID(uid))
			result := make([]widget.TreeNodeID, len(children))
			for i, c := range children {
				result[i] = widget.TreeNodeID(c)
			}
			return result
		},
		// IsBranch - determines if a node can have children
		func(uid widget.TreeNodeID) bool {
			path := string(uid)
			if path == "" {
				return uid == ""
			}
			if path == traktor.Prefix && traktor.IsAvailable() {
				return true
			}
			if path == traktor.PlaylistPrefix {
				return true
			}
			info, err := os.Stat(path)
			if err != nil {
				return false
			}
			return info.IsDir()
		},
		// CreateNode - creates a new tree node widget
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Directory")
		},
		// UpdateNode - updates a tree node with data
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			label := node.(*widget.Label)
			label.SetText(state.getNodeLabel(TreeNodeUID(uid)))
		},
	)

	tree.OnSelected = func(uid widget.TreeNodeID) {
		path := string(uid)
		state.selectedPath = path
		state.loadFilesForPath(path)
	}

	state.tree = tree

	// Column headers for the table
	columnHeaders := []string{"Artist", "Title", "Label", "Size"}

	// Create the file table
	fileTable := widget.NewTableWithHeaders(
		// Length - returns rows and columns
		func() (int, int) {
			return len(state.files), len(columnHeaders)
		},
		// CreateCell - creates a cell widget
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		// UpdateCell - updates a cell with data
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row >= len(state.files) {
				return
			}

			file := state.files[id.Row]
			switch id.Col {
			case 0:
				label.SetText(file.Artist)
			case 1:
				label.SetText(file.Title)
			case 2:
				label.SetText(file.Label)
			case 3:
				label.SetText(formatSize(file.Size))
				label.Alignment = fyne.TextAlignTrailing
			}
		},
	)

	// Set column headers
	fileTable.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("")
	}
	fileTable.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		label := cell.(*widget.Label)
		if id.Row == -1 && id.Col >= 0 && id.Col < len(columnHeaders) {
			label.SetText(columnHeaders[id.Col])
			label.TextStyle = fyne.TextStyle{Bold: true}
		}
	}

	// Set column widths
	fileTable.SetColumnWidth(0, 200) // Artist
	fileTable.SetColumnWidth(1, 250) // Title
	fileTable.SetColumnWidth(2, 150) // Label
	fileTable.SetColumnWidth(3, 100) // Size

	fileTable.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(state.files) {
			// File selected - could be used for future functionality
			_ = state.files[id.Row]
		}
	}

	state.fileTable = fileTable

	// Create buttons for the middle panel
	saveButton := widget.NewButton("Load Traktor collection", func() {
		traktor.LoadCollection()
	})
	saveButton.Importance = widget.HighImportance

	cancelButton := widget.NewButton("Cancel", func() {
		// Cancel functionality to be implemented
	})

	// Layout the panels
	// Left panel: Tree view with scroll
	leftPanel := container.NewBorder(
		widget.NewLabel("Music locations"),
		nil, nil, nil,
		container.NewScroll(tree),
	)

	// Top panel: Empty for now
	topPanel := container.NewBorder(
		widget.NewLabel("Details"),
		nil, nil, nil,
		widget.NewLabel(""), // Empty content
	)

	// Middle panel: Buttons
	buttonContainer := container.NewHBox(
		saveButton,
		cancelButton,
	)
	middlePanel := container.NewCenter(buttonContainer)

	// Bottom panel: File table
	bottomPanel := container.NewBorder(
		widget.NewLabel("Tracks"),
		nil, nil, nil,
		fileTable,
	)

	// Right panels stacked vertically with splits
	// Top empty panel (small), middle with buttons (small), bottom with file list (large)
	topAndMiddle := container.NewVSplit(topPanel, middlePanel)
	topAndMiddle.SetOffset(0.7) // Top panel takes 70% of top section

	rightPanels := container.NewVSplit(topAndMiddle, bottomPanel)
	rightPanels.SetOffset(0.3) // Top section takes 30%, file list takes 70%

	// Main layout: Left tree panel and right panels
	mainSplit := container.NewHSplit(leftPanel, rightPanels)
	mainSplit.SetOffset(0.25) // Left panel takes 25% of width

	// Create toolbar with refresh and save icons
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			// To be implemented
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			// Save functionality - to be implemented
		}),
	)

	// Main content with toolbar at top
	mainContent := container.NewBorder(toolbar, nil, nil, nil, mainSplit)

	window.SetContent(mainContent)
	window.ShowAndRun()
}
