package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/xanzy/go-gitlab"
)

type gitlabWrapper struct {
	BaseURL string
	Client  *gitlab.Client
}

type WrapperConfig struct {
	URL   string
	Token string
}

func ReadWrapperSettings() (WrapperConfig, error) {
	homedir, err := os.UserHomeDir()
	config := WrapperConfig{}
	if err != nil {
		return config, err
	}

	configFile := filepath.Join(homedir, "gl.json")
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatal("Failed to open file: ", err)
		return config, err
	}
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)

	if err != nil {
		log.Fatal("Failed to parse configuration: ", err)
		return config, err
	}
	return config, nil
}

func NewWrapper(token string, baseURL string) (gitlabWrapper, error) {
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return gitlabWrapper{}, err
	}
	g := gitlabWrapper{
		BaseURL: baseURL,
		Client:  client,
	}
	return g, nil
}

func WrapperFromSettings() (gitlabWrapper, error) {
	settings, err := ReadWrapperSettings()

	if err != nil {
		return gitlabWrapper{}, err
	}

	client, err := gitlab.NewClient(settings.Token, gitlab.WithBaseURL(settings.URL))
	g := gitlabWrapper{
		BaseURL: settings.URL,
		Client:  client,
	}

	return g, nil
}

func (wrapper gitlabWrapper) GetProject(project string) (*gitlab.Project, error) {
	git := wrapper.Client
	p, _, err := git.Projects.GetProject(project, &gitlab.GetProjectOptions{})

	if err != nil {
		log.Fatalf("Failed to get proejcts: %v", err)
		return nil, err
	}

	log.Println("Found project ", p.WebURL)

	log.Println(p.WebURL)

	return p, nil
}

// func (wrapper gitlabWrapper) GetMergeRequests(project string) error {
// 	git := wrapper.Client

// 	requests, _, err := git.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
// 		State: gitlab.String("opened"),
// 	})

// 	if err != nil {
// 		return err
// 	}
// 	log.Println("requests: ", requests)
// 	return nil
// }

func (wrapper gitlabWrapper) GetMergeRequestURL(project string, source string) (string, error) {
	git := wrapper.Client

	// todo default branch
	p, _, err := git.Projects.GetProject(project, &gitlab.GetProjectOptions{})
	if err != nil {
		return "", err
	}

	defBranch := p.DefaultBranch
	log.Println(defBranch)

	mrs, _, err := git.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		State:        gitlab.String("opened"),
		SourceBranch: gitlab.String(source),
		TargetBranch: &defBranch,
	})

	if mrs != nil {
		log.Println("The merge request already exists")
		mr := mrs[0]
		log.Println(mr.WebURL)
		return mr.WebURL, nil
	}

	return "", errors.New("No merge requests found")
}

func (wrapper gitlabWrapper) CreateMergeRequest(project string, source string) (*gitlab.MergeRequest, error) {
	git := wrapper.Client

	p, _, err := git.Projects.GetProject(project, &gitlab.GetProjectOptions{})
	if err != nil {
		return nil, err
	}
	log.Printf("Default branch: %s", p.DefaultBranch)

	mrs, _, err := git.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		State:        gitlab.String("opened"),
		SourceBranch: gitlab.String(source),
		TargetBranch: &p.DefaultBranch,
	})

	if len(mrs) > 0 {
		log.Println("The merge request already exists")
		mr := mrs[0]
		log.Println(mr.WebURL)
		return nil, fmt.Errorf("The merge request already exists: %s", mr.WebURL)
	}

	branch, _, err := git.Branches.GetBranch(project, source)
	if err != nil {
		log.Fatalf("Failed to fetch branch %s", source)
		return nil, err
	}
	commit := branch.Commit
	log.Printf("Found commit message: %s", commit.Message)

	options := &gitlab.CreateMergeRequestOptions{
		SourceBranch: gitlab.String(source),
		TargetBranch: &p.DefaultBranch,
		Title:        &commit.Message,
	}
	mr, _, err := git.MergeRequests.CreateMergeRequest(project, options)
	if err != nil {
		log.Fatal("Failed to create merge request: ", err)
	}
	log.Printf("Created merge request: %s", mr.WebURL)
	return mr, nil
}
