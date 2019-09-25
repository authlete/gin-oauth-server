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

// NOTE: THIS IS A DUMMY IMPLEMENTATION JUST FOR DEMONSTRATION

import (
	"fmt"

	"github.com/authlete/authlete-go/dto"
	"github.com/authlete/authlete-go/types"
)

var (
	userDatabaseInstance *UserDatabase
)

func init() {
	db := UserDatabase{}
	db.Users = []UserEntity{
		UserEntity{
			Subject:     `1001`,
			LoginId:     `john`,
			Password:    `john`,
			GivenName:   `John`,
			FamilyName:  `Smith`,
			Email:       `john@example.com`,
			PhoneNumber: `+1 (425) 555-1212`,
			Address: dto.Address{
				Country: `USA`,
			},
		},
		UserEntity{
			Subject:     `1002`,
			LoginId:     `jane`,
			Password:    `jane`,
			GivenName:   `Jane`,
			FamilyName:  `Smith`,
			Email:       `jane@example.com`,
			PhoneNumber: `+56 (2) 687 2400`,
			Address: dto.Address{
				Country: `Chile`,
			},
		},
	}

	userDatabaseInstance = &db
}

type UserDatabase struct {
	Users []UserEntity
}

func UserDatabase_Get() *UserDatabase {
	return userDatabaseInstance
}

func (self *UserDatabase) GetByCredentials(loginId string, password string) *UserEntity {
	for _, entity := range self.Users {
		if entity.LoginId != loginId {
			continue
		}

		if entity.Password != password {
			return nil
		}

		return &entity
	}

	return nil
}

func (self *UserDatabase) GetBySubject(subject string) *UserEntity {
	for _, entity := range self.Users {
		if entity.Subject != subject {
			continue
		}

		return &entity
	}

	return nil
}

type UserEntity struct {
	Subject     string
	LoginId     string
	Password    string
	GivenName   string
	FamilyName  string
	Email       string
	PhoneNumber string
	Address     dto.Address
}

func (self *UserEntity) GetClaim(claimName string, languageTag string) interface{} {
	if claimName == `` {
		return nil
	}

	if languageTag != `` {
		return nil
	}

	// See "OpenID Connect Core 1.0, 5. Claims"
	switch claimName {
	case types.CLAIM_NAME:
		return fmt.Sprintf("%s %s", self.GivenName, self.FamilyName)
	case types.CLAIM_GIVEN_NAME:
		return self.GivenName
	case types.CLAIM_FAMILY_NAME:
		return self.FamilyName
	case types.CLAIM_EMAIL:
		return self.Email
	case types.CLAIM_PHONE_NUMBER:
		return self.PhoneNumber
	case types.CLAIM_ADDRESS:
		return &self.Address
	default:
		return nil
	}
}
