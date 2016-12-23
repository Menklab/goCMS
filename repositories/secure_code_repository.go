package repositories

import (
	"time"
	"github.com/menklab/goCMS/models"
	"github.com/menklab/goCMS/database"
	"github.com/jmoiron/sqlx"
	"log"
)

type ISecureCodeRepository interface {
	Add(*models.SecureCode) error
	Delete(int) error
	GetLatestForUserByType(int, models.SecureCodeType) (*models.SecureCode, error)
}

type SecureCodeRepository struct {
	database *sqlx.DB
}


func DefaultSecureCodeRepository(db *database.Database) *SecureCodeRepository{
	secureCodeRepository := &SecureCodeRepository{
		database: db.Dbx,
	}

	return secureCodeRepository
}

func (scr *SecureCodeRepository) Add(code *models.SecureCode) error {
	code.Created = time.Now()
	// insert row
	result, err := scr.database.NamedExec(`
	INSERT INTO gocms_secure_codes (userId, type, code, created) VALUES (:userId, :type, :code, :created)
	`, code)
	if err != nil {
		log.Printf("Error adding security code to database: %s", err.Error())
		return err
	}

	// add id to user object
	id, _ := result.LastInsertId()
	code.Id = int(id)

	return nil
}

func (scr *SecureCodeRepository) Delete(id int) error {
	_, err := scr.database.Exec(`
	DELETE FROM gocms_secure_codes WHERE id=?
	`, id)
	if err != nil {
		log.Printf("Error deleting security code from database: %s", err.Error())
		return err
	}

	return nil
}

// get all events
func (scr *SecureCodeRepository) GetLatestForUserByType(id int, codeType models.SecureCodeType) (*models.SecureCode, error) {
	var secureCode models.SecureCode
	err := scr.database.Get(&secureCode, `
	SELECT * from gocms_secure_codes WHERE userId=? AND type=? ORDER BY created DESC LIMIT 1
	`, id, codeType)
	if err != nil {
		log.Printf("Error getting getting latest security code for user from database: %s", err.Error())
		return nil, err
	}
	return &secureCode, nil
}
