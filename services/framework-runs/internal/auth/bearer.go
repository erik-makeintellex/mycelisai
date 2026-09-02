package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
)

type contextKey struct{}

type Principal struct {
	ID     string
	Scopes map[string]bool
}

type Credential struct {
	Principal Principal
	digest    [sha256.Size]byte
}

type Authenticator struct{ credentials []Credential }

func NewCredential(id, token string, scopes ...string) (Credential, error) {
	if strings.TrimSpace(id) == "" {
		return Credential{}, errors.New("credential principal id is required")
	}
	if token != strings.TrimSpace(token) || len(token) < 32 {
		return Credential{}, errors.New("credential token must be canonical and at least 32 bytes")
	}
	scopeSet := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == "" {
			return Credential{}, errors.New("credential scope is empty")
		}
		scopeSet[scope] = true
	}
	if len(scopeSet) == 0 {
		return Credential{}, errors.New("credential requires at least one scope")
	}
	return Credential{
		Principal: Principal{ID: id, Scopes: scopeSet},
		digest:    sha256.Sum256([]byte(token)),
	}, nil
}

func New(credentials ...Credential) (*Authenticator, error) {
	if len(credentials) == 0 {
		return nil, errors.New("at least one service credential is required")
	}
	return &Authenticator{credentials: append([]Credential(nil), credentials...)}, nil
}

func (a *Authenticator) Middleware(scope string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		principal, state := a.authenticate(request)
		switch state {
		case http.StatusUnauthorized:
			writeError(response, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
			return
		case http.StatusForbidden:
			writeError(response, http.StatusForbidden, "forbidden", "The service credential lacks the required scope.")
			return
		}
		if !principal.Scopes["*"] && !principal.Scopes[scope] {
			writeError(response, http.StatusForbidden, "forbidden", "The service credential lacks the required scope.")
			return
		}
		ctx := context.WithValue(request.Context(), contextKey{}, principal)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(contextKey{}).(Principal)
	return principal, ok
}

func (a *Authenticator) authenticate(request *http.Request) (Principal, int) {
	values := request.Header.Values("Authorization")
	if len(values) != 1 || !strings.HasPrefix(values[0], "Bearer ") {
		return Principal{}, http.StatusUnauthorized
	}
	token := strings.TrimPrefix(values[0], "Bearer ")
	if token == "" || token != strings.TrimSpace(token) || strings.Contains(token, " ") {
		return Principal{}, http.StatusUnauthorized
	}
	digest := sha256.Sum256([]byte(token))
	for _, credential := range a.credentials {
		if subtle.ConstantTimeCompare(digest[:], credential.digest[:]) == 1 {
			return credential.Principal, http.StatusOK
		}
	}
	return Principal{}, http.StatusUnauthorized
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_, _ = response.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `","recoverable":false}}`))
}
