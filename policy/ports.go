package policy

type Manager interface {
	CheckPermission(sub string, obj string, act string) (bool, error)
}
