package internal

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

var InternalServerError = errors.New("internalServerError")
var BadRequestError = errors.New("badRequest")
var UnauthorizedError = errors.New("unauthorized")
var ForbiddenError = errors.New("forbidden")
var VideoDurationTooShortError = errors.New("videoDurationTooShort")
var AlreadyExistsError = errors.New("alreadyExists")
var NotFoundError = errors.New("notFound")
var MagnetInfoTimeoutError = errors.New("magnetInfoFailed")
var GatewayTimeoutError = errors.New("gatewayTimeout")

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
	// if errors.Is(err, InternalServerError) {
	// 	return http.StatusInternalServerError
	// }
	if errors.Is(err, BadRequestError) {
		return http.StatusBadRequest
	}
	if errors.Is(err, UnauthorizedError) {
		return http.StatusUnauthorized
	}
	if errors.Is(err, ForbiddenError) {
		return http.StatusForbidden
	}
	if errors.Is(err, NotFoundError) {
		return http.StatusNotFound
	}
	if errors.Is(err, VideoDurationTooShortError) {
		return http.StatusInternalServerError
	}
	if errors.Is(err, AlreadyExistsError) {
		return http.StatusConflict
	}
	if errors.Is(err, MagnetInfoTimeoutError) || errors.Is(err, GatewayTimeoutError) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}

// LogErrorWithMessage returns original error after logging for handling purposes
func LogErrorWithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	slog.Error(msg, "err", err)
	return err
}
