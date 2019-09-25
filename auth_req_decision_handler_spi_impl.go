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
	"github.com/gin-gonic/gin"
)

type AuthReqDecisionHandlerSpiImpl struct {
	AuthReqHandlerSpiImpl
	Authorized bool
}

func (self *AuthReqDecisionHandlerSpiImpl) InitWithDecision(ctx *gin.Context, authorized bool) {
	self.Init(ctx)
	self.Authorized = authorized
}

func AuthReqDecisionHandlerSpiImpl_New(ctx *gin.Context, authorized bool) *AuthReqDecisionHandlerSpiImpl {
	impl := AuthReqDecisionHandlerSpiImpl{}
	impl.InitWithDecision(ctx, authorized)

	return &impl
}

func (self *AuthReqDecisionHandlerSpiImpl) IsClientAuthorized() bool {
	return self.Authorized
}
