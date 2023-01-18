package authz

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/form3tech-oss/jwt-go"
)

type CasbinAuthorizer struct {
	Enforcer *casbin.Enforcer
}

func (c *CasbinAuthorizer) Load(params ...interface{}) error {
	var err error
	c.Enforcer, err = casbin.NewEnforcer(params...)
	return err
}

func (c *CasbinAuthorizer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Permission with casbin
		allowed, err := c.CheckPermission(r)
		if err != nil {
			// Casbin.Enforcer not working normal
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if !allowed {
			// Write an error and stop the handler chain
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// Pass down the request to the next middleware (or final handler)
		next.ServeHTTP(w, r)
	})
}

// GetUserName gets the user name from the request.
// Currently, only HTTP basic authentication is supported
func (c *CasbinAuthorizer) GetUserName(r *http.Request) string {
	tokenString := r.Header.Get("Authorization")
	splitToken := strings.Split(tokenString, "Bearer ")
	claims := jwt.MapClaims{}
	_, _ = jwt.ParseWithClaims(splitToken[1], claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("stringtodescriptjwt"), nil
	})
	role := ""
	//fmt.Println("Token:%v", splitToken)20
	for k, v := range claims {
		if k == "Role" {
			role = fmt.Sprintf("%v", v)
		}
	}
	return role
}

// CheckPermission checks the user/method/path combination from the request.
// Returns true (permission granted) or false (permission forbidden)
func (c *CasbinAuthorizer) CheckPermission(r *http.Request) (bool, error) {

	user := c.GetUserName(r)
	method := r.Method
	path := r.URL.Path
	return c.Enforcer.Enforce(user, path, method)
}
