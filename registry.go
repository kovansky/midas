package strapi2hugo

type Registry map[string]string

type RegistryService interface {
	CreateStorage() error
	RemoveStorage() error
	Flush() error
	CreateEntry(id, filename string) error
	ReadEntry(id string) (string, error)
	UpdateEntry(id, newFilename string) error
	DeleteEntry(id string) error
}
