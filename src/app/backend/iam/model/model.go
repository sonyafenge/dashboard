// Copyright 2020 Authors of Arktos.
// Copyright 2020 Authors of Arktos - file modified.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"github.com/kubernetes/dashboard/src/app/backend/api"
	"time"
)

// User schema of the user table
type User struct {
	ID                int64     `json:"id"`
	Username          string    `json:"name"`
	Password          string    `json:"password"`
	Token             string    `json:"token"`
	Type              string    `json:"type"`
	Tenant            string    `json:"tenant"`
	Role              string    `json:"role"`
	NameSpace         string    `json:"namespace"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

type UserDetails struct {
	ObjectMeta User         `json:"objectMeta"`
	TypeMeta   api.TypeMeta `json:"typeMeta"`
	Phase      string       `json:"phase"`
}

type Token struct {
	Token string `json:"token"`
}

type UserList struct {
	ListMeta api.ListMeta `json:"listMeta"`

	// Unordered list of Users.
	Users []UserDetails `json:"users"`

	// List of non-critical errors, that occurred during resource retrieval.
	Errors []error `json:"errors"`
}
