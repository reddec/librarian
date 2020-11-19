// Code generated by librarian. DO NOT EDIT.
package example

import (
	"context"
	"encoding/json"
	"fmt"
	librarian "github.com/reddec/librarian"
	"sync"
)

// NewUserStorage creates indexed storage for User.
func NewUserStorage(backend librarian.Storage, encoder func(item User) ([]byte, error), decoder func(data []byte, item *User) error) *UserStorage {
	return &UserStorage{backend: backend, encoder: encoder, decoder: decoder}
}

// NewUserStorageJSON create new indexed storage for User with custom storage backend, with data encoded in JSON.
func NewUserStorageJSON(backend librarian.Storage) *UserStorage {
	return NewUserStorage(backend, func(item User) ([]byte, error) {
		return json.Marshal(item)
	}, func(data []byte, item *User) error {
		return json.Unmarshal(data, item)
	})
}

// NewUserStorageFilesJSON creates new indexed storage for User with files in directory as backend storage, with data encoded in JSON.
func NewUserStorageFilesJSON(directory string) *UserStorage {
	return NewUserStorage(librarian.Directory(directory), func(item User) ([]byte, error) {
		return json.Marshal(item)
	}, func(data []byte, item *User) error {
		return json.Unmarshal(data, item)
	})
}

// Indexed storage for user.
type UserStorage struct {
	backend                  librarian.Storage
	encoder                  func(item User) ([]byte, error)
	decoder                  func(data []byte, item *User) error
	lock                     sync.RWMutex
	meta                     map[string]metaUserStorage // minimal meta information for indexes
	indexByName              map[string]string
	indexByRole              map[string]map[string]bool
	indexBySocialSecurityNum map[string]string
}

/*
Synchronize backend storage and index.
Will iterate over all items one-by-one and add to internal index indexed fields (not object itself).
Will block storage only at the moment when all items processed.
In case of error - internal index will not be changed.
*/
func (storage *UserStorage) Synchronize(ctx context.Context) error {
	live := NewUserStorage(storage.backend, storage.encoder, storage.decoder)
	err := storage.backend.Iterate(ctx, func(id string, data []byte) error {
		var item User
		err := storage.decoder(data, &item)
		if err != nil {
			return fmt.Errorf("decode User#%s data: %w", id, err)
		}
		live.indexItem(item, id)
		return nil
	})
	if err != nil {
		return fmt.Errorf("synchronize user: %w", err)
	}
	storage.lock.Lock()
	defer storage.lock.Unlock()

	storage.meta = live.meta
	storage.indexByName = live.indexByName
	storage.indexByRole = live.indexByRole
	storage.indexBySocialSecurityNum = live.indexBySocialSecurityNum
	return nil
}

/*
All known objects without filtration.
Returning slice is not sorted and order is not stable.
Use with caution! All objects will be stored in memory and may cause high GC usage as well as high memory consumption.
*/
func (storage *UserStorage) All(ctx context.Context) ([]User, error) {
	storage.lock.RLock()
	defer storage.lock.RUnlock()

	var ans = make([]User, 0, len(storage.meta))
	for id := range storage.meta {
		item, err := storage.get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("list all user: %w", err)
		}
		ans = append(ans, item)
	}
	return ans, nil
}

/*
ByName returns single User object filtered by name.
If nothing found - error will be returned together with empty object.
*/
func (storage *UserStorage) ByName(ctx context.Context, name string) (User, error) {
	storage.lock.RLock()
	defer storage.lock.RUnlock()

	id, found := storage.indexByName[name]
	if !found {
		return User{}, fmt.Errorf("user not found by name %v", name)
	}
	item, err := storage.get(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user by name %v: %w", name, err)
	}
	return item, nil
}

/*
ByRole returns multiple User objects filtered by role.
Returning slice is not sorted and order is not stable.
If nothing found - nil slice will be returned.
*/
func (storage *UserStorage) ByRole(ctx context.Context, role string) ([]User, error) {
	storage.lock.RLock()
	defer storage.lock.RUnlock()

	var result []User
	for id := range storage.indexByRole[role] {
		item, err := storage.get(ctx, id)
		if err != nil {
			return result, fmt.Errorf("list user by role %s: %w", role, err)
		}
		result = append(result, item)
	}
	return result, nil
}

/*
BySocialSecurityNum returns single User object filtered by ssn.
If nothing found - error will be returned together with empty object.
*/
func (storage *UserStorage) BySocialSecurityNum(ctx context.Context, ssn string) (User, error) {
	storage.lock.RLock()
	defer storage.lock.RUnlock()

	id, found := storage.indexBySocialSecurityNum[ssn]
	if !found {
		return User{}, fmt.Errorf("user not found by ssn %v", ssn)
	}
	item, err := storage.get(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user by ssn %v: %w", ssn, err)
	}
	return item, nil
}

/*
RemoveByName removes single User object by name.
If nothing found - operation will be ignored without error.
*/
func (storage *UserStorage) RemoveByName(ctx context.Context, name string) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	id, found := storage.indexByName[name]
	if !found {
		return nil
	}
	err := storage.remove(ctx, id)
	if err != nil {
		return fmt.Errorf("remove user by name %v: %w", name, err)
	}
	return nil
}

/*
RemoveByRole removes multiple User objects filtered by role.
If nothing found - operation will be ignored without error.
*/
func (storage *UserStorage) RemoveByRole(ctx context.Context, role string) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	var ids []string
	for id := range storage.indexByRole[role] {
		ids = append(ids, id)
	}
	for _, id := range ids {
		err := storage.remove(ctx, id)
		if err != nil {
			return fmt.Errorf("remove user by role %v: %w", role, err)
		}
	}
	return nil
}

