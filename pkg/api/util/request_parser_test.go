package util

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/SENERGY-Platform/budget/pkg/log"
)

func TestParseRequestDebugLogOmitsCredentials(t *testing.T) {
	var buf bytes.Buffer
	orig := log.Logger
	defer func() { log.Logger = orig }()
	log.Logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	const (
		token  = "secret-token-value"
		body   = "secret-body-value"
		uriArg = "secret-arg-value"
	)
	input := `{"headers":{"target_method":"GET","target_uri":"/x","x-user-roles":"user, admin","x-userid":"u1","authorization":"Bearer ` + token + `"},` +
		`"uri_args":{"access_token":"` + uriArg + `"},"body_data":"` + body + `"}`

	parsed, err := ParseRequest(strings.NewReader(input), true)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.AuthToken != "Bearer "+token {
		t.Errorf("auth token must still be parsed, got %q", parsed.AuthToken)
	}

	out := buf.String()
	if !strings.Contains(out, "parsed request") {
		t.Fatalf("expected debug line, got %q", out)
	}
	for _, secret := range []string{token, body, uriArg} {
		if strings.Contains(out, secret) {
			t.Errorf("log output contains credential %q: %s", secret, out)
		}
	}
	if !strings.Contains(out, `"authorization_scheme":"Bearer"`) {
		t.Errorf("expected authorization scheme in log, got %s", out)
	}
}
