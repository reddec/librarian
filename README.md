# Librarian

Generates in-memory, type-safe, thread-safe index of data stored from an arbitrary backend, keeping in memory only required
meta-data (indexed fields).

Go generate capable. Go modules capable.

## Install

`go get -u -v github.com/reddec/librarian/cmd/...`

## Usage

`librarian  -out some/storage.go some/package/types.go`


## Configuration

For base type, exported field in struct use: `index:"[NAME][,unique]"`:

* `NAME` - index name (also getter name). If not defined - field name prefixed by `By` will be used.
* `unique` - mark index as unique, by default index is not unique.

### Example

For file

**types.go**

```go
type User struct {
    Name string `index:",unique"`
    Role string `index:""`
    Year int
}
```

will generate (implementation omitted)

**generated.go**

```go
// constructors

// NewUserStorage
// NewUserStorageJSON
// NewUserStorageFilesJSON

// storage
type UserStorage struct {}
func (*UserStorage) Synchronize(context.Context) error {}
func (*UserStorage) Add(ctx context.Context, user User) error {}
func (*UserStorage) ByName(ctx context.Context, name string) (User, error) {}
func (*UserStorage) ByRole(ctx context.Context, role string) ([]User, error) {}
func (*UserStorage) RemoveByName(ctx context.Context, name string) error {}
func (*UserStorage) RemoveByRole(ctx context.Context, role string) error {}
func (*UserStorage) UpdateByName(ctx context.Context, user User) error {}
func (*UserStorage) UpsertByName(ctx context.Context, user User) error {}
```

checkout [examples](example)

Basic usage in code:


```go
func main() {
    directory := "."
    ctx := context.Background()
    store := NewUserStorageFilesJSON(directory)
    if err := store.Synchronize(ctx); err != nil { // synchronize internal state
        panic(err)
    }
    // store.ByName
    // store.Add
    // store.UpdateByName
    // ...
}
```