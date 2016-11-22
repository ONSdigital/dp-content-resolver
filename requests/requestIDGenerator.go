package requests

import (
	"fmt"
	"math/rand"
	"net/http"
)

// RequestIDHeaderParam request header parameter name for unique request ID.
const RequestIDHeaderParam = "X-Request-Id"
const requestIDFmt = "%s-%v"

// ContextIDGenerator generates a unique request header value based on the in bound request X-Request-Id header
type ContextIDGenerator struct {
	originalRequest *http.Request
}

// NewContentIDGenerator creates a new ContextIDGenerator using the provided http.Request
func NewContentIDGenerator(req *http.Request) ContextIDGenerator {
	return ContextIDGenerator{originalRequest: req}
}

// Generate generates a unique RequestContextID to be used to communicate with zebedee.
func (r ContextIDGenerator) Generate() string {
	return fmt.Sprintf(requestIDFmt, r.originalRequest.Header.Get(RequestIDHeaderParam), rand.Int63n(1000))
}
