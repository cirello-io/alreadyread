// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package web

import (
	"cirello.io/alreadyread/pkg/bookmarks"
	"sync"
)

// Ensure, that RepositoryMock does implement bookmarks.Repository.
// If this is not the case, regenerate this file with moq.
var _ bookmarks.Repository = &RepositoryMock{}

// RepositoryMock is a mock implementation of bookmarks.Repository.
//
//	func TestSomethingThatUsesRepository(t *testing.T) {
//
//		// make and configure a mocked bookmarks.Repository
//		mockedRepository := &RepositoryMock{
//			AllFunc: func() ([]*bookmarks.Bookmark, error) {
//				panic("mock out the All method")
//			},
//			BootstrapFunc: func() error {
//				panic("mock out the Bootstrap method")
//			},
//			DeadFunc: func() ([]*bookmarks.Bookmark, error) {
//				panic("mock out the Dead method")
//			},
//			DeleteByIDFunc: func(id int64) error {
//				panic("mock out the DeleteByID method")
//			},
//			DuplicatedFunc: func() ([]*bookmarks.Bookmark, error) {
//				panic("mock out the Duplicated method")
//			},
//			ExpiredFunc: func() ([]*bookmarks.Bookmark, error) {
//				panic("mock out the Expired method")
//			},
//			GetByIDFunc: func(id int64) (*bookmarks.Bookmark, error) {
//				panic("mock out the GetByID method")
//			},
//			InboxFunc: func() ([]*bookmarks.Bookmark, error) {
//				panic("mock out the Inbox method")
//			},
//			InsertFunc: func(bookmark *bookmarks.Bookmark) error {
//				panic("mock out the Insert method")
//			},
//			InvalidFunc: func() ([]*bookmarks.Bookmark, error) {
//				panic("mock out the Invalid method")
//			},
//			SearchFunc: func(term string) ([]*bookmarks.Bookmark, error) {
//				panic("mock out the Search method")
//			},
//			UpdateFunc: func(bookmark *bookmarks.Bookmark) error {
//				panic("mock out the Update method")
//			},
//		}
//
//		// use mockedRepository in code that requires bookmarks.Repository
//		// and then make assertions.
//
//	}
type RepositoryMock struct {
	// AllFunc mocks the All method.
	AllFunc func() ([]*bookmarks.Bookmark, error)

	// BootstrapFunc mocks the Bootstrap method.
	BootstrapFunc func() error

	// DeadFunc mocks the Dead method.
	DeadFunc func() ([]*bookmarks.Bookmark, error)

	// DeleteByIDFunc mocks the DeleteByID method.
	DeleteByIDFunc func(id int64) error

	// DuplicatedFunc mocks the Duplicated method.
	DuplicatedFunc func() ([]*bookmarks.Bookmark, error)

	// ExpiredFunc mocks the Expired method.
	ExpiredFunc func() ([]*bookmarks.Bookmark, error)

	// GetByIDFunc mocks the GetByID method.
	GetByIDFunc func(id int64) (*bookmarks.Bookmark, error)

	// InboxFunc mocks the Inbox method.
	InboxFunc func() ([]*bookmarks.Bookmark, error)

	// InsertFunc mocks the Insert method.
	InsertFunc func(bookmark *bookmarks.Bookmark) error

	// InvalidFunc mocks the Invalid method.
	InvalidFunc func() ([]*bookmarks.Bookmark, error)

	// SearchFunc mocks the Search method.
	SearchFunc func(term string) ([]*bookmarks.Bookmark, error)

	// UpdateFunc mocks the Update method.
	UpdateFunc func(bookmark *bookmarks.Bookmark) error

	// calls tracks calls to the methods.
	calls struct {
		// All holds details about calls to the All method.
		All []struct {
		}
		// Bootstrap holds details about calls to the Bootstrap method.
		Bootstrap []struct {
		}
		// Dead holds details about calls to the Dead method.
		Dead []struct {
		}
		// DeleteByID holds details about calls to the DeleteByID method.
		DeleteByID []struct {
			// ID is the id argument value.
			ID int64
		}
		// Duplicated holds details about calls to the Duplicated method.
		Duplicated []struct {
		}
		// Expired holds details about calls to the Expired method.
		Expired []struct {
		}
		// GetByID holds details about calls to the GetByID method.
		GetByID []struct {
			// ID is the id argument value.
			ID int64
		}
		// Inbox holds details about calls to the Inbox method.
		Inbox []struct {
		}
		// Insert holds details about calls to the Insert method.
		Insert []struct {
			// Bookmark is the bookmark argument value.
			Bookmark *bookmarks.Bookmark
		}
		// Invalid holds details about calls to the Invalid method.
		Invalid []struct {
		}
		// Search holds details about calls to the Search method.
		Search []struct {
			// Term is the term argument value.
			Term string
		}
		// Update holds details about calls to the Update method.
		Update []struct {
			// Bookmark is the bookmark argument value.
			Bookmark *bookmarks.Bookmark
		}
	}
	lockAll        sync.RWMutex
	lockBootstrap  sync.RWMutex
	lockDead       sync.RWMutex
	lockDeleteByID sync.RWMutex
	lockDuplicated sync.RWMutex
	lockExpired    sync.RWMutex
	lockGetByID    sync.RWMutex
	lockInbox      sync.RWMutex
	lockInsert     sync.RWMutex
	lockInvalid    sync.RWMutex
	lockSearch     sync.RWMutex
	lockUpdate     sync.RWMutex
}

