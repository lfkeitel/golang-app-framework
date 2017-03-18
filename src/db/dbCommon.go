package db

import (
	"errors"
	"time"

	"github.com/lfkeitel/golang-app-framework/src/utils"
	"github.com/lfkeitel/verbose"
)

const dbVersion = 2

type dbInit interface {
	init(*utils.DatabaseAccessor, *utils.Config) error
}

var dbInits = make(map[string]dbInit)

func RegisterDatabaseAccessor(name string, db dbInit) {
	dbInits[name] = db
}

func NewDatabaseAccessor(e *utils.Environment) (*utils.DatabaseAccessor, error) {
	da := &utils.DatabaseAccessor{}
	if f, ok := dbInits[e.Config.Database.Type]; ok {
		var err error
		retries := 0
		dur, err := time.ParseDuration(e.Config.Database.RetryTimeout)
		if err != nil {
			return nil, errors.New("Invalid RetryTimeout")
		}

		// This loop will break when no error occurs when connecting to a database
		// Or when the number of attempted retries is greater than configured
		shutdownChan := e.SubscribeShutdown()

		for {
			err = f.init(da, e.Config)

			// If no error occurred, break
			// If an error occurred but retries is not set to inifinite and we've tried
			// too many times already, break
			if err == nil || (e.Config.Database.Retry != 0 && retries >= e.Config.Database.Retry) {
				break
			}

			retries++
			e.Log.WithFields(verbose.Fields{
				"Attempts":    retries,
				"MaxAttempts": e.Config.Database.Retry,
				"Timeout":     e.Config.Database.RetryTimeout,
				"Error":       err,
			}).Error("Failed to connect to database. Retrying after timeout.")

			select {
			case <-shutdownChan:
				return nil, err
			case <-time.After(dur):
			}
		}

		return da, err
	}
	return nil, errors.New("Database " + e.Config.Database.Type + " not supported")
}
