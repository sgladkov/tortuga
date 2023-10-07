package main

import (
	"github.com/sgladkov/tortuga/internal/blockchain"
	"github.com/sgladkov/tortuga/internal/models"
	storage2 "github.com/sgladkov/tortuga/internal/storage"
	"github.com/sgladkov/tortuga/internal/utils"
)

func initStorage(config *utils.Config) (storage2.Storage, error) {
	var storage storage2.Storage
	var err error
	if config.DB != nil {
		storage, err = storage2.NewPgStorage(config.DB)
		if err != nil {
			return nil, err
		}
	} else {
		storage = storage2.NewTestStorage([]models.User{}, []models.Project{}, []models.Bid{}, []models.Rate{})
	}
	return storage, nil
}

func initAddress(config *utils.Config) (string, error) {
	pubKey, err := blockchain.PublicKeyFromPrivateKey(config.WalletKey)
	if err != nil {
		return "", err
	}
	address, err := blockchain.AddressFromPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	return address, nil
}
