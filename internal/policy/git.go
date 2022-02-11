package policy

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
)

type CloneFn func(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error)

type GitPolicySource struct {
	repoUrl       string
	repoBranch    string
	policyDir     string
	policyFileExt string
	cloneFn       CloneFn
}

type GitClient interface {
	Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error)
}

type GitTree interface {
	TreeEntryFile(e *object.TreeEntry) (*object.File, error)
}

type GitTreeWalker interface {
	Next() (name string, entry object.TreeEntry, err error)
	Close()
}

type gitPolicyEntry struct {
	name  string
	entry *object.TreeEntry
}

type GitPolicyFile struct {
	Hash string
	*PolicyFile
}

func NewGitPolicySource(repoURL string, repoBrach string, policyDir string) PolicySource {
	return &GitPolicySource{
		repoUrl:       repoURL,
		repoBranch:    repoBrach,
		policyDir:     policyDir,
		policyFileExt: "rego",
		cloneFn:       git.Clone,
	}
}

func (src *GitPolicySource) GetPolicyFiles() ([]*PolicyFile, error) {
	repo, err := src.clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone GIT repository: %s", err)
	}
	tree, err := src.getHeadTree(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get GIT HEAD ref tree: %s", err)
	}
	entries, err := src.getGitPolicyEntries(object.NewTreeWalker(tree, true, nil))
	if err != nil {
		return nil, fmt.Errorf("failed to list policy files in GIT HEAD ref tree: %s", err)
	}
	files := make([]*PolicyFile, len(entries))
	for i := range entries {
		gitPolicyFile, err := entries[i].readPolicyFile(tree)
		if err != nil {
			return nil, fmt.Errorf("failed to read policy file %q: %s", entries[i].name, err)
		}
		files[i] = gitPolicyFile.PolicyFile
	}
	return files, nil
}

func (src *GitPolicySource) clone() (*git.Repository, error) {
	repo, err := src.cloneFn(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           src.repoUrl,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + src.repoBranch),
		SingleBranch:  true,
	})
	if err != nil {
		return nil, err

	}
	return repo, nil
}

func (src *GitPolicySource) getHeadTree(repo *git.Repository) (*object.Tree, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, err
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}
	return commit.Tree()
}

func (src *GitPolicySource) getGitPolicyEntries(wkr GitTreeWalker) ([]*gitPolicyEntry, error) {
	defer wkr.Close()
	entries := make([]*gitPolicyEntry, 0)
	for {
		name, entry, err := wkr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if entry.Mode == filemode.Dir || entry.Mode == filemode.Submodule {
			continue
		}
		if strings.HasPrefix(name, src.policyDir+"/") && strings.HasSuffix(name, "."+src.policyFileExt) {
			entries = append(entries, &gitPolicyEntry{
				name:  name,
				entry: &entry,
			})
		}
	}
	return entries, nil
}

func (e *gitPolicyEntry) readPolicyFile(t GitTree) (*GitPolicyFile, error) {
	file, err := t.TreeEntryFile(e.entry)
	if err != nil {
		return nil, err
	}
	content, err := file.Contents()
	if err != nil {
		return nil, err
	}
	return &GitPolicyFile{
		Hash: e.entry.Hash.String(),
		PolicyFile: &PolicyFile{
			Name:     e.entry.Name,
			FullName: e.name,
			Content:  content,
		},
	}, nil
}
