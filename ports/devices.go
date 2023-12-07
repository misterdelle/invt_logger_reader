package ports

type Device interface {
	Name() string
	Query() (map[string]interface{}, error)
	QueryStation() (map[string]interface{}, error)
}
