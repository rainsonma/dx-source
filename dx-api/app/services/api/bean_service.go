package api

import (
	"fmt"

	"dx-api/app/models"

	"github.com/goravel/framework/facades"

	"github.com/goravel/framework/contracts/database/orm"
)

// ConsumeBeans debits beans from a user and creates a ledger entry.
func ConsumeBeans(userID string, amount int, slug, reason string) error {
	if amount <= 0 {
		return nil
	}

	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// Lock row to prevent concurrent balance race
	var user models.User
	if err := tx.Where("id", userID).LockForUpdate().First(&user); err != nil || user.ID == "" {
		_ = tx.Rollback()
		return ErrUserNotFound
	}

	if user.Beans < amount {
		_ = tx.Rollback()
		return ErrInsufficientBeans
	}

	newBalance := user.Beans - amount

	if _, err := tx.Model(&models.User{}).Where("id", userID).
		Update("beans", newBalance); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to update bean balance: %w", err)
	}

	ledger := models.UserBean{
		ID:     newID(),
		UserID: userID,
		Beans:  -amount,
		Origin: user.Beans,
		Result: newBalance,
		Slug:   slug,
		Reason: reason,
	}
	if err := tx.Create(&ledger); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to create bean ledger entry: %w", err)
	}

	return tx.Commit()
}

// ConsumeBeansInTx debits beans within an existing transaction.
func ConsumeBeansInTx(tx orm.Query, userID string, amount int, slug, reason string) error {
	if amount <= 0 {
		return nil
	}

	var user models.User
	if err := tx.Where("id", userID).LockForUpdate().First(&user); err != nil || user.ID == "" {
		return ErrUserNotFound
	}

	if user.Beans < amount {
		return ErrInsufficientBeans
	}

	newBalance := user.Beans - amount

	if _, err := tx.Model(&models.User{}).Where("id", userID).
		Update("beans", newBalance); err != nil {
		return fmt.Errorf("failed to update bean balance: %w", err)
	}

	ledger := models.UserBean{
		ID:     newID(),
		UserID: userID,
		Beans:  -amount,
		Origin: user.Beans,
		Result: newBalance,
		Slug:   slug,
		Reason: reason,
	}
	return tx.Create(&ledger)
}

// RefundBeans credits beans back to a user and creates a ledger entry.
func RefundBeans(userID string, amount int, slug, reason string) error {
	if amount <= 0 {
		return nil
	}

	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.Where("id", userID).LockForUpdate().First(&user); err != nil || user.ID == "" {
		_ = tx.Rollback()
		return ErrUserNotFound
	}

	newBalance := user.Beans + amount

	if _, err := tx.Model(&models.User{}).Where("id", userID).
		Update("beans", newBalance); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to update bean balance: %w", err)
	}

	ledger := models.UserBean{
		ID:     newID(),
		UserID: userID,
		Beans:  amount,
		Origin: user.Beans,
		Result: newBalance,
		Slug:   slug,
		Reason: reason,
	}
	if err := tx.Create(&ledger); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to create bean ledger entry: %w", err)
	}

	return tx.Commit()
}

// GetBeanBalance returns the current bean balance for a user.
func GetBeanBalance(userID string) (int, error) {
	var user models.User
	if err := facades.Orm().Query().Where("id", userID).First(&user); err != nil || user.ID == "" {
		return 0, ErrUserNotFound
	}
	return user.Beans, nil
}
