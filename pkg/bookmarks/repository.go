package bookmarks

//go:generate moq -out repository_mocks_test.go . Repository
type Repository interface {
	// All returns all known bookmarks.
	All() ([]*Bookmark, error)

	// Bootstrap creates table if missing.
	Bootstrap() error

	// DeleteByID excludes the bookmark from the repository.
	DeleteByID(id int64) error

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
