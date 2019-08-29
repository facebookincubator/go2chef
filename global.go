package go2chef

import (
	"errors"
	"log"
	"os"
)

var (
	// AutoRegisterPlugins is a central place for plugins to check
	// whether they should auto-register. Normally they should.
	AutoRegisterPlugins = true
	// EarlyLogger is the logger used for pre-config logging. You can
	// substitute it if your use case requires.
	EarlyLogger = log.New(os.Stderr, "GO2CHEF ", log.LstdFlags)
)

var (
	// ErrConfigHasNoNameKey is thrown when a config block doesn't have a `name` key
	ErrConfigHasNoNameKey = errors.New("config map has no `name` key")
	// ErrConfigHasNoTypeKey is thrown when a config block doesn't have a `type` key
	ErrConfigHasNoTypeKey = errors.New("config map has no `type` key")
)

// GetNameType gets the name and type attributes from a config map
func GetNameType(config map[string]interface{}) (string, string, error) {
	cname, err := GetName(config)
	if err != nil {
		return "", "", ErrConfigHasNoNameKey
	}
	ctype, err := GetType(config)
	if err != nil {
		return "", "", ErrConfigHasNoTypeKey
	}
	return cname, ctype, nil
}

// GetName gets the name attribute from a config map
func GetName(config map[string]interface{}) (string, error) {
	cname, ok := config["name"]
	if !ok {
		return "", ErrConfigHasNoNameKey
	}
	if _, ok := cname.(string); !ok {
		return "", ErrConfigHasNoNameKey
	}
	return cname.(string), nil
}

// GetType gets the type attribute from a config map
func GetType(config map[string]interface{}) (string, error) {
	ctype, ok := config["type"]
	if !ok {
		return "", ErrConfigHasNoTypeKey
	}
	if _, ok := ctype.(string); !ok {
		return "", ErrConfigHasNoTypeKey
	}
	return ctype.(string), nil
}

// ErrComponentDoesNotExist represents errors where a requested component (plugin)
// hasn't been registered.
type ErrComponentDoesNotExist struct {
	Component string
}

// Error returns the error string
func (e *ErrComponentDoesNotExist) Error() string {
	return "component " + e.Component + " does not exist"
}

// ErrChefAlreadyInstalled represents errors where Chef is already installed and
// provides a mechanism to pass metadata regarding the installed and requested
// versions back up the error chain.
type ErrChefAlreadyInstalled struct {
	Installed string
	Requested string
}

// Error returns the error string
func (e *ErrChefAlreadyInstalled) Error() string {
	return "Chef is already installed: " + e.Installed + ", requested " + e.Requested
}

// PathExists returns whether the given file or directory exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
