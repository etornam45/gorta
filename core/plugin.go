package core

import "net/http"

type Plugin interface {
	Name()   string
	Routes() []Route
}

type RouteMethod string

const (
	GET     RouteMethod = "GET"
	POST    RouteMethod = "POST"
	PUT     RouteMethod = "PUT"
	DELETE  RouteMethod = "DELETE"
	PATCH   RouteMethod = "PATCH"
	OPTIONS RouteMethod = "OPTIONS"
	HEAD    RouteMethod = "HEAD"
)

type Route struct {
	Method  RouteMethod
	Path    string
	Handler http.Handler
}