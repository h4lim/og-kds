package infra

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	MessageEN map[string]string
	MessageID map[string]string
)

type MessageModel struct {
	Path     string
	FileName string
}

type IMessageConfig interface {
	Setup() *error
}

func NewMessageConfig(model MessageModel) IMessageConfig {
	return MessageModel{
		Path:     model.Path,
		FileName: model.FileName,
	}
}

func (m MessageModel) Setup() *error {

	fullPath := filepath.Join(m.Path, m.FileName)
	file, err := os.Open(fullPath)
	if err != nil {
		return &err
	}

	byteJson, err := ioutil.ReadAll(file)
	if err != nil {
		return &err
	}

	messageMap := make(map[string]map[string]string)
	if err := json.Unmarshal(byteJson, &messageMap); err != nil {
		return &err
	}

	messageEN := make(map[string]string)
	messageID := make(map[string]string)
	for k, v := range messageMap {
		if strings.ToUpper(k) == "EN" {
			for enK, enV := range v {
				messageEN[enK] = enV
			}
		} else {
			for idK, idV := range v {
				messageID[idK] = idV
			}
		}
	}

	MessageEN = messageEN
	MessageID = messageID

	return nil
}
