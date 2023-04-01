package helpers

import (
	"fmt"
	"net/http"
)

const InternalServerError = "internalServerError"
const BadRequest = "badRequest"
const Unauthorized = "unauthorized"

var (
	InfoMsg  = Teal
	WarnMsg  = Yellow
	FatalMsg = Red
)

var (
	Red    = Color("\033[1;31m%s\033[0m")
	Yellow = Color("\033[1;33m%s\033[0m")
	Teal   = Color("\033[1;36m%s\033[0m")
)

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

func GetErrorStatusCode(err error) int {
	statusCode := 500
	switch err.Error() {
	case InternalServerError:
		statusCode = http.StatusInternalServerError
	case BadRequest:
		statusCode = http.StatusBadRequest
	case Unauthorized:
		statusCode = http.StatusUnauthorized
	}
	return statusCode
}

// LogErrorWithMessage returns original error after logging for handling purposes
func LogErrorWithMessage(err error, msg string) error {
	fmt.Println(FatalMsg("{ERROR} [" + err.Error() + "] " + msg))
	return err
}