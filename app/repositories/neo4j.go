package repositories

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/db"
	log "github.com/sirupsen/logrus"
)

const NeoErrNoRecordsMsg = "Result contains no more records"

func neo4jReadTxSingle(
	ctx context.Context,
	driver neo4j.Driver,
	query string,
	params map[string]interface{},
) (*db.Record, error) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() {
		err := session.Close()
		if err != nil {
			log.Errorf("failed to close session: %v", err)
		}
	}()

	result, err := session.ReadTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {
			res, err := session.Run(
				query, params,
			)

			if err != nil {
				return 0, err
			}

			return res.Single()
		},
	)

	if err != nil {
		return nil, err
	}

	return result.(*db.Record), nil
}

func neo4jWriteTxSingle(ctx context.Context, driver neo4j.Driver, query string) (
	*db.Record,
	error,
) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		err := session.Close()
		if err != nil {
			log.Errorf("failed to close session: %v", err)
		}
	}()

	result, err := session.WriteTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {
			res, err := session.Run(
				query, map[string]interface{}{},
			)

			if err != nil {
				return 0, err
			}

			return res.Single()
		},
	)

	if err != nil {
		return nil, err
	}

	return result.(*db.Record), nil
}
