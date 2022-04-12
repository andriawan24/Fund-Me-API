package transaction

import (
	"errors"
	"fund-me/campaign"
	"fund-me/payment"
	"strconv"
)

type service struct {
	repository         Repository
	campaignRepository campaign.Repository
	paymentService     payment.Service
}

type Service interface {
	GetTransactionsByCampaignID(input GetCampaignTransactionInput) ([]Transaction, error)
	GetTransactionsByUserID(userID int) ([]Transaction, error)
	ProccessPayment(input TransactionNotificationInput) error
	CreateTransaction(input CreateTransactionInput) (Transaction, error)
	GetAllTransactions() ([]Transaction, error)
}

func NewService(repository Repository, campaignRepository campaign.Repository, paymentService payment.Service) *service {
	return &service{repository, campaignRepository, paymentService}
}

func (s *service) GetTransactionsByCampaignID(input GetCampaignTransactionInput) ([]Transaction, error) {

	campaign, err := s.campaignRepository.FindByID(input.ID)
	if err != nil {
		return []Transaction{}, nil
	}

	if campaign.UserID != input.User.ID {
		return []Transaction{}, errors.New("not an owner of the campaign")
	}

	transactions, err := s.repository.GetByCampaignID(input.ID)
	if err != nil {
		return transactions, err
	}

	return transactions, nil
}

func (s *service) GetTransactionsByUserID(userID int) ([]Transaction, error) {
	transations, err := s.repository.GetByUserID(userID)
	if err != nil {
		return transations, err
	}

	return transations, nil
}

func (s *service) CreateTransaction(input CreateTransactionInput) (Transaction, error) {
	transaction := Transaction{}
	transaction.CampaignID = input.CampaignID
	transaction.Amount = input.Amount
	transaction.UserID = input.User.ID
	transaction.Status = "pending"

	newTransaction, err := s.repository.Save(transaction)
	if err != nil {
		return newTransaction, err
	}

	inputTransaction := payment.Transaction{
		Amount: newTransaction.Amount,
		ID:     newTransaction.ID,
	}

	paymentURL, err := s.paymentService.GetPaymentURL(inputTransaction, input.User)
	if err != nil {
		return newTransaction, err
	}

	newTransaction.PaymentURL = paymentURL

	newTransaction, err = s.repository.Update(newTransaction)
	if err != nil {
		return newTransaction, err
	}

	return newTransaction, nil
}

func (s *service) ProccessPayment(input TransactionNotificationInput) error {
	transaction_id, _ := strconv.Atoi(input.OrderID)

	transaction, err := s.repository.GetByID(transaction_id)
	if err != nil {
		return err
	}

	if input.PaymentType == "credit_card" && input.TransactionStatus == "capture" && input.FraudStatus == "accept" {
		transaction.Status = "paid"
	} else if input.TransactionStatus == "settlement" {
		transaction.Status = "paid"
	} else if input.TransactionStatus == "deny" || input.TransactionStatus == "expire" || input.TransactionStatus == "cancel" {
		transaction.Status = "CANCELLED"
	}

	updatedTransaction, err := s.repository.Update(transaction)
	if err != nil {
		return err
	}

	campaign, err := s.campaignRepository.FindByID(updatedTransaction.CampaignID)
	if err != nil {
		return err
	}

	if updatedTransaction.Status == "paid" {
		campaign.BackerCount += 1
		campaign.CurrentAmount += updatedTransaction.Amount

		_, err := s.campaignRepository.Update(campaign)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) GetAllTransactions() ([]Transaction, error) {
	transactions, err := s.repository.FindAll()
	if err != nil {
		return transactions, err
	}

	return transactions, nil
}
