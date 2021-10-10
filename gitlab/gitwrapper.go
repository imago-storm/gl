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
)

type repoWrapper struct {
	Repository *git.Repository
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

func OpenRemote(path string) error {
	r, err := git.PlainOpen(path)

	if err != nil {
		return err
	}

	log.Println("Remotes", r)

	list, err := r.Remotes()
	if err != nil {
		return err
	}

	// git@gitlab.com:925043/cool-project.git

	for _, r := range list {
		log.Println("Remote name: ", r.Config().Name)

		for _, url := range r.Config().URLs {
			log.Println("URL: ", url)
			parts := strings.Split(url, "@")
			if len(parts) != 2 {
				return errors.New("Wrong tokenization of URL: " + url)
			}
			url = parts[1]

			url = string(regexp.MustCompile(`\.git$`).ReplaceAll([]byte(url), []byte("")))
			log.Println(url)

			parts = strings.Split(url, `:`)
			if len(parts) != 2 {
				return errors.New("Wrong tokenization of URL: " + url)
			}
			gitlabSite := parts[0]
			project := parts[1]

			repositoryUrl := fmt.Sprintf("https://%s/%s", gitlabSite, project)
			log.Println("Remote: ", repositoryUrl)

			openBrowser(repositoryUrl)
		}
	}

	return nil
}

// func GetProjectName(path string) (string, error) {
// 	r, err := git.PlainOpen(path)

// 	if err != nil {
// 		return "", err
// 	}

// 	list, err := r.Remotes()
// 	if err != nil {
// 		return "", err
// 	}

// 	// git@gitlab.com:925043/cool-project.git

// 	for _, r := range list {
// 		log.Println("Remote name: ", r.Config().Name)

// 		for _, url := range r.Config().URLs {
// 			log.Println("URL: ", url)
// 			parts := strings.Split(url, "@")
// 			if len(parts) != 2 {
// 				return "", errors.New("Wrong tokenization of URL: " + url)
// 			}
// 			url = parts[1]

// 			url = string(regexp.MustCompile(`\.git$`).ReplaceAll([]byte(url), []byte("")))
// 			log.Println(url)

// 			parts = strings.Split(url, `:`)
// 			if len(parts) != 2 {
// 				return "", errors.New("Wrong tokenization of URL: " + url)
// 			}
// 			// gitlabSite := parts[0]

// 			project := parts[1]
// 			// todo default remote
// 			return project, nil
// 		}
// 	}
// 	return "", errors.New("No remotes found")
// }

// func OpenCurrentMergeRequest(path string) error {
// 	r, err := git.PlainOpen(path)

// 	if err != nil {
// 		return err
// 	}

// 	ref, err := r.Head()
// 	if err != nil {
// 		return err
// 	}
// 	log.Println(ref.Target())
// 	log.Println(ref.Name())

// 	log.Println("Ref: ", ref)
// 	s := strings.Split(ref.Name().String(), "/")
// 	branch := s[2]
// 	log.Println("branch: ", s[2])

// 	projectName, err := GetProjectName(path)
// 	if err != nil {
// 		return err
// 	}

// 	gitlab, err := WrapperFromSettings()
// 	url, err := gitlab.GetMergeRequestURL(projectName, branch)
// 	if err != nil {
// 		return err
// 	}

// 	openBrowser(url)
// 	return nil
// }

// func CreateMergeRequest(path string) error {
// 	r, err := git.PlainOpen(path)

// 	if err != nil {
// 		return err
// 	}

// 	ref, err := r.Head()
// 	if err != nil {
// 		return err
// 	}
// 	log.Println(ref.Target())
// 	log.Println(ref.Name())

// 	log.Println("Ref: ", ref)
// 	s := strings.Split(ref.Name().String(), "/")
// 	branch := s[2]
// 	log.Println("branch: ", s[2])

// 	projectName, err := GetProjectName(path)
// 	if err != nil {
// 		return err
// 	}

// 	gitlab, err := WrapperFromSettings()
// 	if err != nil {
// 		return err
// 	}
// 	log.Println(gitlab)
// 	log.Println("Found project ", projectName)
// 	err = gitlab.CreateMergeRequest(projectName, branch)

// 	return err
// }

func (wrapper repoWrapper) getProjectName() (string, error) {
	list, err := wrapper.Repository.Remotes()
	if err != nil {
		return "", err
	}

	// git@gitlab.com:925043/cool-project.git

	for _, r := range list {
		log.Println("Remote name: ", r.Config().Name)

		for _, url := range r.Config().URLs {
			log.Println("URL: ", url)
			parts := strings.Split(url, "@")
			if len(parts) != 2 {
				return "", errors.New("Wrong tokenization of URL: " + url)
			}
			url = parts[1]

			url = string(regexp.MustCompile(`\.git$`).ReplaceAll([]byte(url), []byte("")))
			log.Println(url)

			parts = strings.Split(url, `:`)
			if len(parts) != 2 {
				return "", errors.New("Wrong tokenization of URL: " + url)
			}
			// gitlabSite := parts[0]

			project := parts[1]
			// todo default remote
			return project, nil
		}
	}
	return "", errors.New("No remotes found")
}

func (wrapper repoWrapper) CreateMergeRequest() error {
	branch, err := wrapper.getCurrentBranch()

	if err != nil {
		return err
	}

	projectName, err := wrapper.getProjectName()
	if err != nil {
		return err
	}

	gitlab, err := WrapperFromSettings()
	if err != nil {
		return err
	}
	err = gitlab.CreateMergeRequest(projectName, branch)

	return err
}

func (repo repoWrapper) OpenMergeRequest() error {
	project, err := repo.getProjectName()
	if err != nil {
		return err
	}
	branch, err := repo.getCurrentBranch()
	if err != nil {
		return err
	}

	gitlab, err := WrapperFromSettings()
	url, err := gitlab.GetMergeRequestURL(project, branch)
	if err != nil {
		return err
	}

	openBrowser(url)
	return nil
}

func (wrapper repoWrapper) getCurrentBranch() (string, error) {
	ref, err := wrapper.Repository.Head()
	if err != nil {
		return "", err
	}
	s := strings.Split(ref.Name().String(), "/")
	if len(s) != 3 {
		return "", fmt.Errorf("Invalid reference: %s", ref.Name())
	}
	return s[2], nil
}

func openBrowser(url string) error {
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
