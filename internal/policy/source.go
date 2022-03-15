package policy

type PolicySource interface {
	GetPolicyFiles() ([]*PolicyFile, error)
	String() string
}

type PolicyFile struct {
	Name     string
	FullName string
	Content  string
}
