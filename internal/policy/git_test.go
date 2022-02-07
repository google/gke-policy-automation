package policy

import (
	"io"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
)

type gitClientMock struct {
	cloneFn func(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error)
}

func (m gitClientMock) Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error) {
	return m.cloneFn(s, worktree, o)
}

type gitTreeWalkerResult struct {
	name  string
	entry *object.TreeEntry
}

type gitTreeWalkerMock struct {
	wkrResults []*gitTreeWalkerResult
	idx        int
}

func (m gitTreeWalkerMock) Close() {
}

func (m *gitTreeWalkerMock) Next() (name string, entry object.TreeEntry, err error) {
	if m.idx >= len(m.wkrResults) {
		return "", object.TreeEntry{}, io.EOF
	}
	result := m.wkrResults[m.idx]
	name = result.name
	entry = *result.entry
	m.idx++
	return
}

type gitTreeMock struct {
	TreeEntryFileFn func(e *object.TreeEntry) (*object.File, error)
}

func (m gitTreeMock) TreeEntryFile(e *object.TreeEntry) (*object.File, error) {
	return m.TreeEntryFileFn(e)
}

type encObjectMock struct {
	HashFn    func() plumbing.Hash
	TypeFn    func() plumbing.ObjectType
	SetTypeFn func(plumbing.ObjectType)
	SizeFn    func() int64
	SetSizeFn func(int64)
	ReaderFn  func() (io.ReadCloser, error)
	WriterFn  func() (io.WriteCloser, error)
}

func (o encObjectMock) Hash() plumbing.Hash {
	return o.HashFn()
}

func (o encObjectMock) Type() plumbing.ObjectType {
	return o.TypeFn()
}

func (o encObjectMock) SetType(t plumbing.ObjectType) {
	o.SetTypeFn(t)
}

func (o encObjectMock) Size() int64 {
	return o.SizeFn()
}

func (o encObjectMock) SetSize(s int64) {
	o.SetSizeFn(s)
}

func (o encObjectMock) Reader() (io.ReadCloser, error) {
	return o.ReaderFn()
}

func (o encObjectMock) Writer() (io.WriteCloser, error) {
	return o.WriterFn()
}

func TestNewGitPolicySource(t *testing.T) {
	repoURL := "https://test.com/repo"
	repoBranch := "main"
	policyDir := "dir"

	src := NewGitPolicySource(repoURL, repoBranch, policyDir)
	gitSrc, ok := src.(*GitPolicySource)
	if !ok {
		t.Errorf("Result of NewGitPolicySource is not *GitPolicySource")
	}
	if gitSrc.repoUrl != repoURL {
		t.Errorf("repoUrl = %s; want %s", gitSrc.repoUrl, repoURL)
	}
	if gitSrc.repoBranch != repoBranch {
		t.Errorf("repoBranch = %s; want %s", gitSrc.repoBranch, repoBranch)
	}
	if gitSrc.policyDir != policyDir {
		t.Errorf("policyDir = %s; want %s", gitSrc.policyDir, policyDir)
	}
	if gitSrc.policyFileExt != "rego" {
		t.Errorf("policyFileExt = %s; want %s", gitSrc.policyFileExt, "rego")
	}
}

func TestClone(t *testing.T) {
	var opts git.CloneOptions
	mock := &gitClientMock{
		cloneFn: func(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error) {
			opts = *o
			return &git.Repository{}, nil
		},
	}

	policySrc := &GitPolicySource{
		repoUrl:    "https://test.com/repository",
		repoBranch: "main",
		cli:        *mock,
	}

	_, err := policySrc.clone()
	if err != nil {
		t.Errorf("err is not nil; want nil")
	}
	if opts.URL != policySrc.repoUrl {
		t.Errorf("URL = %s; want %s", opts.URL, policySrc.repoUrl)
	}
	refName := plumbing.ReferenceName("refs/heads" + policySrc.repoBranch)
	if opts.ReferenceName != refName {
		t.Errorf("referenceName = %s; want %s", opts.ReferenceName, refName)
	}
	if opts.Depth != 1 {
		t.Errorf("depth = %d; want %d", opts.Depth, 1)
	}
	if !opts.SingleBranch {
		t.Errorf("singleBranch = false; want true")
	}
}

