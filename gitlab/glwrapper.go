package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

type MultiHostConfigs struct {
	Configs []WrapperConfig
}

func SaveConfig(host string, token string) error {
	configs, err := ReadWrapperSettingsMultiple()

	if err != nil {
		configs = &MultiHostConfigs{}
	}

	configs.Configs = append(configs.Configs, WrapperConfig{URL: "https://" + host, Token: token})

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(homedir, ".gl", "config.json")

	file, err := os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(*configs); err != nil {
		return err
	}
	return nil
}

func ReadWrapperSettingsMultiple() (*MultiHostConfigs, error) {
	homedir, err := os.UserHomeDir()
	config := MultiHostConfigs{}

	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(homedir, ".gl", "config.json")
	file, err := os.Open(configFile)
	if err != nil {
		log.Println("Failed to open file: ", err)
		return nil, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)

	if err != nil {
		log.Println("Failed to parse configuration: ", err)
		return nil, err
	}
	return &config, nil
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
		log.Println("Failed to open file: ", err)
		return config, err
	}
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&config)

	if err != nil {
		log.Println("Failed to parse configuration: ", err)
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

func WrapperFromSettingsMultipleHosts(hostname string) (*gitlabWrapper, error) {
	homedir, err := os.UserHomeDir()
	config := MultiHostConfigs{}

	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(homedir, ".gl", "config.json")
	file, err := os.Open(configFile)
	if err != nil {
		log.Println("Failed to open file: ", err)
		return nil, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)

	if err != nil {
		log.Println("Failed to parse configuration: ", err)
		return nil, err
	}

	for _, cfg := range config.Configs {
		if strings.Contains(cfg.URL, hostname) {
			g, err := NewWrapper(cfg.Token, cfg.URL)
			if err != nil {
				return nil, err
			}
			return &g, nil
		}
	}
	log.Println("Failed to find configuration for host ", hostname)
	return nil, errors.New("Failed to find configuration for the provided hostname")
}

func (wrapper gitlabWrapper) GetProject(project string) (*gitlab.Project, error) {
	git := wrapper.Client
	p, _, err := git.Projects.GetProject(project, &gitlab.GetProjectOptions{})

	if err != nil {
		log.Printf("Failed to get proejcts: %v", err)
		return nil, err
	}

	log.Println("Found project ", p.WebURL)

	log.Println(p.WebURL)

	return p, nil
}

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
		// TargetBranch: &defBranch,
	})

	if mrs != nil {
		for _, mr := range mrs {
			if mr.SourceProjectID == p.ID {
				log.Println("The merge request already exists")
				log.Println(mr.WebURL)
				return mr.WebURL, nil
			}
		}
	}

	return "", errors.New("No merge requests found")
}

type CreateMergeRequestOptions struct {
	Branch        *string
	Project       *string
	SourceBranch  *string
	DeleteOnMerge *bool
	Draft         *bool
}

func (wrapper gitlabWrapper) CreateMergeRequest(options *CreateMergeRequestOptions) (*gitlab.MergeRequest, error) {
	git := wrapper.Client

	project := *options.Project
	var source string = *options.SourceBranch

	var targetBranchName string
	p, _, err := git.Projects.GetProject(project, &gitlab.GetProjectOptions{})
	if err != nil {
		return nil, err
	}

	if options.Branch != nil {
		targetBranchName = *options.Branch
		// todo check if exists
	} else {
		if err != nil {
			return nil, err
		}
		// log.Printf("Default branch: %s", p.DefaultBranch)
		targetBranchName = p.DefaultBranch
	}

	mrs, _, err := git.MergeRequests.ListMergeRequests(&gitlab.ListMergeRequestsOptions{
		State:        gitlab.String("opened"),
		SourceBranch: gitlab.String(source),
		TargetBranch: gitlab.String(targetBranchName),
	})

	for _, mr := range mrs {
		if mr.TargetProjectID == p.ID {
			log.Println("The merge request already exists")
			mr := mrs[0]
			log.Println(mr.WebURL)
			return nil, fmt.Errorf("The merge request already exists: %s", mr.WebURL)
		}
	}

	branch, _, err := git.Branches.GetBranch(project, source)
	if err != nil {
		log.Printf("Failed to fetch branch %s", source)
		return nil, err
	}
	commit := branch.Commit
	log.Printf("Found commit message: %s", commit.Message)

	title := "Resolve " + branch.Name
	log.Printf("Title: " + title)
	if *options.Draft {
		title = "Draft: " + title
	}

	mr, _, err := git.MergeRequests.CreateMergeRequest(project, &gitlab.CreateMergeRequestOptions{
		SourceBranch:       gitlab.String(source),
		TargetBranch:       gitlab.String(targetBranchName),
		Title:              gitlab.String(title),
		RemoveSourceBranch: options.DeleteOnMerge,
	})
	if err != nil {
		log.Println("Failed to create merge request: ", err)
		return nil, err
	}
	log.Printf("Created merge request: %s", mr.WebURL)
	return mr, nil
}
