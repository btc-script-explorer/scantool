package rest

import (
	"fmt"
	"strings"
	"encoding/json"
	"net/http"
)

type RestError struct {
	Error string
}

func RestHandler (response http.ResponseWriter, request *http.Request) {

	// remove leading and trailing slashes
	modifiedPath := request.URL.Path
	if len (modifiedPath) > 0 && modifiedPath [0] == '/' { modifiedPath = modifiedPath [1:] }
	lastChar := len (modifiedPath) - 1
	if lastChar >= 0 && modifiedPath [lastChar] == '/' { modifiedPath = modifiedPath [: lastChar] }

	formatError := len (modifiedPath) == 0

	// at a minimum, there must be at least 3 url parameters

	responseJson := ""
	if !formatError {
		requestParts := strings.Split (modifiedPath, "/")
		formatError = requestParts [0] != "rest"
		if !formatError {
			restAPIVersion := requestParts [1]
			restAPIEndpoint := requestParts [2]
			restAPIParamString := requestParts [3:]
			switch restAPIVersion {
				case "v1":
					restApiV1 := RestApiV1 {}
					responseJson = restApiV1.HandleRequest (request.Method, restAPIEndpoint, restAPIParamString, request.Body)
			}
		}
	}

	if formatError {
		errorMessage := fmt.Sprintf ("Invalid request: %s", request.URL.Path)
		fmt.Println (errorMessage)
		errBytes, _ := json.Marshal (RestError { Error: errorMessage })
		responseJson = string (errBytes)
	}

	fmt.Fprint (response, responseJson)
}

