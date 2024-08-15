// Package database
// Author: Evsikov Artem

package oauth

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/auth0/go-auth0/authentication"
	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout)
// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
	Sub   string `json:"sub"`
}

// Validate does nothing for this example, but we need
// it to satisfy validator.CustomClaims interface.
func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
        // Fatal because URL doesn't affected by user
		logger.Err(err).Msgf("")
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		logger.Err(err).Msgf("")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

func ValidateToken(token string) bool {
    issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
    if err != nil {
        // Fatal because URL doesn't affected by user
        logger.Err(err).Msgf("")
    }

    provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

    jwtValidator, err := validator.New(
        provider.KeyFunc,
        validator.RS256,
        issuerURL.String(),
        []string{os.Getenv("AUTH0_AUDIENCE")},
        validator.WithCustomClaims(
            func() validator.CustomClaims {
                return &CustomClaims{}
            },
        ),
        validator.WithAllowedClockSkew(time.Minute),
    )
    if err != nil {
        logger.Err(err).Msgf("")
    }

    _, err = jwtValidator.ValidateToken(context.TODO(), token)
    if err != nil {
        return false
    }
    return true
}

// HasScope checks whether our claims have a specific scope.
func (c CustomClaims) HasScope(expectedScope string) bool {
	result := strings.Split(c.Scope, " ")
	for i := range result {
		if result[i] == expectedScope {
			return true
		}
	}

	return false
}

func GetAuthClient() (*authentication.Authentication, error) {
    a, err := authentication.New(
        context.TODO(),
        os.Getenv("AUTH0_DOMAIN"),
        authentication.WithClientID(os.Getenv("AUTH0_CLIENT_ID")),
        authentication.WithClientSecret(os.Getenv("AUTH0_CLIENT_SECRET")), // Optional depending on the grants used
    )
    return a, err    
}