/*
RemoveBySocialSecurityNum removes single User object by ssn.
If nothing found - operation will be ignored without error.
*/
func (storage *UserStorage) RemoveBySocialSecurityNum(ctx context.Context, ssn string) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	id, found := storage.indexBySocialSecurityNum[ssn]
	if !found {
		return nil
	}
	err := storage.remove(ctx, id)
	if err != nil {
		return fmt.Errorf("remove user by ssn %v: %w", ssn, err)
	}
	return nil
}

/*
UpdateByName updates single User object by name.
If object not exists - error will be returned.
*/
func (storage *UserStorage) UpdateByName(ctx context.Context, user User) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	id, exists := storage.indexByName[user.Name]
	if !exists {
		return fmt.Errorf("user with name %v doesn't exist", user.Name)
	}
	err := storage.update(ctx, id, user)
	if err != nil {
		return fmt.Errorf("update user with name %v: %w", user.Name, err)
	}
	return nil
}

/*
UpdateBySocialSecurityNum updates single User object by ssn.
If object not exists - error will be returned.
*/
func (storage *UserStorage) UpdateBySocialSecurityNum(ctx context.Context, user User) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	id, exists := storage.indexBySocialSecurityNum[user.SSN]
	if !exists {
		return fmt.Errorf("user with ssn %v doesn't exist", user.SSN)
	}
	err := storage.update(ctx, id, user)
	if err != nil {
		return fmt.Errorf("update user with ssn %v: %w", user.SSN, err)
	}
	return nil
}

/*
UpsertByName updates or creates single User object by name.
If object not exists - new object will be created.
*/
func (storage *UserStorage) UpsertByName(ctx context.Context, user User) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	id, exists := storage.indexByName[user.Name]
	if !exists {
		return storage.add(ctx, user)
	}
	err := storage.update(ctx, id, user)
	if err != nil {
		return fmt.Errorf("update user with name %v: %w", user.Name, err)
	}
	return nil
}

/*
UpsertBySocialSecurityNum updates or creates single User object by ssn.
If object not exists - new object will be created.
*/
func (storage *UserStorage) UpsertBySocialSecurityNum(ctx context.Context, user User) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	id, exists := storage.indexBySocialSecurityNum[user.SSN]
	if !exists {
		return storage.add(ctx, user)
	}
	err := storage.update(ctx, id, user)
	if err != nil {
		return fmt.Errorf("update user with ssn %v: %w", user.SSN, err)
	}
	return nil
}

/*
Add new user to the storage and index.
If some of unique fields already exists - error will be returned (unique constraint violation).
*/
func (storage *UserStorage) Add(ctx context.Context, user User) error {
	storage.lock.Lock()
	defer storage.lock.Unlock()

	err := storage.unique(user)
	if err != nil {
		return fmt.Errorf("user unique constraint violation: %w", err)
	}
	return storage.add(ctx, user)
}

func (storage *UserStorage) update(ctx context.Context, id string, user User) error {
	data, err := storage.encoder(user)
	if err != nil {
		return fmt.Errorf("encode user %s: %w", id, err)
	}
	err = storage.backend.Update(ctx, id, data)
	if err != nil {
		return fmt.Errorf("update user %s: %w", id, err)
	}
	storage.removeItemFromIndex(id)
	storage.indexItem(user, id)
	return nil
}

func (storage *UserStorage) indexItem(value User, id string) {
	storage.indexByName[value.Name] = id

	indexByRole, indexExists := storage.indexByRole[value.Role]
	if !indexExists {
		indexByRole = make(map[string]bool)
		storage.indexByRole[value.Role] = indexByRole
	}
	storage.indexByRole[value.Role][id] = true

	storage.indexBySocialSecurityNum[value.SSN] = id

	storage.meta[id] = metaUserStorage{Name: value.Name, Role: value.Role, SSN: value.SSN}
}

func (storage *UserStorage) removeItemFromIndex(id string) {
	value, exists := storage.meta[id]
	if !exists {
		return
	}
	delete(storage.meta, id)
	// remove from index by name
	delete(storage.indexByName, value.Name)

	// remove from index by role
	indexByRole := storage.indexByRole[value.Role]
	delete(indexByRole, id)
	if len(indexByRole) == 0 {
		delete(storage.indexByRole, value.Role)
	}

	// remove from index by ssn
	delete(storage.indexBySocialSecurityNum, value.SSN)
}

func (storage *UserStorage) remove(ctx context.Context, id string) error {
	err := storage.backend.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("delete User#%s: %w", id, err)
	}
	storage.removeItemFromIndex(id)
	return nil
}

func (storage *UserStorage) get(ctx context.Context, id string) (User, error) {
	data, err := storage.backend.Get(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("fetch User#%s data: %w", id, err)
	}
	var result User
	err = storage.decoder(data, &result)
	if err != nil {
		return User{}, fmt.Errorf("decode User#%s fetched data: %w", id, err)
	}
	return result, nil
}

func (storage *UserStorage) unique(value User) error {
	if id, exists := storage.indexByName[value.Name]; exists {
		return fmt.Errorf("user with name %s already exists as %s", value.Name, id)
	}

	if id, exists := storage.indexBySocialSecurityNum[value.SSN]; exists {
		return fmt.Errorf("user with ssn %s already exists as %s", value.SSN, id)
	}

	return nil
}

func (storage *UserStorage) add(ctx context.Context, user User) error {
	data, err := storage.encoder(user)
	if err != nil {
		return fmt.Errorf("encode user: %w", err)
	}
	id, err := storage.backend.Create(ctx, data)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	storage.indexItem(user, id)
	return nil
}

type metaUserStorage struct {
	Name string
	Role string
	SSN  string
}
