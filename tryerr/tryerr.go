package tryerr

import "errors"

var (
	// ErrTenantExists is returned when the provided tenant already exists in the store
	ErrTenantExists = errors.New("tenant exists in db")
	// ErrIDNotNull is returned when an ID is provided on item creation
	ErrIDNotNull = errors.New("id provided and not expected")
	// ErrInvalidContext is returned when no context is provided or it is invalid
	ErrInvalidContext = errors.New("invalid context")
	// ErrStorerAlreadyRegistered is returned when the Storer has been previously registered
	ErrStorerAlreadyRegistered = errors.New("provider already registered")
	// ErrNilStore is returned when the provided store is nil, not a valid Storer instance
	ErrNilStore = errors.New("store cannot be nil")
	// ErrStoreNotRegistered is returned when trying to access a store that has not been registered
	ErrStoreNotRegistered = errors.New("store not registered")
	// ErrAccountNotProvided is returned when the a required account is needed and not provided
	ErrAccountNotProvided = errors.New("account not provided")
	// ErrAccountNotFound is returned when the required account was not found in the store
	ErrAccountNotFound = errors.New("account not found")
	// ErrEmailNotFound is returned when no account is found in the store with the given email
	ErrEmailNotFound = errors.New("email not found")
	// ErrDupEmail is returned when the provided email is already found in the store
	ErrDupEmail = errors.New("email exists in db")
	// ErrInvalidName notifies than the name is not valid
	ErrInvalidName = errors.New("invalid name")
	// ErrInvalidPassword alert about an invalid password
	ErrInvalidPassword = errors.New("invalid password")
	// ErrInvalidEmail notifies than the provided email is no valid
	ErrInvalidEmail = errors.New("invalid email address")
	// ErrKeyExists is returned when the provided key already exists in the store
	ErrKeyExists = errors.New("key exists in db")
	// ErrKeyNotFound is returned when the requested key is not found in the store
	ErrKeyNotFound = errors.New("key not found")
	// ErrNilKey is returned when the provided key is nil
	ErrNilKey = errors.New("key is nil")
	// ErrNilUID is returned when the provided key is empty
	ErrNilUID = errors.New("user id is empty")
	// ErrJWTWrongSigningMethod is returned when the provided token signing method do not match the one expected
	ErrJWTWrongSigningMethod = errors.New("unexpected signing method")
	// ErrTokenNotFound is returned when the requested token is not found in the store
	ErrTokenNotFound = errors.New("token not found")
	// ErrInvalidToken  is returned when the token is not a JWT valid token
	ErrInvalidToken = errors.New("token not valid")
	// ErrNilToken is returned when the provided token is nil
	ErrNilToken = errors.New("token is nil")
	// ErrUnauthorized is returned if the provided token is not authorized to access the resource
	ErrUnauthorized = errors.New("unauthorized token")
	// ErrRbacRoleNotFound is returned when the role is not found
	ErrRbacRoleNotFound = errors.New("RBAC role not found")
	// ErrRbacPermissionNotFound is returned when the permission is not found
	ErrRbacPermissionNotFound = errors.New("RBAC permission not found")
	// ErrRbacUserNotProvided is returned when the affected user id is not provided
	ErrRbacUserNotProvided = errors.New("RBAC user not found")
	// ErrNotImplemented is returned when the functionality required is not implemented
	ErrNotImplemented = errors.New("function not implemented")
)
