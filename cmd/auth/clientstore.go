package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	"gorm.io/gorm"
)

type Inventory struct {
	oauth2.ClientInfo
	ctx       context.Context
	tableName string
	db        *gorm.DB
	stdout    io.Writer
}

type clientInfo struct {
}

type ClientInfoInter interface {
	GetID() string
	GetSecret() string
	GetDomain() string
	IsPublic() bool
	GetUserID() string
}

func NewInventory(ctx context.Context, db *gorm.DB) *Inventory {
	s, err := db.DB()
	if err != nil {
		panic(err)
	}
	s.SetMaxIdleConns(10)
	s.SetMaxOpenConns(100)
	s.SetConnMaxLifetime(time.Hour)

	store := &Inventory{
		db:        db,
		tableName: "oauth2_clients",
		stdout:    os.Stderr,
	}
	if !db.Migrator().HasTable(store.tableName) {
		if err := db.Table(store.tableName).Migrator().CreateTable(&User{}); err != nil {
			panic(err)
		}
	}

	return store
}

func NewClientInfo() *clientInfo {
	clientStore := &ClientInfo{}
	return clientStore
}

func (s *ClientInfo) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}
	var user User
	err := s.db.WithContext(ctx).Table(s.tableName).Limit(1).Find(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	// need adapter implemetation
	return ClientInfo{
		PublicID:     user.ID,
		PasswordHash: Passw,
		Domain:       ctx.Value("domain").(string),
	}, nil
}

func (s *Inventory) Create(ctx context.Context, info oauth2.ClientInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	item := &models.Client{
		ID:     generateNewID(),
		Secret: info.GetSecret(),
		Domain: info.GetDomain(),
		Data:   info.GetData(data),
	}

	return s.db.WithContext(ctx).Table(s.tableName).Create(item).Error
}

func (store.ClientStore) GetSecret()
