//
// Copyright (C) 2019 Authlete, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific
// language governing permissions and limitations under the
// License.

package main

import (
	"fmt"
	"math/rand"

	"github.com/authlete/authlete-go-gin/endpoint"
	"github.com/authlete/authlete-go-gin/middleware"
	"github.com/authlete/authlete-go-gin/web"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

type AuthorizationServer struct {
	Engine *gin.Engine
}

func AuthorizationServer_New() *AuthorizationServer {
	server := AuthorizationServer{}
	server.init()

	return &server
}

func (self *AuthorizationServer) Run(addr ...string) error {
	return self.Engine.Run(addr...)
}

func (self *AuthorizationServer) init() {
	self.Engine = gin.Default()

	self.setupStatic()
	self.setupTemplates()
	self.setupSession()
	self.setupAuthleteApi()
	self.setupAuthorizationEndpoint(`/api/authorization`)
	self.setupAuthorizationDecisionEndpoint(`/api/authorization/decision`)
	self.setupDiscoveryEndpoint(`/.well-known/openid-configuration`)
	self.setupIntrospectionEndpoint(`/api/introspection`)
	self.setupJwksEndpoint(`/api/jwks`)
	self.setupRevocationEndpoint(`/api/revocation`)
	self.setupTokenEndpoint(`/api/token`)
}

func (self *AuthorizationServer) setupStatic() {
	self.Engine.Static(`css`, `./css`)
}

func (self *AuthorizationServer) setupTemplates() {
	self.Engine.LoadHTMLGlob("templates/*")
}

func (self *AuthorizationServer) setupSession() {
	// This implementation uses the memory as the store for sessions.
	// Change this as necessary.

	// Key for memstore
	key := make([]byte, 32)
	rand.Read(key)

	// Store for session
	store := memstore.NewStore(key)

	// Session for gin
	self.Engine.Use(sessions.Sessions("AuthorizationServerSession", store))
}

func (self *AuthorizationServer) setupAuthleteApi() {
	// Register middleware that creates an instance of api.AuthleteApi and
	// sets the instance to the given gin contxt with the key `AuthleteApi`.
	//
	// middleware.AuthleteApi_Toml(file string) loads settings from a TOML
	// file. middleware.AuthleteApi_Env() reads settings from the environment.
	// middleware.AuthleteApi_Conf(conf.AuthleteConfiguration) reads settings
	// from a given AuthleteConfiguration.
	//
	// The following code loads `authlete.toml`.
	self.Engine.Use(middleware.AuthleteApi_Toml(`authlete.toml`))
}

func (self *AuthorizationServer) setupAuthorizationEndpoint(path string) {
	handler := AuthorizationEndpoint_Handler()

	// Authorization endpoint (RFC 6749)
	self.Engine.GET(path, handler)
	self.Engine.POST(path, handler)
}

func (self *AuthorizationServer) setupAuthorizationDecisionEndpoint(path string) {
	// Authorization decision endpoint
	self.Engine.POST(path, AuthorizationDecisionEndpoint_Handler())
}

func (self *AuthorizationServer) setupDiscoveryEndpoint(path string) {
	// Discovery endpoint (OpenID Connect Discovery 1.0)
	self.Engine.GET(path, endpoint.DiscoveryEndpoint_Handler())
}

func (self *AuthorizationServer) setupIntrospectionEndpoint(path string) {
	// Function to authenticate the API caller.
	authenticate := authenticateFunc

	// Function to reject the introspection request.
	reject := func(ctx *gin.Context) {
		rejectFunc(ctx, path)
	}

	handler := endpoint.IntrospectionEndpoint_Handler(authenticate, reject)

	// Introspection endpoint (RFC 7662)
	self.Engine.POST(path, handler)
}

func (self *AuthorizationServer) setupJwksEndpoint(path string) {
	// JWK Set Document (RFC 7517)
	self.Engine.GET(path, endpoint.JwksEndpoint_Handler())
}

func (self *AuthorizationServer) setupRevocationEndpoint(path string) {
	// Revocation endpoint (RFC 7009)
	self.Engine.POST(path, endpoint.RevocationEndpoint_Handler())
}

func (self *AuthorizationServer) setupTokenEndpoint(path string) {
	// Token endpoint (RFC 6749)
	spi := TokenReqHandlerSpiImpl{}
	self.Engine.POST(path, endpoint.TokenEndpoint_Handler(&spi))
}

// NOTE: The following functions are for demonstration purposes only.

func authenticateFunc(ctx *gin.Context) bool {
	// Assuming that the request contains data for Basic Authentication.
	user, _, ok := ctx.Request.BasicAuth()

	// If the request does not contain data for Basic Authentication or
	// the user is "nobody".
	if !ok || user == `nobody` {
		// Reject the request.
		return false
	}

	// Accept anybody except "nobody" regardless of whatever the value
	// of the password is.
	return true
}

func rejectFunc(ctx *gin.Context, path string) {
	// 401 Unauthorized
	challenge := fmt.Sprintf(`Basic realm="%s"`, path)
	resutil := web.ResponseUtility{}
	resutil.Unauthorized(ctx, challenge)
}
