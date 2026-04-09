package postgres

import "database/sql"

type BootstrapStore struct {
	db *sql.DB
}

func NewBootstrapStore(db *sql.DB) BootstrapStore {
	return BootstrapStore{db: db}
}

type InboundReceiptStore struct {
	db *sql.DB
}

func NewInboundReceiptStore(db *sql.DB) InboundReceiptStore {
	return InboundReceiptStore{db: db}
}
