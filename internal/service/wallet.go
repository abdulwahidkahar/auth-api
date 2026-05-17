package service

import (
	"auth-api/internal/model"
	"auth-api/internal/repository"
	"context"
	"database/sql"
	"errors"
	"log/slog"
)

type WalletService struct {
	db         *sql.DB
	walletRepo *repository.WalletRepository
}

func NewWalletService(db *sql.DB, walletRepo *repository.WalletRepository) *WalletService {
	return &WalletService{
		db:         db,
		walletRepo: walletRepo,
	}
}

func (s *WalletService) TopUp(ctx context.Context, userID int, amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Failed to begin transaction for top-up", "error", err, "user_id", userID, "amount", amount)
		return err
	}
	defer tx.Rollback()

	err = s.walletRepo.UpdateBalanceTx(ctx, tx, userID, amount)
	if err != nil {
		slog.Error("Failed to update wallet balance", "error", err, "user_id", userID, "amount", amount)
		return err
	}

	if err := s.walletRepo.CreateTopUpHistoryTx(ctx, tx, userID, amount); err != nil {
		slog.Error("Failed to create top-up history", "error", err, "user_id", userID, "amount", amount)
		return err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit top-up transaction", "error", err, "user_id", userID, "amount", amount)
		return err
	}

	slog.Info("Wallet topped up successfully", "user_id", userID, "amount", amount)

	return nil
}

func (s *WalletService) GetBalance(ctx context.Context, userID int) (int64, error) {
	balance, err := s.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (s *WalletService) GetWallet(ctx context.Context, userID int) (model.WalletResponse, error) {
	wallet, err := s.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return model.WalletResponse{}, err
	}

	return wallet, nil
}

func (s *WalletService) Transfer(ctx context.Context, fromUserID, toWalletID int, amount int64) (model.TransferResponse, error) {
	if amount <= 0 {
		return model.TransferResponse{}, errors.New("amount must be greater than 0")
	}

	fromWallet, err := s.walletRepo.GetWalletByUserID(ctx, fromUserID)
	if err == sql.ErrNoRows {
		slog.Error("Sender wallet not found", "error", err, "from_user_id", fromUserID, "to_wallet_id", toWalletID, "amount", amount)
		return model.TransferResponse{}, errors.New("sender wallet not found")
	}
	if err != nil {
		slog.Error("Failed to fetch sender wallet", "error", err, "from_user_id", fromUserID, "to_wallet_id", toWalletID, "amount", amount)
		return model.TransferResponse{}, err
	}

	if fromWallet.ID == toWalletID {
		slog.Error("Transfer to the same wallet is not allowed", "from_user_id", fromUserID, "to_wallet_id", toWalletID, "amount", amount)
		return model.TransferResponse{}, errors.New("cannot transfer to the same wallet")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Failed to begin transaction for transfer", "error", err, "from_user_id", fromUserID, "to_wallet_id", toWalletID, "amount", amount)
		return model.TransferResponse{}, err
	}
	defer tx.Rollback()

	transfer, err := s.walletRepo.TransferTx(ctx, tx, fromWallet.ID, toWalletID, amount)
	if err != nil {
		slog.Error("Transfer failed", "error", err, "from_user_id", fromUserID, "to_wallet_id", toWalletID, "amount", amount)
		return model.TransferResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit transfer transaction", "error", err, "from_user_id", fromUserID, "to_wallet_id", toWalletID, "amount", amount)
		return model.TransferResponse{}, err
	}

	slog.Info("Transfer successful", "from_user_id", fromUserID, "to_wallet_id", toWalletID, "amount", amount)

	return transfer, nil
}

func (s *WalletService) GetHistoryTransfer(ctx context.Context, userID int, page, limit int) ([]model.TransferHistory, error) {
	history, err := s.walletRepo.TransferHistory(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func (s *WalletService) GetHistoryTopUp(ctx context.Context, userID int, page, limit int) ([]model.TopUpHistory, error) {
	history, err := s.walletRepo.TopUpHistory(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	return history, nil
}
