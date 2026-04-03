package api

import (
	"fmt"
	"time"

	"dx-api/app/consts"
	"dx-api/app/models"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
)

// OrderDetail is the API response for an order.
type OrderDetail struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Product       string  `json:"product"`
	Amount        int     `json:"amount"`
	Status        string  `json:"status"`
	PaymentMethod *string `json:"paymentMethod"`
	ExpiresAt     string  `json:"expiresAt"`
	CreatedAt     string  `json:"createdAt"`
}

// CreateOrder creates a pending order with a 30-minute expiry.
func CreateOrder(userID, orderType, product string, amount int, paymentMethod string) (*OrderDetail, error) {
	now := time.Now()
	expiresAt := now.Add(30 * time.Minute)

	order := models.Order{
		ID:            newID(),
		UserID:        userID,
		Type:          orderType,
		Product:       product,
		Amount:        amount,
		Status:        consts.OrderStatusPending,
		PaymentMethod: &paymentMethod,
		ExpiresAt:     *carbon.NewDateTime(carbon.FromStdTime(expiresAt)),
	}

	if err := facades.Orm().Query().Create(&order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &OrderDetail{
		ID:            order.ID,
		Type:          order.Type,
		Product:       order.Product,
		Amount:        order.Amount,
		Status:        order.Status,
		PaymentMethod: order.PaymentMethod,
		ExpiresAt:     expiresAt.Format(time.RFC3339),
		CreatedAt:     now.Format(time.RFC3339),
	}, nil
}

// GetOrder fetches an order by ID with ownership check.
func GetOrder(orderID, userID string) (*OrderDetail, error) {
	var order models.Order
	if err := facades.Orm().Query().Where("id", orderID).First(&order); err != nil || order.ID == "" {
		return nil, ErrOrderNotFound
	}

	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	return &OrderDetail{
		ID:            order.ID,
		Type:          order.Type,
		Product:       order.Product,
		Amount:        order.Amount,
		Status:        order.Status,
		PaymentMethod: order.PaymentMethod,
		ExpiresAt:     order.ExpiresAt.StdTime().Format(time.RFC3339),
		CreatedAt:     order.CreatedAt.StdTime().Format(time.RFC3339),
	}, nil
}

// MarkPaid sets the order to paid and triggers fulfillment.
func MarkPaid(orderID, paymentNo string) error {
	var order models.Order
	if err := facades.Orm().Query().Where("id", orderID).First(&order); err != nil || order.ID == "" {
		return ErrOrderNotFound
	}

	if order.Status != consts.OrderStatusPending {
		return ErrOrderNotPending
	}

	now := carbon.Now()
	if _, err := facades.Orm().Query().Model(&models.Order{}).Where("id", orderID).Update(map[string]any{
		"status":     consts.OrderStatusPaid,
		"payment_no": paymentNo,
		"paid_at":    now,
	}); err != nil {
		return fmt.Errorf("failed to mark order paid: %w", err)
	}

	return FulfillOrder(orderID)
}

// FulfillOrder processes the paid order: upgrades membership or credits beans.
func FulfillOrder(orderID string) error {
	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	var order models.Order
	if err := tx.Where("id", orderID).LockForUpdate().First(&order); err != nil || order.ID == "" {
		_ = tx.Rollback()
		return ErrOrderNotFound
	}

	if order.Status != consts.OrderStatusPaid {
		_ = tx.Rollback()
		return ErrOrderNotPending
	}

	switch order.Type {
	case consts.OrderTypeMembership:
		if err := fulfillMembership(tx, &order); err != nil {
			_ = tx.Rollback()
			return err
		}
	case consts.OrderTypeBeans:
		if err := fulfillBeans(tx, &order); err != nil {
			_ = tx.Rollback()
			return err
		}
	default:
		_ = tx.Rollback()
		return ErrInvalidProduct
	}

	now := carbon.Now()
	if _, err := tx.Model(&models.Order{}).Where("id", orderID).Update(map[string]any{
		"status":       consts.OrderStatusFulfilled,
		"fulfilled_at": now,
	}); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to mark order fulfilled: %w", err)
	}

	return tx.Commit()
}

// fulfillMembership upgrades user grade, extends VIP, and grants beans.
// Uses same logic as RedeemCode in redeem_service.go.
func fulfillMembership(tx orm.Query, order *models.Order) error {
	var user models.User
	if err := tx.Where("id", order.UserID).LockForUpdate().First(&user); err != nil || user.ID == "" {
		return ErrUserNotFound
	}

	newVipDueAt := calcVipDueAt(order.Product, user.VipDueAt)

	if _, err := tx.Exec(
		"UPDATE users SET grade = ?, vip_due_at = ?, updated_at = NOW() WHERE id = ?",
		order.Product, newVipDueAt, order.UserID,
	); err != nil {
		return fmt.Errorf("failed to update user VIP: %w", err)
	}

	beanAmount := 10000
	if order.Product == consts.UserGradeLifetime {
		beanAmount = 15000
	}

	newBalance := user.Beans + beanAmount
	if _, err := tx.Model(&models.User{}).Where("id", order.UserID).
		Update("beans", newBalance); err != nil {
		return fmt.Errorf("failed to grant beans: %w", err)
	}

	ledger := models.UserBean{
		ID:     newID(),
		UserID: order.UserID,
		Beans:  beanAmount,
		Origin: user.Beans,
		Result: newBalance,
		Slug:   consts.BeanSlugMembershipGrant,
		Reason: consts.BeanReasonMembershipGrant,
	}
	return tx.Create(&ledger)
}

// fulfillBeans credits base + bonus beans to user balance.
func fulfillBeans(tx orm.Query, order *models.Order) error {
	pkg, ok := consts.BeanPackages[order.Product]
	if !ok {
		return ErrInvalidProduct
	}

	var user models.User
	if err := tx.Where("id", order.UserID).LockForUpdate().First(&user); err != nil || user.ID == "" {
		return ErrUserNotFound
	}

	totalBeans := pkg.Beans + pkg.Bonus
	newBalance := user.Beans + totalBeans

	if _, err := tx.Model(&models.User{}).Where("id", order.UserID).
		Update("beans", newBalance); err != nil {
		return fmt.Errorf("failed to credit beans: %w", err)
	}

	ledger := models.UserBean{
		ID:     newID(),
		UserID: order.UserID,
		Beans:  totalBeans,
		Origin: user.Beans,
		Result: newBalance,
		Slug:   consts.BeanSlugPurchaseGrant,
		Reason: consts.BeanReasonPurchaseGrant,
	}
	return tx.Create(&ledger)
}

// ExpireStaleOrders bulk-expires pending orders past their expiry time.
func ExpireStaleOrders() (int64, error) {
	res, err := facades.Orm().Query().Exec(
		"UPDATE orders SET status = ?, updated_at = NOW() WHERE status = ? AND expires_at < NOW()",
		consts.OrderStatusExpired, consts.OrderStatusPending,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to expire stale orders: %w", err)
	}
	return res.RowsAffected, nil
}
