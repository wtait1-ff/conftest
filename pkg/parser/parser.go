package parser

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/instrumenta/conftest/pkg/parser/cue"
	"github.com/instrumenta/conftest/pkg/parser/ini"
	"github.com/instrumenta/conftest/pkg/parser/terraform"
	"github.com/instrumenta/conftest/pkg/parser/toml"
	"github.com/instrumenta/conftest/pkg/parser/yaml"
	"github.com/instrumenta/conftest/pkg/parser/docker"
)

// ValidInputs returns string array in order to passing valid input types to viper
func ValidInputs() []string {
	return []string{
		"toml",
		"tf|hcl",
		"cue",
		"ini",
		"yaml",
	}
}

// Parser is the interface implemented by objects that can unmarshal
// bytes into a golang interface
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// ConfigDoc stores file contents and it's original filename
type ConfigDoc struct {
	ReadCloser io.ReadCloser
	Filepath   string
}

// ReadUnmarshaller is an interface that allows for bulk unmarshalling
// and setting of io.Readers to be unmarshalled.
type ReadUnmarshaller interface {
	BulkUnmarshal(readerList []ConfigDoc) (map[string]interface{}, error)
}

// ConfigManager the implementation of ReadUnmarshaller and io.Reader
// byte storage.
type ConfigManager struct {
	parser         Parser
	configContents map[string][]byte
}

// BulkUnmarshal iterates through the given cached io.Readers and
// runs the requested parser on the data.
func (s *ConfigManager) BulkUnmarshal(configList []ConfigDoc) (map[string]interface{}, error) {
	err := s.setConfigs(configList)
	if err != nil {
		return nil, fmt.Errorf("we should not have any errors on setting our readers: %v", err)
	}
	var allContents = make(map[string]interface{})
	for filepath, config := range s.configContents {
		var singleContent interface{}
		err := s.parser.Unmarshal(config, &singleContent)
		if err != nil {
			return nil, fmt.Errorf("we should not have any errors on unmarshalling: %v", err)
		}
		allContents[filepath] = singleContent
	}
	return allContents, nil
}

func (s *ConfigManager) setConfigs(configList []ConfigDoc) error {
	s.configContents = make(map[string][]byte)
	for _, config := range configList {
		if config.ReadCloser == nil {
			return fmt.Errorf("we recieved a nil reader, which should not happen")
		}
		contents, err := ioutil.ReadAll(config.ReadCloser)
		defer config.ReadCloser.Close()
		if err != nil {
			return fmt.Errorf("Error while reading Reader contents; err is: %s", err)
		}
		s.configContents[config.Filepath] = contents
	}
	return nil
}

// NewConfigManager is the instatiation function for ConfigManager
func NewConfigManager(fileType string) ReadUnmarshaller {

	return &ConfigManager{
		parser: GetParser(fmt.Sprintf("file.%s", fileType)),
	}
}

// GetParser gets a parser that works on a given filename
func GetParser(fileName string) Parser {
	suffix := filepath.Ext(fileName)

	switch suffix {
	case ".toml":
		return &toml.Parser{
			FileName: fileName,
		}
	case ".tf":
		return &terraform.Parser{
			FileName: fileName,
		}
	case ".cue":
		return &cue.Parser{
			FileName: fileName,
		}
	case ".ini":
		return &ini.Parser{
			FileName: fileName,
		}
	case "Dockerfile":
		return &docker.Parser{
			FileName: i.fName,
		}
	default:
		return &yaml.Parser{
			FileName: fileName,
		}
	}
}
