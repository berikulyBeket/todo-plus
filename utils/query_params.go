package utils

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseRequiredQueryParam retrieves a required query parameter
func ParseRequiredQueryParam(c *gin.Context, key string) (string, error) {
	param := c.Query(key)
	if param == "" {
		return "", errors.New(key + " must be provided")
	}

	return param, nil
}

// ParseOptionalParamAsInt parses an optional query parameter as an integer
func ParseOptionalParamAsInt(c *gin.Context, key string, optionalParam **int) error {
	param := c.Query(key)
	if param != "" {
		value, err := strconv.Atoi(param)
		if err != nil {
			return errors.New(key + " must be a valid integer")
		}

		*optionalParam = &value
	}

	return nil
}

// ParseRequiredParamAsInt retrieves and parses a required query parameter as an integer
func ParseRequiredParamAsInt(c *gin.Context, key string) (int, error) {
	param := c.Query(key)
	if param == "" {
		return 0, errors.New(key + " must be provided")
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, errors.New(key + " must be a valid integer")
	}

	return value, nil
}

// ParseOptionalParamAsBool parses an optional query parameter as a boolean
func ParseOptionalParamAsBool(c *gin.Context, key string, optionalParam **bool) error {
	param := c.Query(key)
	if param != "" {
		value, err := strconv.ParseBool(param)
		if err != nil {
			return errors.New(key + " must be a valid boolean")
		}

		*optionalParam = &value
	}

	return nil
}
