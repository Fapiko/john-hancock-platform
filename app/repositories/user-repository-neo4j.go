package repositories

import (
	"context"
	"time"

	"github.com/fapiko/john-hancock-platform/app/contracts"
	"github.com/fapiko/john-hancock-platform/app/repositories/daos"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"golang.org/x/crypto/bcrypt"
)

var _ UserRepository = (*UserRepositoryNeo4j)(nil)

type UserRepositoryNeo4j struct {
	driver neo4j.Driver
}

func NewUserRepositoryNeo4j(driver neo4j.Driver) *UserRepositoryNeo4j {
	return &UserRepositoryNeo4j{
		driver: driver,
	}
}

func (r *UserRepositoryNeo4j) CleanupSessions(ctx context.Context) (int, error) {
	query := `MATCH (s:Session)
		WHERE s.expires <= datetime({timezone: 'UTC'})
		DETACH DELETE s RETURN count(s) AS deleted`

	record, err := neo4jWriteTxSingle(ctx, r.driver, query)

	return int(record.Values[0].(int64)), err
}

func (r *UserRepositoryNeo4j) CreateSession(
	ctx context.Context,
	userID string,
) (*contracts.SessionResponse, error) {
	sessionResp := &contracts.SessionResponse{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		Expires:   time.Now().Add(sessionExpiration),
	}

	session := r.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	cypher := `MATCH (u:User {uuid: $userID})
				CREATE (u)-[:HAS_SESSION]->(s:Session {
					uuid: $uuid,
					createdAt: $createdAt, 
					expires: $expires
				})`
	_, err := session.Run(
		cypher, map[string]interface{}{
			"userID":    userID,
			"uuid":      sessionResp.ID,
			"createdAt": sessionResp.CreatedAt.In(time.UTC),
			"expires":   sessionResp.Expires.In(time.UTC),
		},
	)
	return sessionResp, err
}

func (r *UserRepositoryNeo4j) CreateUser(
	ctx context.Context,
	createUser *contracts.CreateUserRequest,
) (*daos.User, error) {
	passwordHashData, err := bcrypt.GenerateFromPassword([]byte(createUser.Password), bcryptCost)
	if err != nil {
		return nil, err
	}
	passwordHash := string(passwordHashData)
	userID := uuid.New().String()

	session := r.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	cypher := "CREATE (u:User {uuid: $uuid, firstName: $firstName, lastName: $lastName, email: $email, password: $password})"
	_, err = session.Run(
		cypher, map[string]interface{}{
			"uuid":      userID,
			"firstName": createUser.FirstName,
			"lastName":  createUser.LastName,
			"email":     createUser.Email,
			"password":  passwordHash,
		},
	)
	if err != nil {
		return nil, err
	}

	user := &daos.User{
		ID:        userID,
		FirstName: createUser.FirstName,
		LastName:  createUser.LastName,
		Email:     createUser.Email,
		Password:  passwordHash,
	}
	return user, nil
}

func (r *UserRepositoryNeo4j) GetUserByEmail(ctx context.Context, email string) (
	*daos.User,
	error,
) {
	session := r.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	cypher := "MATCH (u:User {email: $email}) RETURN u"
	result, err := session.Run(
		cypher, map[string]interface{}{
			"email": email,
		},
	)
	if err != nil {
		return nil, err
	}

	record, err := result.Single()
	if err != nil {
		return nil, err
	}

	props := record.Values[0].(neo4j.Node).Props
	user := daos.NewUserFromProps(props)

	return user, nil
}

func (r *UserRepositoryNeo4j) GetUserBySessionID(ctx context.Context, sessionID string) (
	*daos.User,
	error,
) {
	cypher := `MATCH (u:User)-[:HAS_SESSION]->(s:Session {uuid: $sessionID}) RETURN u`
	params := map[string]interface{}{
		paramSessionID: sessionID,
	}

	result, err := neo4jReadTxSingle(ctx, r.driver, cypher, params)
	if err != nil {
		return nil, err
	}

	return daos.NewUserFromProps(result.Values[0].(neo4j.Node).Props), nil
}
