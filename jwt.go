package jsonpath

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/steinfletcher/apitest"
)

const (
	jwtHeaderIndex  = 0
	jwtPayloadIndex = 1
)

func JWTHeaderEqual(tokenSelector func(*http.Response) (string, error), expression string, expected interface{}) apitest.Assert {
	return jwtEqual(tokenSelector, expression, expected, jwtHeaderIndex)
}

func JWTPayloadEqual(tokenSelector func(*http.Response) (string, error), expression string, expected interface{}) apitest.Assert {
	return jwtEqual(tokenSelector, expression, expected, jwtPayloadIndex)
}

func jwtEqual(tokenSelector func(*http.Response) (string, error), expression string, expected interface{}, index int) apitest.Assert {
	return func(response *http.Response, request *http.Request) error {
		token, err := tokenSelector(response)
		if err != nil {
			return err
		}

		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			splitErr := errors.New("Invalid token: token should contain header, payload and secret")
			return splitErr
		}

		decodedPayload, PayloadErr := base64Decode(parts[index])
		if PayloadErr != nil {
			return fmt.Errorf("Invalid jwt: %s", PayloadErr.Error())
		}

		value, err := jsonPath(bytes.NewReader(decodedPayload), expression)
		if err != nil {
			return err
		}

		if !objectsAreEqual(value, expected) {
			return errors.New(fmt.Sprintf("\"%s\" not equal to \"%s\"", value, expected))
		}

		return nil
	}
}

func base64Decode(src string) ([]byte, error) {
	if l := len(src) % 4; l > 0 {
		src += strings.Repeat("=", 4-l)
	}

	decoded, err := base64.URLEncoding.DecodeString(src)
	if err != nil {
		errMsg := fmt.Errorf("Decoding Error %s", err)
		return nil, errMsg
	}
	return decoded, nil
}
