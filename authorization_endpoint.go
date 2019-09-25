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
	"github.com/authlete/authlete-go/api"
	"github.com/authlete/authlete-go/dto"
	"github.com/authlete/authlete-go/types"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type AuthorizationEndpoint struct {
	endpoint.BaseEndpoint
}

func AuthorizationEndpoint_Handler() gin.HandlerFunc {
	// Instance of authorization endpoint
	endpoint := AuthorizationEndpoint{}

	return func(ctx *gin.Context) {
		endpoint.Handle(ctx)
	}
}

func (self *AuthorizationEndpoint) Handle(ctx *gin.Context) {
	api := self.GetAuthleteApi(ctx)
	if api == nil {
		return
	}

	// Query parameters or form parameters. OIDC Core 1.0 requires that
	// the authorization endpoint support both GET and POST methods.
	params := self.ReqUtil.ExtractParams(ctx)

	// Call Authlete's /api/auth/authorization API.
	res, err := self.callAuthorizationApi(ctx, params)
	if err != nil {
		return
	}

	// 'action' in the response denotes the next action which this
	// authorization endpoint implementation should take.
	action := res.Action

	switch action {
	case dto.AuthorizationAction_INTERACTION:
		// Process the authorization request with user interaction.
		self.handleInteraction(ctx, res)
	case dto.AuthorizationAction_NO_INTERACTION:
		// Process the authorization request without user interaction.
		// This happens only when the authorization request contains
		// 'prompt=none'.
		self.handleNoInteraction(ctx, res)
	default:
		// Handle other error cases.
		self.handleError(ctx, res)
	}
}

func (self *AuthorizationEndpoint) callAuthorizationApi(
	ctx *gin.Context, params string) (
	res *dto.AuthorizationResponse, err *api.AuthleteError) {
	// Preapre a request for /api/auth/authorization API.
	req := dto.AuthorizationRequest{}
	req.Parameters = params

	// Call /api/auth/authorization API.
	res, err = self.Api.Authorization(&req)

	if err != nil {
		self.ResUtil.WithAuthleteError(ctx, err)
	}

	return
}

func (self *AuthorizationEndpoint) handleNoInteraction(
	ctx *gin.Context, res *dto.AuthorizationResponse) {
	// Processing the request with user interaction
	msg := "authorization_endpoint: Processing the request without user interaction."
	log.Debug().Msg(msg)

	// Let NoInteractionHandler handle the case of 'prompt=none'
	spi := NoInteractionHandlerSpiImpl{}
	spi.Init(ctx)
	handler := handler.NoInteractionHandler_New(self.Api, &spi)
	handler.Handle(ctx, res)
}

func (self *AuthorizationEndpoint) handleError(
	ctx *gin.Context, res *dto.AuthorizationResponse) {
	// The request caused an error
	msg := fmt.Sprintf("authorization_endpoint: The request caused an error: %s", res.ResultMessage)
	log.Debug().Msg(msg)

	// Let AuthReqErrorHandler handle the error case.
	handler := handler.AuthReqErrorHandler_New(self.Api)
	handler.Handle(ctx, res)
}

func (self *AuthorizationEndpoint) handleInteraction(
	ctx *gin.Context, res *dto.AuthorizationResponse) {
	// Processing the request with user interaction
	msg := "authorization_endpoint: Processing the request with user interaction."
	log.Debug().Msg(msg)

	// Session
	session := sessions.Default(ctx)

	// Prepare a model object which is used to render the authorization page.
	model := prepareModel(ctx, res, session)

	// 'model' is nil only when there is no use who has the required subject.
	if model == nil {
		reason := dto.AuthorizationFailReason_NOT_AUTHENTICATED
		self.authorizationFail(ctx, res.Ticket, reason)
		return
	}

	// Store some variables into the session so that they can be referred to
	// later in authorization_decision_endpoint.go.
	session.Set(`ticket`, res.Ticket)
	session.Set(`claimNames`, res.Claims)
	session.Set(`claimLocales`, res.ClaimsLocales)
	session.Save()

	// Render the authorization page.
	ctx.HTML(200, `authorization.html`, gin.H{"model": model})
}

