package usecase

import (
	"errors"

	"github.com/the-go-dragons/final-project2/internal/domain"
	"github.com/the-go-dragons/final-project2/internal/interfaces/persistence"
)

type WalletService struct {
	walletRepo  persistence.WalletRepository
	paymentRepo persistence.PaymentRepository
	trxRepo     persistence.TransactionRepository
}

func NewWallet(
	walletRepo persistence.WalletRepository,
	paymentRepo persistence.PaymentRepository,
	trxRepo persistence.TransactionRepository,
) WalletService {
	return WalletService{
		walletRepo:  walletRepo,
		paymentRepo: paymentRepo,
		trxRepo:     trxRepo,
	}
}
func (w WalletService) ChargeRequest(walletId uint, amount uint64) (uint, error) {
	_, err := w.walletRepo.Get(walletId)
	if err != nil {
		return 0, WallertNotFound{walletId}
	}
	payment := domain.Payment{Amount: amount, WalletID: walletId}
	payment, err = w.paymentRepo.Create(payment)
	if err != nil {
		return 0, err
	}
	return payment.ID, nil
}

func (w WalletService) FinalizeCharge(paymentID int) (uint, error) {
	payment, err := w.paymentRepo.Get(paymentID)
	if err != nil {
		return 0, PaymentNotFound{paymentID}
	}
	if payment.Status == domain.UNPAID {
		return 0, PaymentNotPaid{paymentID}
	}
	if payment.Status == domain.APPLIED {
		return 0, PaymentAlreadyApplied{paymentID}
	}
	if payment.Status != domain.PAID {
		return 0, InvalidPaymentStatus{paymentID, payment.Status}
	}
	err = w.walletRepo.ChargeWallet(payment.WalletID, payment.Amount)
	if err != nil {
		return 0, err
	}
	transaction := domain.Transaction{
		Amount:   payment.Amount,
		WalletID: payment.WalletID,
		Status:   domain.DEPOSIT,
	}
	_, err = w.trxRepo.Create(transaction)
	if err != nil {
		return 0, err
	}
	payment.Status = domain.APPLIED
	_, err = w.paymentRepo.Update(payment)
	if err != nil {
		return 0, err
	}
	return uint(payment.WalletID), nil
}

func (w WalletService) GetByUserId(id uint) (domain.Wallet, error) {
	return w.walletRepo.GetByUserId(id)
}

func (ws WalletService) CheckTheWalletBalance(user domain.User, price uint) error {
	wallet, err := ws.walletRepo.GetByUserId(user.ID)
	if err != nil || wallet.ID == 0 {
		return errors.New("can't get the wallet")
	}
	if wallet.Balance < price {
		return errors.New("not enough balance")
	}
	return nil
}