func TestGetRegoFileEntries(t *testing.T) {
	policySrc := &GitPolicySource{
		policyDir:     "policies",
		policyFileExt: "rego",
	}
	mock := &gitTreeWalkerMock{}
	mock.wkrResults = []*gitTreeWalkerResult{
		{
			name:  "policies",
			entry: &object.TreeEntry{Name: "policies", Mode: filemode.Dir},
		},
		{
			name:  "policies/policy_one.rego",
			entry: &object.TreeEntry{Name: "policy_one.rego", Mode: filemode.Regular},
		},
		{
			name:  "policies/policy_two.rego",
			entry: &object.TreeEntry{Name: "policy_two.rego", Mode: filemode.Regular},
		},
		{
			name:  "bogusDir",
			entry: &object.TreeEntry{Name: "bogusDir", Mode: filemode.Dir},
		},
		{
			name:  "bogusDir/bogus_policy.rego",
			entry: &object.TreeEntry{Name: "bogus_policy.rego", Mode: filemode.Regular},
		},
	}

	gitPolicyEntries, err := policySrc.getGitPolicyEntries(mock)
	if err != nil {
		t.Errorf("err is not nil; want nil")
	}
	if len(gitPolicyEntries) != 2 {
		t.Errorf("len(gitPolicyEntries) = %d; want %d", len(gitPolicyEntries), 2)
	}
	for i, entry := range gitPolicyEntries {
		if !strings.HasPrefix(entry.name, policySrc.policyDir+"/") {
			t.Errorf("gitPolicyEntries[%d] name = %s; want name with prefix %s", i, entry.name, policySrc.policyDir+"/")
		}
		if !strings.HasSuffix(entry.name, "."+policySrc.policyFileExt) {
			t.Errorf("gitPolicyEntries[%d] name = %s; want name with suffix %s", i, entry.name, "."+policySrc.policyFileExt)
		}
		if !strings.HasSuffix(entry.entry.Name, "."+policySrc.policyFileExt) {
			t.Errorf("gitPolicyEntries[%d] entry.Name = %s; want name with suffix %s", i, entry.entry.Name, "."+policySrc.policyFileExt)
		}
	}
}

func TestGetPolicyFile(t *testing.T) {
	hashString := "0d25de62c8d1e282b4d07ea74e6ca0912aa401fd"
	hash := plumbing.NewHash(hashString)
	entry := &gitPolicyEntry{
		name: "policies/policy_one.rego",
		entry: &object.TreeEntry{
			Name: "policy_one.rego",
			Mode: filemode.Regular,
			Hash: hash,
		},
	}
	content := "some policy file content"
	objMock := encObjectMock{
		TypeFn: func() plumbing.ObjectType {
			return plumbing.BlobObject
		},
		HashFn: func() plumbing.Hash {
			return hash
		},
		SizeFn: func() int64 {
			return 0
		},
		ReaderFn: func() (io.ReadCloser, error) {
			reader := strings.NewReader(content)
			return io.NopCloser(reader), nil
		},
	}
	blob, err := object.DecodeBlob(objMock)
	if err != nil {
		t.Fatalf("error when decoding mocked encoded object: %s", err)
	}

	mock := gitTreeMock{
		TreeEntryFileFn: func(e *object.TreeEntry) (*object.File, error) {
			return &object.File{
				Name: entry.name,
				Mode: entry.entry.Mode,
				Blob: *blob,
			}, nil
		},
	}
	file, err := entry.readPolicyFile(mock)
	if err != nil {
		t.Errorf("err is not nil; want nil")
	}
	if file.Hash != hashString {
		t.Errorf("hash = %s; want %s", file.Hash, hashString)
	}
	if file.Content != content {
		t.Errorf("content = %s; want %s", file.Content, content)
	}
	if file.FullName != entry.name {
		t.Errorf("fullName = %s; want %s", file.FullName, entry.name)
	}
	if file.Name != entry.entry.Name {
		t.Errorf("name = %s; want %s", file.Name, entry.entry.Name)
	}
}