func prepareModel(ctx *gin.Context, res *dto.AuthorizationResponse,
	session sessions.Session) *AuthorizationPageModel {
	// Model object used to render the authorization page.
	model := AuthorizationPageModel_New(res)

	// User in the session.
	user := getUserFromSession(session)

	// Check if login is required.
	model.LoginRequired = isLoginRequired(ctx, res, session, user)

	if model.LoginRequired == false {
		// The user's name that will be referred to in the authorization page.
		model.UserName = user.GivenName
		return model
	}

	// Logout the user.
	logoutUser(session)

	// If the authorization request does not require a specific 'subject'.
	if res.Subject == `` {
		// This simple implementation uses 'login_hint' as the initial value
		// of the login ID.
		model.LoginId = res.LoginHint
		return model
	}

	// The authorization request requires a specific 'subject' be used.

	// Try to find a user whose subject is equal to the required subject.
	user = UserDatabase_Get().GetBySubject(res.Subject)

	if user == nil {
		// There is no user who has the required subject.
		msg := "authorization_endpoint: The request fails because there is no user who has the required subject."
		log.Debug().Msg(msg)
		return nil
	}

	// The user who is identified by the subject exists.
	model.LoginId = user.LoginId
	model.LoginIdReadOnly = `readonly`

	return model
}

func getUserFromSession(session sessions.Session) *UserEntity {
	value := session.Get(`user`)

	if value == nil {
		return nil
	}

	bytes, _ := value.([]byte)

	user := UserEntity{}
	json.Unmarshal(bytes, &user)

	return &user
}

func logoutUser(session sessions.Session) {
	session.Delete(`user`)
	session.Delete(`authenciatedAt`)
}

func isLoginRequired(ctx *gin.Context, res *dto.AuthorizationResponse,
	session sessions.Session, user *UserEntity) bool {
	// If no user has logged in.
	if user == nil {
		return true
	}

	// Check if the 'prompt' parameter includes 'login'.
	included := isLoginIncludedInPrompt(res)
	if included {
		// Login is explicitly required by the client.
		// The user has to re-login.
		msg := "authorization_endpoint: Login is required because 'prompt' includes 'login'."
		log.Debug().Msg(msg)
		return true
	}

	// If the authorization request requires a specific subject.
	if res.Subject != `` {
		// If the current user's subject does not match the required one.
		if user.Subject != res.Subject {
			// The user needs to login with another user account.
			msg := "authorization_endpoint: Login is required because the current user's subject does not match the required one."
			log.Debug().Msg(msg)
			return true
		}
	}

	// Check if the max age has passed since the last time the user logged in.
	exceeded := isMaxAgeExceeded(res, session)
	if exceeded {
		// The user has to re-login.
		msg := "authorization_endpoint: Login is required because the max age has passed since the last login."
		log.Debug().Msg(msg)
		return true
	}

	// Login is not required.
	return false
}

func isLoginIncludedInPrompt(res *dto.AuthorizationResponse) bool {
	// If the authorization request does not include a 'prompt' parameter.
	if res.Prompts == nil {
		return false
	}

	// For each value in the 'prompt' parameter.
	for _, prompt := range res.Prompts {
		if prompt == types.Prompt_LOGIN {
			// 'login' is included in the 'prompt' parameter.
			return true
		}
	}

	// The 'prompt' parameter does not include 'login'.
	return false
}

func isMaxAgeExceeded(res *dto.AuthorizationResponse, session sessions.Session) bool {
	maxAge := uint64(res.MaxAge)

	// If the authorization request does not include a 'max_age' parameter
	// and the 'default_max_age' metadata of the client is not set.
	if maxAge == 0 {
		// Don't have to care about the maximum authentication age.
		return false
	}

	// The last time when the user was authenticated.
	authAt := uint64(0)
	value := session.Get(`authenticatedAt`)
	if value != nil {
		authAt, _ = value.(uint64)
	}

	// The current time.
	current := uint64(time.Now().Unix())

	// Calculate the number of seconds that have elapsed since the last login.
	age := current - authAt

	if age <= maxAge {
		// The max age is not exceeded yet.
		return false
	}

	// The max age has been exceeded.
	return true
}

func (self *AuthorizationEndpoint) authorizationFail(
	ctx *gin.Context, ticket string, reason dto.AuthorizationFailReason) {
	handler := handler.AuthReqBaseHandler{}
	handler.Init(self.Api)
	handler.AuthorizationFail(ctx, ticket, reason)
}
