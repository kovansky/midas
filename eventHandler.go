package strapi2hugo

// ToDo: Will be used?

type EventHandler interface {
	Execute(payload Payload) (bool, error)
}
