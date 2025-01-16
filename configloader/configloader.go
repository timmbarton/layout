package configloader

import (
	"encoding/json"
	"os"

	"github.com/timmbarton/utils/validation"
)

//goland:noinspection ALL
func Load(dest any) error {
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if configFilePath == "" {
		configFilePath = "./.config/config.json"
	}
	return LoadFromJsonFile(configFilePath, dest)
}

func LoadFromJsonFile(filePath string, dest any) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	err = json.NewDecoder(file).Decode(dest)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(dest)
	if err != nil {
		return err
	}

	return nil
}
