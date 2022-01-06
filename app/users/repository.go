package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"golang.org/x/crypto/bcrypt"
)

var bcryptCost = 14

type Repository interface {
	CleanupSessions(ctx context.Context) (int, error)
	CreateSession(ctx context.Context, userID string, session *Session) error
	CreateUser(ctx context.Context, user *CreateUserRequest) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type RepositoryNeo4j struct {
	session neo4j.Session
}

func NewRepositoryNeo4j(session neo4j.Session) *RepositoryNeo4j {
	return &RepositoryNeo4j{
		session: session,
	}
}

func (r *RepositoryNeo4j) CleanupSessions(ctx context.Context) (int, error) {
	res, err := r.session.Run(
		`MATCH (s:Session)
		WHERE s.expires <>> datetime({timezone: 'UTC'})
		DETACH DELETE s RETURN count(s) AS deleted`, map[string]interface{}{})

	if err != nil {
		return 0, err
	}

	record, err := res.Single()

	return int(record.Values[0].(int64)), err
}

func (r *RepositoryNeo4j) CreateSession(ctx context.Context, userID string, session *Session) error {
	cypher := `MATCH (u:User {uuid: $userID})
				CREATE (u)-[:HAS_SESSION]->(s:Session {
					uuid: $uuid,
					createdAt: $createdAt, 
					expires: $expires
				})`
	_, err := r.session.Run(cypher, map[string]interface{}{
		"userID":    userID,
		"uuid":      session.ID,
		"createdAt": session.CreatedAt.In(time.UTC),
		"expires":   session.Expires.In(time.UTC),
	})
	return err
}

func (r *RepositoryNeo4j) CreateUser(ctx context.Context, createUser *CreateUserRequest) (*User, error) {
	passwordHashData, err := bcrypt.GenerateFromPassword([]byte(createUser.Password), bcryptCost)
	if err != nil {
		return nil, err
	}
	passwordHash := string(passwordHashData)
	userID := uuid.New().String()

	cypher := "CREATE (u:User {uuid: $uuid, firstName: $firstName, lastName: $lastName, email: $email, password: $password})"
	_, err = r.session.Run(cypher, map[string]interface{}{
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

func (r *RepositoryNeo4j) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	cypher := "MATCH (u:User {email: $email}) RETURN u"
	result, err := r.session.Run(cypher, map[string]interface{}{
		"email": email,
	})
	if err != nil {
		return nil, err
	}

	record, err := result.Single()
	if err != nil {
		return nil, err
	}

	props := record.Values[0].(neo4j.Node).Props
	user := &User{
		ID:        props["uuid"].(string),
		FirstName: props["firstName"].(string),
		LastName:  props["lastName"].(string),
		Email:     props["email"].(string),
		Password:  props["password"].(string),
	}

	return user, nil
}
