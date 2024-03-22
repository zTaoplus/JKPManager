package storage

import "zjuici.com/tablegpt/jkpmanager/src/models"

type SessionClient interface {
	GetSessions() ([]*models.Session, error)
	GetSessionByID(id string) (*models.Session, error)
	DeleteSessionByIDS(ids []string) (int64, error)
	SaveSession(id string, sessionJson []byte) error
}
