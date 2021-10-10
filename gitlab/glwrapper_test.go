package gitlab

import (
	"log"
	"os"
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
