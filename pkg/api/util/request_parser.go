package util

import (
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/SENERGY-Platform/budget/pkg/log"
	"github.com/SENERGY-Platform/budget/pkg/models"
)

type checkRequest struct {
	Headers  headers           `json:"headers"`
	UriArgs  map[string]string `json:"uri_args"`
	BodyData string            `json:"body_data"`
}

type headers struct {
	TargetMethod  string `json:"target_method"`
	TargetUri     string `json:"target_uri"`
	XUserRoles    string `json:"x-user-roles"`
	XUserId       string `json:"x-userid"`
	Authorization string `json:"authorization"`
}

func ParseRequest(body io.Reader, debug bool) (*models.ParsedRequest, error) {
	var checkR checkRequest
	err := json.NewDecoder(body).Decode(&checkR)
	if err != nil {
		return nil, err
	}

	if debug {
		logParsedRequest(checkR)
	}

	return &models.ParsedRequest{
		TargetMethod: checkR.Headers.TargetMethod,
		TargetUri:    checkR.Headers.TargetUri,
		Roles:        strings.Split(checkR.Headers.XUserRoles, ", "),
		UserId:       checkR.Headers.XUserId,
		AuthToken:    checkR.Headers.Authorization,
		UriArgs:      checkR.UriArgs,
		BodyData:     []byte(checkR.BodyData),
	}, nil
}

// logParsedRequest logs the request without credentials: the authorization header is reduced to its scheme,
// uri args to their keys and the body to its size.
func logParsedRequest(r checkRequest) {
	argKeys := make([]string, 0, len(r.UriArgs))
	for k := range r.UriArgs {
		argKeys = append(argKeys, k)
	}
	sort.Strings(argKeys)
	log.Logger.Debug("parsed request",
		"target_method", r.Headers.TargetMethod,
		"target_uri", r.Headers.TargetUri,
		"x-user-roles", r.Headers.XUserRoles,
		"x-userid", r.Headers.XUserId,
		"authorization_scheme", authorizationScheme(r.Headers.Authorization),
		"uri_arg_keys", argKeys,
		"body_data_size", len(r.BodyData),
	)
}

func authorizationScheme(authorization string) string {
	if authorization == "" {
		return "none"
	}
	scheme, _, _ := strings.Cut(authorization, " ")
	switch strings.ToLower(scheme) {
	case "bearer":
		return "Bearer"
	case "basic":
		return "Basic"
	default:
		return "other"
	}
}
