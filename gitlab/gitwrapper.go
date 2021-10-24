package gitlab

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/xanzy/go-gitlab"
)

type repoWrapper struct {
	Repository *git.Repository
}

type CreateMergeRequest struct {
	Branch        *string
	DeleteOnMerge *bool
	Draft         *bool
}

func OpenRepository(path string) (repoWrapper, error) {
	r, err := git.PlainOpen(path)
	wrapper := repoWrapper{}
	if err != nil {
		return wrapper, err
	}

	wrapper.Repository = r
	return wrapper, nil
}

func OpenRepositoryCwd() (repoWrapper, error) {
	path, err := os.Getwd()
	if err != nil {
		return repoWrapper{}, err
	}
	return OpenRepository(path)
}

func (repo repoWrapper) OpenRemoteURL() error {
	site, project, err := repo.getRemoteData()
	if err != nil {
		return err
	}
	repositoryURL := fmt.Sprintf("https://%s/%s", site, project)
	log.Printf("Remote repository URL: %s", repositoryURL)
	return OpenBrowser(repositoryURL)
}

func (repo repoWrapper) getRemoteData() (string, string, error) {
	list, err := repo.Repository.Remotes()
	if err != nil {
		return "", "", err
	}

	for _, r := range list {
		for _, url := range r.Config().URLs {

			parts := strings.Split(url, "@")
			if len(parts) != 2 {
				return "", "", errors.New("Wrong tokenization of URL: " + url)
			}
			url = parts[1]

			url = string(regexp.MustCompile(`\.git$`).ReplaceAll([]byte(url), []byte("")))
			log.Println(url)

			parts = strings.Split(url, `:`)
			if len(parts) != 2 {
				return "", "", errors.New("Wrong tokenization of URL: " + url)
			}
			gitlabSite := parts[0]

			project := parts[1]
			// todo default remote
			return gitlabSite, project, nil
		}
	}
	return "", "", errors.New("No remotes found")
}

func (repo repoWrapper) CreateMergeRequest(options *CreateMergeRequest) (*gitlab.MergeRequest, error) {
	branch, err := repo.getCurrentBranch()

	if err != nil {
		return nil, err
	}

	gitlabSite, projectName, err := repo.getRemoteData()
	if err != nil {
		return nil, err
	}

	gitlab, err := WrapperFromSettingsMultipleHosts(gitlabSite)
	if err != nil {
		return nil, err
	}
	return gitlab.CreateMergeRequest(&CreateMergeRequestOptions{
		Project:       &projectName,
		SourceBranch:  &branch,
		Branch:        options.Branch,
		DeleteOnMerge: options.DeleteOnMerge,
		Draft:         options.Draft,
	})
}

func (repo repoWrapper) OpenMergeRequest() error {
	gitlabSite, project, err := repo.getRemoteData()
	if err != nil {
		return err
	}
	branch, err := repo.getCurrentBranch()
	if err != nil {
		return err
	}

	gitlab, err := WrapperFromSettingsMultipleHosts(gitlabSite)
	url, err := gitlab.GetMergeRequestURL(project, branch)
	if err != nil {
		return err
	}

	OpenBrowser(url)
	return nil
}

func (repo repoWrapper) getCurrentBranch() (string, error) {
	ref, err := repo.Repository.Head()
	if err != nil {
		return "", err
	}
	log.Println(ref)
	s := strings.Split(ref.Name().String(), "/")

	if len(s) < 3 {
		return "", fmt.Errorf("Invalid reference: %s", ref.Name())
	}
	branchParts := s[2:]

	return strings.Join(branchParts, "/"), nil
}

func OpenBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
		return err
	}

	return err
}
