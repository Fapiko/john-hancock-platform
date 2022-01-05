package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"golang.org/x/crypto/bcrypt"
)

var bcryptCost = 14

type Repository interface {
	CreateUser(ctx context.Context, user *CreateUserRequest) (*User, error)
}

type RepositoryNeo4j struct {
	session neo4j.Session
}

func NewRepositoryNeo4j(session neo4j.Session) *RepositoryNeo4j {
	return &RepositoryNeo4j{
		session: session,
	}
}

func (u RepositoryNeo4j) CreateUser(ctx context.Context, createUser *CreateUserRequest) (*User, error) {
	passwordHashData, err := bcrypt.GenerateFromPassword([]byte(createUser.Password), bcryptCost)
	if err != nil {
		return nil, err
	}
	passwordHash := string(passwordHashData)
	userID := uuid.New().String()

	cypher := "CREATE (u:User {uuid: $uuid, firstName: $firstName, lastName: $lastName, email: $email, password: $password})"
	_, err = u.session.Run(cypher, map[string]interface{}{
		"uuid":      userID,
		"firstName": createUser.FirstName,
		"lastName":  createUser.LastName,
		"email":     createUser.Email,
		"password":  passwordHash,
	})
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:        userID,
		FirstName: createUser.FirstName,
		LastName:  createUser.LastName,
		Email:     createUser.Email,
		Password:  passwordHash,
	}
	return user, nil
}
