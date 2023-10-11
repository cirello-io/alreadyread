package bookmarks

type Repository interface {
	// All returns all known bookmarks.
	All() ([]*Bookmark, error)

	// Bootstrap creates table if missing.
	Bootstrap() error

	// Delete one bookmark.
	Delete(*Bookmark) error

	// Expired return all valid but expired bookmarks.
	Expired() ([]*Bookmark, error)

	// GetByID loads one bookmark.
	GetByID(id int64) (*Bookmark, error)

	// Insert one bookmark.
	Insert(*Bookmark) (*Bookmark, error)

	// Invalid return all invalid bookmarks.
	Invalid() ([]*Bookmark, error)

	// Update one bookmark.
	Update(*Bookmark) error
}
