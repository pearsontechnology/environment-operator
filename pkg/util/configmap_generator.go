package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// ConfigMapGenerator supports stable generation of a configMap.
type ConfigMapGenerator struct {
	// Name of configMap (required)
	Name string
	// FileSources to derive the configMap from (optional)
	FileSources []string
	// AppendHash; if true, derive a hash from the ConfigMap and append it to the name
	AppendHash bool
}

// Generate outputs a configMap object using the configured fields.
func (s ConfigMapGenerator) Generate() (*v1.ConfigMap, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	configMap := &v1.ConfigMap{}
	configMap.Name = s.Name
	configMap.Data = map[string]string{}
	if len(s.FileSources) > 0 {
		if err := handleConfigMapFromFileSources(configMap, s.FileSources); err != nil {
			return nil, err
		}
	}
	if s.AppendHash {
		h, err := configMapHash(configMap)
		if err != nil {
			return nil, err
		}
		configMap.Name = fmt.Sprintf("%s-%s", configMap.Name, h)
	}
	return configMap, nil
}

// validate validates required fields are set to support structured generation.
func (s ConfigMapGenerator) validate() error {
	if len(s.Name) == 0 {
		return fmt.Errorf("name must be specified")
	}
	if len(s.FileSources) > 0 {
		return fmt.Errorf("from-env-file cannot be combined with from-file or from-literal")
	}
	return nil
}

// handleConfigMapFromFileSources adds the specified file source information
// into the provided configMap
func handleConfigMapFromFileSources(configMap *v1.ConfigMap, fileSources []string) error {
	for _, filePath := range fileSources {
		keyName := path.Base(filePath)

		info, err := os.Stat(filePath)
		if err != nil {
			switch err := err.(type) {
			case *os.PathError:
				return fmt.Errorf("error reading %s: %v", filePath, err.Err)
			default:
				return fmt.Errorf("error reading %s: %v", filePath, err)
			}
		}
		if info.IsDir() {
			fileList, err := ioutil.ReadDir(filePath)
			if err != nil {
				return fmt.Errorf("error listing files in %s: %v", filePath, err)
			}
			for _, item := range fileList {
				itemPath := path.Join(filePath, item.Name())
				if item.Mode().IsRegular() {
					keyName = item.Name()
					err = addKeyFromFileToConfigMap(configMap, keyName, itemPath)
					if err != nil {
						return err
					}
				}
			}
		} else {
			if err := addKeyFromFileToConfigMap(configMap, keyName, filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

// configMapHash returns a hash of the ConfigMap.
// The Data, Kind, and Name are taken into account.
func configMapHash(cm *v1.ConfigMap) (string, error) {
	encoded, err := encodeConfigMap(cm)
	if err != nil {
		return "", err
	}
	h, err := encodeHash(hash(encoded))
	if err != nil {
		return "", err
	}
	return h, nil
}

// encodeConfigMap encodes a ConfigMap.
// Data, Kind, and Name are taken into account.
func encodeConfigMap(cm *v1.ConfigMap) (string, error) {
	// json.Marshal sorts the keys in a stable order in the encoding
	data, err := json.Marshal(map[string]interface{}{"kind": "ConfigMap", "name": cm.Name, "data": cm.Data})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// encodeHash extracts the first 40 bits of the hash from the hex string
// (1 hex char represents 4 bits), and then maps vowels and vowel-like hex
// characters to consonants to prevent bad words from being formed (the theory
// is that no vowels makes it really hard to make bad words). Since the string
// is hex, the only vowels it can contain are 'a' and 'e'.
// We picked some arbitrary consonants to map to from the same character set as GenerateName.
// See: https://github.com/kubernetes/apimachinery/blob/dc1f89aff9a7509782bde3b68824c8043a3e58cc/pkg/util/rand/rand.go#L75
// If the hex string contains fewer than ten characters, returns an error.
func encodeHash(hex string) (string, error) {
	if len(hex) < 10 {
		return "", fmt.Errorf("the hex string must contain at least 10 characters")
	}
	enc := []rune(hex[:10])
	for i := range enc {
		switch enc[i] {
		case '0':
			enc[i] = 'g'
		case '1':
			enc[i] = 'h'
		case '3':
			enc[i] = 'k'
		case 'a':
			enc[i] = 'm'
		case 'e':
			enc[i] = 't'
		}
	}
	return string(enc), nil
}

// hash hashes `data` with sha256 and returns the hex string
func hash(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

// addKeyFromFileToConfigMap adds a key with the given name to a ConfigMap, populating
// the value with the content of the given file path, or returns an error.
func addKeyFromFileToConfigMap(configMap *v1.ConfigMap, keyName, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return addKeyFromLiteralToConfigMap(configMap, keyName, string(data))
}

// addKeyFromLiteralToConfigMap adds the given key and data to the given config map,
// returning an error if the key is not valid or if the key already exists.
func addKeyFromLiteralToConfigMap(configMap *v1.ConfigMap, keyName, data string) error {
	// Note, the rules for ConfigMap keys are the exact same as the ones for SecretKeys.
	if errs := validation.IsConfigMapKey(keyName); len(errs) != 0 {
		return fmt.Errorf("%q is not a valid key name for a ConfigMap: %s", keyName, strings.Join(errs, ";"))
	}
	if _, entryExists := configMap.Data[keyName]; entryExists {
		return fmt.Errorf("cannot add key %s, another key by that name already exists: %v.", keyName, configMap.Data)
	}
	configMap.Data[keyName] = data
	return nil
}
