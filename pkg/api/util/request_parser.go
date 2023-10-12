package util

import (
	"encoding/json"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"io"
	"log"
	"strings"
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
		b, err := json.Marshal(checkR)
		if err != nil {
			log.Println("WARN: Could not marshal checkR: " + err.Error())
		} else {
			log.Println(string(b))
		}
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
