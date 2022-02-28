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
