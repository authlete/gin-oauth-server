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
	"encoding/json"
	"fmt"
	"time"

	"github.com/authlete/authlete-go-gin/endpoint"
	"github.com/authlete/authlete-go-gin/handler"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AuthorizationDecisionEndpoint struct {
	endpoint.BaseEndpoint
}

func AuthorizationDecisionEndpoint_Handler() gin.HandlerFunc {
	// Instance of authorization decision endpoint
	endpoint := AuthorizationDecisionEndpoint{}

	return func(ctx *gin.Context) {
		endpoint.Handle(ctx)
	}
}

func (self *AuthorizationDecisionEndpoint) Handle(ctx *gin.Context) {
	api := self.GetAuthleteApi(ctx)
	if api == nil {
		return
	}

	// Session
	session := sessions.Default(ctx)

	// Authenticate the user if necessary.
	authenticateUserIfNecessary(ctx, session)

	// Flag which indicates whether the user has given authorization
	// to the client application or not.
	authorized := isClientAuthorized(ctx)

	// Process the authorization request according to the user's decision.
	self.handleDecision(ctx, session, authorized)
}

func authenticateUserIfNecessary(ctx *gin.Context, session sessions.Session) {
	if session.Get(`user`) != nil {
		// The user has already logged in.
		return
	}

	// Credentials that the user input in the login form.
	loginId := ctx.PostForm(`loginId`)
	password := ctx.PostForm(`password`)

	// Authenticate the user.
	user := UserDatabase_Get().GetByCredentials(loginId, password)

	if user == nil {
		// User authentication failed.
		msg := fmt.Sprintf("authorization_decision_endpoint: User authentication failed. The presented login ID is '%s'.", loginId)
		log.Debug().Msg(msg)
		return
	}

	// User authentication succeeded.
	msg := fmt.Sprintf("authorization_decision_endpoint: User authentication succeeded. The presented login ID is '%s'.", loginId)
	log.Debug().Msg(msg)

	// Let the user log in.
	loginUser(session, user)
}

func loginUser(session sessions.Session, user *UserEntity) {
	// The current time.
	current := uint64(time.Now().Unix())

	// Convert 'user' into JSON.
	bytes, _ := json.Marshal(user)

	session.Set(`user`, bytes)
	session.Set(`authenticatedAt`, current)
	session.Save()
}

func isClientAuthorized(ctx *gin.Context) bool {
	authorized := ctx.PostForm(`authorized`)

	return authorized != ``
}

func (self *AuthorizationDecisionEndpoint) handleDecision(
	ctx *gin.Context, session sessions.Session, authorized bool) {
	spi := AuthReqDecisionHandlerSpiImpl_New(ctx, authorized)
	handler := handler.AuthReqDecisionHandler_New(self.Api, spi)

	// Parameters contained in the response from /api/auth/authorization API.
	value := session.Get(`ticket`)
	ticket, _ := value.(string)

	value = session.Get(`claimNames`)
	claimNames, _ := value.([]string)

	value = session.Get(`claimLocales`)
	claimLocales, _ := value.([]string)

	handler.Handle(ctx, ticket, claimNames, claimLocales)
}
