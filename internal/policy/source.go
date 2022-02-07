package policy

type PolicySource interface {
	GetPolicyFiles() ([]*PolicyFile, error)
}

type PolicyFile struct {
	Name     string
	FullName string
	Content  string
}
