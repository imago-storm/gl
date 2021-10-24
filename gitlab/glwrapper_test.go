package gitlab

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestGetProject(t *testing.T) {
	token := os.Getenv("TOKEN")
	gl, err := NewWrapper(token, "https://gitlab.com")
	log.Println("GL: ", gl)
	if err != nil {
		t.Error("Failed test: err ", err)
	}
	p, err := gl.GetProject("925043/cool-project")
	log.Println("Project: ", p)
}

func TestReadWrapperSettings(t *testing.T) {
	settings, err := ReadWrapperSettings()
	if err != nil {
		t.Error("Failed", err)
	}
	log.Println(settings)
}

func TestSaveConfig(t *testing.T) {
	err := SaveConfig("host", "token")
	if err != nil {
		t.Error("Failed", err)
	}

	homedir, _ := os.UserHomeDir()
	dat, err := ioutil.ReadFile(filepath.Join(homedir, ".gl", "config.json"))
	log.Println(dat)
}
