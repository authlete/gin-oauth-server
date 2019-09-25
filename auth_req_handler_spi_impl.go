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

	"github.com/authlete/authlete-go-gin/handler/spi"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthReqHandlerSpiImpl struct {
	spi.AuthReqHandlerSpiAdapter
	Context *gin.Context
	session sessions.Session
	user    *UserEntity
	tried   bool
}

func (self *AuthReqHandlerSpiImpl) Init(ctx *gin.Context) {
	self.Context = ctx
	self.session = sessions.Default(ctx)
}

func (self *AuthReqHandlerSpiImpl) GetUserClaimValue(
	subject string, claimName string, languageTag string) interface{} {
	user := self.getUserBySubject(subject)

	if user == nil {
		return nil
	}

	return user.GetClaim(claimName, languageTag)
}

func (self *AuthReqHandlerSpiImpl) GetUserAuthenticatedAt() uint64 {
	if self.session.Get(`user`) == nil {
		return 0
	}

	value := self.session.Get(`authenticatedAt`)
	authAt, _ := value.(uint64)

	return authAt
}

func (self *AuthReqHandlerSpiImpl) GetUserSubject() string {
	value := self.session.Get(`user`)

	if value == nil {
		return ``
	}

	bytes, _ := value.([]byte)

	user := UserEntity{}
	json.Unmarshal(bytes, &user)

	return user.Subject
}

func (self *AuthReqHandlerSpiImpl) getUserBySubject(subject string) *UserEntity {
	if self.tried == false {
		self.user = UserDatabase_Get().GetBySubject(subject)
		self.tried = true
	}

	return self.user
}