// All calls AllFunc.
func (mock *RepositoryMock) All() ([]*bookmarks.Bookmark, error) {
	if mock.AllFunc == nil {
		panic("RepositoryMock.AllFunc: method is nil but Repository.All was just called")
	}
	callInfo := struct {
	}{}
	mock.lockAll.Lock()
	mock.calls.All = append(mock.calls.All, callInfo)
	mock.lockAll.Unlock()
	return mock.AllFunc()
}

// AllCalls gets all the calls that were made to All.
// Check the length with:
//
//	len(mockedRepository.AllCalls())
func (mock *RepositoryMock) AllCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockAll.RLock()
	calls = mock.calls.All
	mock.lockAll.RUnlock()
	return calls
}

// Bootstrap calls BootstrapFunc.
func (mock *RepositoryMock) Bootstrap() error {
	if mock.BootstrapFunc == nil {
		panic("RepositoryMock.BootstrapFunc: method is nil but Repository.Bootstrap was just called")
	}
	callInfo := struct {
	}{}
	mock.lockBootstrap.Lock()
	mock.calls.Bootstrap = append(mock.calls.Bootstrap, callInfo)
	mock.lockBootstrap.Unlock()
	return mock.BootstrapFunc()
}

// BootstrapCalls gets all the calls that were made to Bootstrap.
// Check the length with:
//
//	len(mockedRepository.BootstrapCalls())
func (mock *RepositoryMock) BootstrapCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockBootstrap.RLock()
	calls = mock.calls.Bootstrap
	mock.lockBootstrap.RUnlock()
	return calls
}

// Dead calls DeadFunc.
func (mock *RepositoryMock) Dead() ([]*bookmarks.Bookmark, error) {
	if mock.DeadFunc == nil {
		panic("RepositoryMock.DeadFunc: method is nil but Repository.Dead was just called")
	}
	callInfo := struct {
	}{}
	mock.lockDead.Lock()
	mock.calls.Dead = append(mock.calls.Dead, callInfo)
	mock.lockDead.Unlock()
	return mock.DeadFunc()
}

// DeadCalls gets all the calls that were made to Dead.
// Check the length with:
//
//	len(mockedRepository.DeadCalls())
func (mock *RepositoryMock) DeadCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockDead.RLock()
	calls = mock.calls.Dead
	mock.lockDead.RUnlock()
	return calls
}

