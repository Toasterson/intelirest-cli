package runtime

import (
	"encoding/json"
	"fmt"
	"os"
)

// EnvironmentFileName is the default name for the environments file from inteliJ
const EnvironmentFileName = "rest-client.env.json"

// EnvFile is a Helper type to parse the environments file into
type EnvFile map[string]map[string]string

// ReadEnvironment gets the environment variables from the default file location returns nil if it does not exist
func ReadEnvironment(name string) (map[string]string, error) {
	if name == "" {
		return nil, nil
	}

	if _, err := os.Stat(EnvironmentFileName); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var fileStruct EnvFile

	f, err := os.Open(EnvironmentFileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&fileStruct); err != nil {
		return nil, err
	}

	if env, ok := fileStruct[name]; ok {
		return env, nil
	}

	return nil, fmt.Errorf("environment %s does not exist in file", name)

}
