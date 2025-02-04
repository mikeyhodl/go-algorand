// Package private provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/algorand/oapi-codegen DO NOT EDIT.
package private

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	. "github.com/algorand/go-algorand/daemon/algod/api/server/v2/generated/model"
	"github.com/algorand/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Return a list of participation keys
	// (GET /v2/participation)
	GetParticipationKeys(ctx echo.Context) error
	// Add a participation key to the node
	// (POST /v2/participation)
	AddParticipationKey(ctx echo.Context) error
	// Delete a given participation key by ID
	// (DELETE /v2/participation/{participation-id})
	DeleteParticipationKeyByID(ctx echo.Context, participationId string) error
	// Get participation key info given a participation ID
	// (GET /v2/participation/{participation-id})
	GetParticipationKeyByID(ctx echo.Context, participationId string) error
	// Append state proof keys to a participation key
	// (POST /v2/participation/{participation-id})
	AppendKeys(ctx echo.Context, participationId string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetParticipationKeys converts echo context to params.
func (w *ServerInterfaceWrapper) GetParticipationKeys(ctx echo.Context) error {
	var err error

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetParticipationKeys(ctx)
	return err
}

// AddParticipationKey converts echo context to params.
func (w *ServerInterfaceWrapper) AddParticipationKey(ctx echo.Context) error {
	var err error

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.AddParticipationKey(ctx)
	return err
}

// DeleteParticipationKeyByID converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteParticipationKeyByID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "participation-id" -------------
	var participationId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "participation-id", runtime.ParamLocationPath, ctx.Param("participation-id"), &participationId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter participation-id: %s", err))
	}

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteParticipationKeyByID(ctx, participationId)
	return err
}

// GetParticipationKeyByID converts echo context to params.
func (w *ServerInterfaceWrapper) GetParticipationKeyByID(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "participation-id" -------------
	var participationId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "participation-id", runtime.ParamLocationPath, ctx.Param("participation-id"), &participationId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter participation-id: %s", err))
	}

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetParticipationKeyByID(ctx, participationId)
	return err
}

// AppendKeys converts echo context to params.
func (w *ServerInterfaceWrapper) AppendKeys(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "participation-id" -------------
	var participationId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "participation-id", runtime.ParamLocationPath, ctx.Param("participation-id"), &participationId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter participation-id: %s", err))
	}

	ctx.Set(Api_keyScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.AppendKeys(ctx, participationId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface, m ...echo.MiddlewareFunc) {
	RegisterHandlersWithBaseURL(router, si, "", m...)
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string, m ...echo.MiddlewareFunc) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/v2/participation", wrapper.GetParticipationKeys, m...)
	router.POST(baseURL+"/v2/participation", wrapper.AddParticipationKey, m...)
	router.DELETE(baseURL+"/v2/participation/:participation-id", wrapper.DeleteParticipationKeyByID, m...)
	router.GET(baseURL+"/v2/participation/:participation-id", wrapper.GetParticipationKeyByID, m...)
	router.POST(baseURL+"/v2/participation/:participation-id", wrapper.AppendKeys, m...)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"H7/Ue/y1vgzcXl99ufr/AQAA//+KH5i1MrgAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
