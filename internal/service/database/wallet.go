package database

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/wajidp/micro-payment-gateway/internal/logger"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

// UserWalletRepo is an in-memory implementation of the WalletRepository interface
// It stores user wallets and transactions in a thread-safe manner using a read-write mutex.
type UserWalletRepo struct {
	data         map[string]*model.Wallet      // data holds the wallet information for each user, keyed by user ID.
	transactions map[string]*model.Transaction // transactions holds transaction details for each transaction ID.
	mu           sync.RWMutex                  // mu is a read-write mutex used to ensure thread-safe access to the data.
}

// NewUserWalletRepo creates a new instance of UserWalletRepo and returns it as a WalletRepository.
// This repository stores wallet and transaction data in memory.
func NewUserWalletRepo() model.WalletRepository {
	return &UserWalletRepo{
		data:         make(map[string]*model.Wallet),
		transactions: make(map[string]*model.Transaction),
	}
}

// GetWallet retrieves the wallet for a given userID from the repository.
// If the wallet does not exist, it initializes a new one and stores it in the repository.
func (r *UserWalletRepo) GetWallet(userID string) (*model.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wallet, exists := r.data[userID]
	if !exists {
		wallet = &model.Wallet{}
		r.data[userID] = wallet
	}

	return wallet, nil
}

// UpdateWallet updates the wallet for a given userID in the repository.
// The function locks the repository for writing, updates the wallet, and logs the change.
func (r *UserWalletRepo) UpdateWallet(userID string, _wallet *model.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[userID] = _wallet

	jw, _ := json.Marshal(_wallet)
	logger.Infof("Wallet Update for User %s --> %v", userID, string(jw))

	return nil
}

// GetTransaction retrieves a transaction by its ID from the repository.
// If the transaction does not exist, it returns an error.
func (r *UserWalletRepo) GetTransaction(txnID string) (*model.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	txn, exists := r.transactions[txnID]
	if !exists {
		return nil, errors.New("transaction not found")
	}

	return txn, nil
}

// UpdateTransaction updates the transaction in the repository.
// The function locks the repository for writing, updates the transaction, and logs the change.
func (r *UserWalletRepo) UpdateTransaction(txn *model.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.transactions[txn.ID] = txn

	jw, _ := json.Marshal(txn)
	logger.Infof("Transaction Update for User %s --> %v", txn.UserID, string(jw))

	return nil
}
