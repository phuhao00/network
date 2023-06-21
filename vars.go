package network

import (
	"context"
	"net/http"
)

func Vars(r *http.Request) map[string]string {
	// vars doesnt exist yet, return empty map
	rawVars := r.Context().Value(varsKey)
	if rawVars == nil {
		return map[string]string{}
	}

	vars, _ := rawVars.(map[string]string)
	return vars
}

func SetRouteVars(r *http.Request, val interface{}) {
	if val == nil {
		return
	}

	r2 := r.WithContext(context.WithValue(r.Context(), varsKey, val))
	*r = *r2
}

type contextKey int

const varsKey contextKey = 2
