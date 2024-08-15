package router

import (
	"einstein-server/database"
	"einstein-server/oauth"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

func GetUser(r *http.Request) (database.User, string) {
	var user database.User
	db := database.GetConnection()
	token := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	claims := token.CustomClaims.(*oauth.CustomClaims)
	db.Get(&user, database.SqlUserSelectByExtId, claims.Sub)
	db.Close()
	return user, claims.Sub
}
