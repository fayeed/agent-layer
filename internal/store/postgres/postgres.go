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

type ReadStore struct {
	db *sql.DB
}

func NewReadStore(db *sql.DB) ReadStore {
	return ReadStore{db: db}
}

type WebhookDeliveryStore struct {
	db *sql.DB
}

func NewWebhookDeliveryStore(db *sql.DB) WebhookDeliveryStore {
	return WebhookDeliveryStore{db: db}
}

type ContactMemoryStore struct {
	db *sql.DB
}

func NewContactMemoryStore(db *sql.DB) ContactMemoryStore {
	return ContactMemoryStore{db: db}
}

type ContactStore struct {
	db *sql.DB
}

func NewContactStore(db *sql.DB) ContactStore {
	return ContactStore{db: db}
}

type ThreadStore struct {
	db *sql.DB
}

func NewThreadStore(db *sql.DB) ThreadStore {
	return ThreadStore{db: db}
}

type MessageStore struct {
	db *sql.DB
}

func NewMessageStore(db *sql.DB) MessageStore {
	return MessageStore{db: db}
}
