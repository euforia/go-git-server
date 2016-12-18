package repository

// Repository represents a single repo.
type Repository struct {
	ID   string                `json:"id"`
	Refs *RepositoryReferences `json:"refs"`
	// sha256 hash of the binary data from the store.  This is stored upon
	// retreival and used as the previous hash when trying to update.
	hash [32]byte
}

// NewRepository instantiates an empty repo.
func NewRepository(id string) *Repository {
	return &Repository{ID: id, Refs: NewRepositoryReferences()}
}

// Hash of the repo as seen by the datastore.  This is used as the previous hash
// to issue any updates back to the datastore.
func (repo *Repository) Hash() [32]byte {
	return repo.hash
}

func (repo *Repository) String() string {
	return repo.ID
}