// DeleteByID calls DeleteByIDFunc.
func (mock *RepositoryMock) DeleteByID(id int64) error {
	if mock.DeleteByIDFunc == nil {
		panic("RepositoryMock.DeleteByIDFunc: method is nil but Repository.DeleteByID was just called")
	}
	callInfo := struct {
		ID int64
	}{
		ID: id,
	}
	mock.lockDeleteByID.Lock()
	mock.calls.DeleteByID = append(mock.calls.DeleteByID, callInfo)
	mock.lockDeleteByID.Unlock()
	return mock.DeleteByIDFunc(id)
}

// DeleteByIDCalls gets all the calls that were made to DeleteByID.
// Check the length with:
//
//	len(mockedRepository.DeleteByIDCalls())
func (mock *RepositoryMock) DeleteByIDCalls() []struct {
	ID int64
} {
	var calls []struct {
		ID int64
	}
	mock.lockDeleteByID.RLock()
	calls = mock.calls.DeleteByID
	mock.lockDeleteByID.RUnlock()
	return calls
}

// Duplicated calls DuplicatedFunc.
func (mock *RepositoryMock) Duplicated() ([]*bookmarks.Bookmark, error) {
	if mock.DuplicatedFunc == nil {
		panic("RepositoryMock.DuplicatedFunc: method is nil but Repository.Duplicated was just called")
	}
	callInfo := struct {
	}{}
	mock.lockDuplicated.Lock()
	mock.calls.Duplicated = append(mock.calls.Duplicated, callInfo)
	mock.lockDuplicated.Unlock()
	return mock.DuplicatedFunc()
}

// DuplicatedCalls gets all the calls that were made to Duplicated.
// Check the length with:
//
//	len(mockedRepository.DuplicatedCalls())
func (mock *RepositoryMock) DuplicatedCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockDuplicated.RLock()
	calls = mock.calls.Duplicated
	mock.lockDuplicated.RUnlock()
	return calls
}

// Expired calls ExpiredFunc.
func (mock *RepositoryMock) Expired() ([]*bookmarks.Bookmark, error) {
	if mock.ExpiredFunc == nil {
		panic("RepositoryMock.ExpiredFunc: method is nil but Repository.Expired was just called")
	}
	callInfo := struct {
	}{}
	mock.lockExpired.Lock()
	mock.calls.Expired = append(mock.calls.Expired, callInfo)
	mock.lockExpired.Unlock()
	return mock.ExpiredFunc()
}

// ExpiredCalls gets all the calls that were made to Expired.
// Check the length with:
//
//	len(mockedRepository.ExpiredCalls())
func (mock *RepositoryMock) ExpiredCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockExpired.RLock()
	calls = mock.calls.Expired
	mock.lockExpired.RUnlock()
	return calls
}

// GetByID calls GetByIDFunc.
func (mock *RepositoryMock) GetByID(id int64) (*bookmarks.Bookmark, error) {
	if mock.GetByIDFunc == nil {
		panic("RepositoryMock.GetByIDFunc: method is nil but Repository.GetByID was just called")
	}
	callInfo := struct {
		ID int64
	}{
		ID: id,
	}
	mock.lockGetByID.Lock()
	mock.calls.GetByID = append(mock.calls.GetByID, callInfo)
	mock.lockGetByID.Unlock()
	return mock.GetByIDFunc(id)
}

// GetByIDCalls gets all the calls that were made to GetByID.
// Check the length with:
//
//	len(mockedRepository.GetByIDCalls())
func (mock *RepositoryMock) GetByIDCalls() []struct {
	ID int64
} {
	var calls []struct {
		ID int64
	}
	mock.lockGetByID.RLock()
	calls = mock.calls.GetByID
	mock.lockGetByID.RUnlock()
	return calls
}

// Inbox calls InboxFunc.
func (mock *RepositoryMock) Inbox() ([]*bookmarks.Bookmark, error) {
	if mock.InboxFunc == nil {
		panic("RepositoryMock.InboxFunc: method is nil but Repository.Inbox was just called")
	}
	callInfo := struct {
	}{}
	mock.lockInbox.Lock()
	mock.calls.Inbox = append(mock.calls.Inbox, callInfo)
	mock.lockInbox.Unlock()
	return mock.InboxFunc()
}

// InboxCalls gets all the calls that were made to Inbox.
// Check the length with:
//
//	len(mockedRepository.InboxCalls())
func (mock *RepositoryMock) InboxCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockInbox.RLock()
	calls = mock.calls.Inbox
	mock.lockInbox.RUnlock()
	return calls
}

// Insert calls InsertFunc.
func (mock *RepositoryMock) Insert(bookmark *bookmarks.Bookmark) error {
	if mock.InsertFunc == nil {
		panic("RepositoryMock.InsertFunc: method is nil but Repository.Insert was just called")
	}
	callInfo := struct {
		Bookmark *bookmarks.Bookmark
	}{
		Bookmark: bookmark,
	}
	mock.lockInsert.Lock()
	mock.calls.Insert = append(mock.calls.Insert, callInfo)
	mock.lockInsert.Unlock()
	return mock.InsertFunc(bookmark)
}

// InsertCalls gets all the calls that were made to Insert.
// Check the length with:
//
//	len(mockedRepository.InsertCalls())
func (mock *RepositoryMock) InsertCalls() []struct {
	Bookmark *bookmarks.Bookmark
} {
	var calls []struct {
		Bookmark *bookmarks.Bookmark
	}
	mock.lockInsert.RLock()
	calls = mock.calls.Insert
	mock.lockInsert.RUnlock()
	return calls
}

// Invalid calls InvalidFunc.
func (mock *RepositoryMock) Invalid() ([]*bookmarks.Bookmark, error) {
	if mock.InvalidFunc == nil {
		panic("RepositoryMock.InvalidFunc: method is nil but Repository.Invalid was just called")
	}
	callInfo := struct {
	}{}
	mock.lockInvalid.Lock()
	mock.calls.Invalid = append(mock.calls.Invalid, callInfo)
	mock.lockInvalid.Unlock()
	return mock.InvalidFunc()
}

// InvalidCalls gets all the calls that were made to Invalid.
// Check the length with:
//
//	len(mockedRepository.InvalidCalls())
func (mock *RepositoryMock) InvalidCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockInvalid.RLock()
	calls = mock.calls.Invalid
	mock.lockInvalid.RUnlock()
	return calls
}

// Search calls SearchFunc.
func (mock *RepositoryMock) Search(term string) ([]*bookmarks.Bookmark, error) {
	if mock.SearchFunc == nil {
		panic("RepositoryMock.SearchFunc: method is nil but Repository.Search was just called")
	}
	callInfo := struct {
		Term string
	}{
		Term: term,
	}
	mock.lockSearch.Lock()
	mock.calls.Search = append(mock.calls.Search, callInfo)
	mock.lockSearch.Unlock()
	return mock.SearchFunc(term)
}

// SearchCalls gets all the calls that were made to Search.
// Check the length with:
//
//	len(mockedRepository.SearchCalls())
func (mock *RepositoryMock) SearchCalls() []struct {
	Term string
} {
	var calls []struct {
		Term string
	}
	mock.lockSearch.RLock()
	calls = mock.calls.Search
	mock.lockSearch.RUnlock()
	return calls
}

// Update calls UpdateFunc.
func (mock *RepositoryMock) Update(bookmark *bookmarks.Bookmark) error {
	if mock.UpdateFunc == nil {
		panic("RepositoryMock.UpdateFunc: method is nil but Repository.Update was just called")
	}
	callInfo := struct {
		Bookmark *bookmarks.Bookmark
	}{
		Bookmark: bookmark,
	}
	mock.lockUpdate.Lock()
	mock.calls.Update = append(mock.calls.Update, callInfo)
	mock.lockUpdate.Unlock()
	return mock.UpdateFunc(bookmark)
}

// UpdateCalls gets all the calls that were made to Update.
// Check the length with:
//
//	len(mockedRepository.UpdateCalls())
func (mock *RepositoryMock) UpdateCalls() []struct {
	Bookmark *bookmarks.Bookmark
} {
	var calls []struct {
		Bookmark *bookmarks.Bookmark
	}
	mock.lockUpdate.RLock()
	calls = mock.calls.Update
	mock.lockUpdate.RUnlock()
	return calls
}
