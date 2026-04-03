# Purchase, Order & Payment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a purchase flow for membership subscriptions and energy bean packages with orders, payment stubs, and dedicated frontend pages.

**Architecture:** Single `orders` table with type discriminator (`membership`/`beans`). Backend creates orders via `order_controller`, serves order details via `payment_controller` (with stub callbacks). Frontend has a `/purchase` route group with membership, beans, and payment pages. All prices stored in fen (integer).

**Tech Stack:** Go/Goravel (backend), Next.js 16 + React 19 + TailwindCSS v4 (frontend), PostgreSQL, UUID v7 IDs

---

## File Structure

### Backend — New Files

| File | Responsibility |
|------|---------------|
| `dx-api/app/consts/order_type.go` | Order type constants |
| `dx-api/app/consts/order_status.go` | Order status constants |
| `dx-api/app/consts/payment_method.go` | Payment method constants |
| `dx-api/app/consts/bean_package.go` | Bean package definitions (slug, price, beans, bonus) |
| `dx-api/app/models/order.go` | Order GORM model |
| `dx-api/database/migrations/20260403000001_create_orders_table.go` | Orders table migration |
| `dx-api/app/http/requests/api/order_request.go` | Request validation for order creation |
| `dx-api/app/services/api/order_service.go` | Order business logic (create, get, fulfill, expire) |
| `dx-api/app/http/controllers/api/order_controller.go` | Create membership/beans orders |
| `dx-api/app/http/controllers/api/payment_controller.go` | Get order details, payment callback stubs |
| `dx-api/app/console/commands/expire_stale_orders.go` | Scheduled command for order expiry |

### Backend — Modified Files

| File | Change |
|------|--------|
| `dx-api/app/consts/error_code.go` | Add `CodeOrderNotFound`, `CodeOrderNotPending`, `CodeInvalidProduct` |
| `dx-api/app/consts/bean_slug.go` | Add `BeanSlugPurchaseGrant` |
| `dx-api/app/consts/bean_reason.go` | Add `BeanReasonPurchaseGrant` |
| `dx-api/app/services/api/errors.go` | Add `ErrOrderNotFound`, `ErrOrderNotPending`, `ErrInvalidProduct` |
| `dx-api/app/helpers/enum_rules.go` | Add `grade`, `payment_method`, `bean_package` enums |
| `dx-api/bootstrap/migrations.go` | Register orders migration |
| `dx-api/bootstrap/app.go` | Register expire command + schedule |
| `dx-api/routes/api.go` | Add order + payment routes |

### Frontend — New Files

| File | Responsibility |
|------|---------------|
| `dx-web/src/consts/order-type.ts` | Order type constants |
| `dx-web/src/consts/order-status.ts` | Order status constants |
| `dx-web/src/consts/payment-method.ts` | Payment method constants |
| `dx-web/src/consts/bean-package.ts` | Bean package data |
| `dx-web/src/features/web/purchase/types/order.types.ts` | Order TypeScript types |
| `dx-web/src/features/web/purchase/components/pricing-grid.tsx` | Moved from auth, with order creation |
| `dx-web/src/features/web/purchase/components/pricing-card.tsx` | Moved from auth, with onClick handler |
| `dx-web/src/features/web/purchase/components/bean-package-grid.tsx` | Bean tier cards grid |
| `dx-web/src/features/web/purchase/components/bean-package-card.tsx` | Single bean card |
| `dx-web/src/features/web/purchase/components/order-payment.tsx` | Payment page client component |
| `dx-web/src/features/web/purchase/components/testimonials-grid.tsx` | Moved from auth (unchanged) |
| `dx-web/src/features/web/purchase/components/testimonial-card.tsx` | Moved from auth (unchanged) |
| `dx-web/src/features/web/purchase/components/faq-section.tsx` | Moved from auth (unchanged) |
| `dx-web/src/app/(web)/purchase/layout.tsx` | Purchase layout with back button |
| `dx-web/src/app/(web)/purchase/membership/page.tsx` | Membership purchase page |
| `dx-web/src/app/(web)/purchase/beans/page.tsx` | Bean recharge page |
| `dx-web/src/app/(web)/purchase/payment/[orderId]/page.tsx` | Payment page |

### Frontend — Modified Files

| File | Change |
|------|--------|
| `dx-web/src/lib/api-client.ts` | Add `orderApi` namespace |
| `dx-web/src/consts/bean-slug.ts` | Add `PURCHASE_GRANT` |
| `dx-web/src/consts/bean-reason.ts` | Add `PURCHASE_GRANT` |
| `dx-web/src/components/in/insufficient-beans-dialog.tsx` | `/recharge` -> `/purchase/beans` |
| `dx-web/src/features/web/me/components/membership-block.tsx` | `/auth/membership` -> `/purchase/membership` |
| `dx-web/src/features/web/hall/components/ad-cards-row.tsx` | `/auth/membership` -> `/purchase/membership` |
| `dx-web/src/features/web/hall/components/hall-sidebar.tsx` | `/auth/membership` -> `/purchase/membership` |
| `dx-web/src/features/web/auth/components/user-profile-menu.tsx` | `/auth/membership` -> `/purchase/membership` |

### Frontend — Deleted Files

| File | Reason |
|------|--------|
| `dx-web/src/app/(web)/auth/membership/page.tsx` | Moved to `/purchase/membership` |
| `dx-web/src/app/(web)/auth/membership/pay/confirm/page.tsx` | Replaced by `/purchase/payment/[orderId]` |
| `dx-web/src/app/(web)/auth/membership/pay/[method]/page.tsx` | Replaced by `/purchase/payment/[orderId]` |
| `dx-web/src/app/(web)/recharge/page.tsx` | Replaced by `/purchase/beans` |

---

## Task 1: Backend Constants

**Files:**
- Create: `dx-api/app/consts/order_type.go`
- Create: `dx-api/app/consts/order_status.go`
- Create: `dx-api/app/consts/payment_method.go`
- Create: `dx-api/app/consts/bean_package.go`
- Modify: `dx-api/app/consts/error_code.go:60` (add new codes)
- Modify: `dx-api/app/consts/bean_slug.go:18` (add purchase grant)
- Modify: `dx-api/app/consts/bean_reason.go:18` (add purchase grant)

- [ ] **Step 1: Create `order_type.go`**

```go
package consts

const (
	OrderTypeMembership = "membership"
	OrderTypeBeans      = "beans"
)
```

- [ ] **Step 2: Create `order_status.go`**

```go
package consts

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusFulfilled = "fulfilled"
	OrderStatusExpired   = "expired"
	OrderStatusCancelled = "cancelled"
)
```

- [ ] **Step 3: Create `payment_method.go`**

```go
package consts

const (
	PaymentMethodWechat = "wechat"
	PaymentMethodAlipay = "alipay"
)
```

- [ ] **Step 4: Create `bean_package.go`**

```go
package consts

// BeanPackage defines a purchasable bean package.
type BeanPackage struct {
	Price int // price in fen (1 yuan = 100 fen)
	Beans int // base bean amount
	Bonus int // bonus bean amount
}

// BeanPackages maps package slugs to their definitions.
var BeanPackages = map[string]BeanPackage{
	"beans-1":   {Price: 100, Beans: 1000, Bonus: 0},
	"beans-5":   {Price: 500, Beans: 5000, Bonus: 0},
	"beans-10":  {Price: 1000, Beans: 10000, Bonus: 1000},
	"beans-50":  {Price: 5000, Beans: 50000, Bonus: 7500},
	"beans-100": {Price: 10000, Beans: 100000, Bonus: 20000},
}

// BeanPackageSlugs lists valid package slugs for validation.
var BeanPackageSlugs = []string{"beans-1", "beans-5", "beans-10", "beans-50", "beans-100"}
```

- [ ] **Step 5: Add error codes to `error_code.go`**

Append after line 60 (before closing `)`):

```go
	// 404xx: Not Found (continued)
	CodeOrderNotFound = 40411

	// 400xx: Validation (continued)
	CodeOrderNotPending = 40013
	CodeInvalidProduct  = 40014
```

- [ ] **Step 6: Add purchase grant bean slug and reason**

In `bean_slug.go`, add after `BeanSlugSeederGrant` (line 18):

```go
	BeanSlugPurchaseGrant = "purchase-grant"
```

In `bean_slug.go` `BeanSlugLabels` map, add:

```go
	BeanSlugPurchaseGrant: "能量豆充值",
```

In `bean_reason.go`, add after `BeanReasonSeederGrant` (line 18):

```go
	BeanReasonPurchaseGrant = "能量豆充值"
```

- [ ] **Step 7: Add enums to `enum_rules.go`**

In `dx-api/app/helpers/enum_rules.go`, add these entries to the `enumValues` map:

```go
	"grade":          {consts.UserGradeMonth, consts.UserGradeSeason, consts.UserGradeYear, consts.UserGradeLifetime},
	"payment_method": {consts.PaymentMethodWechat, consts.PaymentMethodAlipay},
	"bean_package":   consts.BeanPackageSlugs,
```

Note: the `grade` enum excludes `free` since free cannot be purchased.

- [ ] **Step 8: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 9: Commit**

```bash
cd dx-api
git add app/consts/order_type.go app/consts/order_status.go app/consts/payment_method.go app/consts/bean_package.go app/consts/error_code.go app/consts/bean_slug.go app/consts/bean_reason.go app/helpers/enum_rules.go
git commit -m "feat: add order, payment, and bean package constants"
```

---

## Task 2: Order Model + Migration

**Files:**
- Create: `dx-api/app/models/order.go`
- Create: `dx-api/database/migrations/20260403000001_create_orders_table.go`
- Modify: `dx-api/bootstrap/migrations.go:60` (register migration)

- [ ] **Step 1: Create Order model**

Create `dx-api/app/models/order.go`:

```go
package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
)

type Order struct {
	orm.Timestamps
	ID            string           `gorm:"column:id;primaryKey" json:"id"`
	UserID        string           `gorm:"column:user_id" json:"user_id"`
	Type          string           `gorm:"column:type" json:"type"`
	Product       string           `gorm:"column:product" json:"product"`
	Amount        int              `gorm:"column:amount" json:"amount"`
	Status        string           `gorm:"column:status" json:"status"`
	PaymentMethod *string          `gorm:"column:payment_method" json:"payment_method"`
	PaymentNo     *string          `gorm:"column:payment_no" json:"payment_no"`
	PaidAt        *carbon.DateTime `gorm:"column:paid_at" json:"paid_at"`
	FulfilledAt   *carbon.DateTime `gorm:"column:fulfilled_at" json:"fulfilled_at"`
	ExpiresAt     carbon.DateTime  `gorm:"column:expires_at" json:"expires_at"`
}

func (o *Order) TableName() string {
	return "orders"
}
```

- [ ] **Step 2: Create migration**

Create `dx-api/database/migrations/20260403000001_create_orders_table.go`:

```go
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260403000001CreateOrdersTable struct{}

func (r *M20260403000001CreateOrdersTable) Signature() string {
	return "20260403000001_create_orders_table"
}

func (r *M20260403000001CreateOrdersTable) Up() error {
	if !facades.Schema().HasTable("orders") {
		return facades.Schema().Create("orders", func(table schema.Blueprint) {
			table.Uuid("id")
			table.Primary("id")
			table.Uuid("user_id")
			table.Text("type").Default("")
			table.Text("product").Default("")
			table.Integer("amount").Default(0)
			table.Text("status").Default("pending")
			table.Text("payment_method").Nullable()
			table.Text("payment_no").Nullable()
			table.TimestampTz("paid_at").Nullable()
			table.TimestampTz("fulfilled_at").Nullable()
			table.TimestampTz("expires_at")
			table.TimestampsTz()
			table.Index("user_id")
			table.Index("user_id", "status")
			table.Index("status", "expires_at")
		})
	}
	return nil
}

func (r *M20260403000001CreateOrdersTable) Down() error {
	return facades.Schema().DropIfExists("orders")
}
```

- [ ] **Step 3: Register migration**

In `dx-api/bootstrap/migrations.go`, add at the end of the return slice (after line 59, the `M20260325000002AddSessionIndexes` entry):

```go
		&migrations.M20260403000001CreateOrdersTable{},
```

- [ ] **Step 4: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
cd dx-api
git add app/models/order.go database/migrations/20260403000001_create_orders_table.go bootstrap/migrations.go
git commit -m "feat: add Order model and create_orders migration"
```

---

## Task 3: Order Service + Error Sentinels

**Files:**
- Create: `dx-api/app/services/api/order_service.go`
- Modify: `dx-api/app/services/api/errors.go:58` (add error sentinels)

- [ ] **Step 1: Add error sentinels**

In `dx-api/app/services/api/errors.go`, add after the last error (line 57, `ErrGroupSubgroupsFull`):

```go
	ErrOrderNotFound   = errors.New("订单不存在")
	ErrOrderNotPending = errors.New("订单状态异常")
	ErrInvalidProduct  = errors.New("无效的商品")
```

- [ ] **Step 2: Create order service**

Create `dx-api/app/services/api/order_service.go`:

```go
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
		ExpiresAt:     carbon.FromStdTime(expiresAt),
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
	if _, err := facades.Orm().Query().Model(&models.Order{}).Where("id", orderID).Updates(map[string]any{
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
	if _, err := tx.Model(&models.Order{}).Where("id", orderID).Updates(map[string]any{
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
```

Note: `calcVipDueAt` is already defined in `redeem_service.go` in the same `api` package, so it's accessible directly.

- [ ] **Step 3: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
cd dx-api
git add app/services/api/order_service.go app/services/api/errors.go
git commit -m "feat: add order service with create, get, fulfill, and expire logic"
```

---

## Task 4: Request Validation + Controllers

**Files:**
- Create: `dx-api/app/http/requests/api/order_request.go`
- Create: `dx-api/app/http/controllers/api/order_controller.go`
- Create: `dx-api/app/http/controllers/api/payment_controller.go`

- [ ] **Step 1: Create request validation**

Create `dx-api/app/http/requests/api/order_request.go`:

```go
package api

import (
	"dx-api/app/helpers"

	"github.com/goravel/framework/contracts/http"
)

type CreateMembershipOrderRequest struct {
	Grade         string `form:"grade" json:"grade"`
	PaymentMethod string `form:"paymentMethod" json:"paymentMethod"`
}

func (r *CreateMembershipOrderRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateMembershipOrderRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"grade":         "required|" + helpers.InEnum("grade"),
		"paymentMethod": "required|" + helpers.InEnum("payment_method"),
	}
}
func (r *CreateMembershipOrderRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"grade":         "trim",
		"paymentMethod": "trim",
	}
}
func (r *CreateMembershipOrderRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"grade.required":         "请选择会员等级",
		"grade.in":               "无效的会员等级",
		"paymentMethod.required": "请选择支付方式",
		"paymentMethod.in":       "无效的支付方式",
	}
}

type CreateBeansOrderRequest struct {
	Package       string `form:"package" json:"package"`
	PaymentMethod string `form:"paymentMethod" json:"paymentMethod"`
}

func (r *CreateBeansOrderRequest) Authorize(ctx http.Context) error { return nil }
func (r *CreateBeansOrderRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"package":       "required|" + helpers.InEnum("bean_package"),
		"paymentMethod": "required|" + helpers.InEnum("payment_method"),
	}
}
func (r *CreateBeansOrderRequest) Filters(ctx http.Context) map[string]string {
	return map[string]string{
		"package":       "trim",
		"paymentMethod": "trim",
	}
}
func (r *CreateBeansOrderRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"package.required":       "请选择能量豆套餐",
		"package.in":             "无效的能量豆套餐",
		"paymentMethod.required": "请选择支付方式",
		"paymentMethod.in":       "无效的支付方式",
	}
}
```

- [ ] **Step 2: Create order controller**

Create `dx-api/app/http/controllers/api/order_controller.go`:

```go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	requests "dx-api/app/http/requests/api"
	services "dx-api/app/services/api"
)

type OrderController struct{}

func NewOrderController() *OrderController {
	return &OrderController{}
}

func (c *OrderController) CreateMembershipOrder(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateMembershipOrderRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	price, ok := consts.UserGradePrices[req.Grade]
	if !ok || price == 0 {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidProduct, "无效的会员等级")
	}

	// Convert yuan to fen
	amountFen := price * 100

	order, err := services.CreateOrder(userID, consts.OrderTypeMembership, req.Grade, amountFen, req.PaymentMethod)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to create order")
	}

	return helpers.Success(ctx, order)
}

func (c *OrderController) CreateBeansOrder(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	var req requests.CreateBeansOrderRequest
	if resp := helpers.Validate(ctx, &req); resp != nil {
		return resp
	}

	pkg, ok := consts.BeanPackages[req.Package]
	if !ok {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeInvalidProduct, "无效的能量豆套餐")
	}

	order, err := services.CreateOrder(userID, consts.OrderTypeBeans, req.Package, pkg.Price, req.PaymentMethod)
	if err != nil {
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to create order")
	}

	return helpers.Success(ctx, order)
}
```

- [ ] **Step 3: Create payment controller**

Create `dx-api/app/http/controllers/api/payment_controller.go`:

```go
package api

import (
	"errors"
	"net/http"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"dx-api/app/consts"
	"dx-api/app/helpers"
	services "dx-api/app/services/api"
)

type PaymentController struct{}

func NewPaymentController() *PaymentController {
	return &PaymentController{}
}

// GetOrder returns order details for the payment page.
func (c *PaymentController) GetOrder(ctx contractshttp.Context) contractshttp.Response {
	userID, err := facades.Auth(ctx).Guard("user").ID()
	if err != nil || userID == "" {
		return helpers.Error(ctx, http.StatusUnauthorized, consts.CodeUnauthorized, "unauthorized")
	}

	orderID := ctx.Request().Route("id")
	if orderID == "" {
		return helpers.Error(ctx, http.StatusBadRequest, consts.CodeValidationError, "missing order id")
	}

	order, err := services.GetOrder(orderID, userID)
	if err != nil {
		if errors.Is(err, services.ErrOrderNotFound) {
			return helpers.Error(ctx, http.StatusNotFound, consts.CodeOrderNotFound, "订单不存在")
		}
		return helpers.Error(ctx, http.StatusInternalServerError, consts.CodeInternalError, "failed to get order")
	}

	return helpers.Success(ctx, order)
}

// WechatCallback handles WeChat Pay callback notifications.
// Stub — will be implemented when payment license is obtained.
func (c *PaymentController) WechatCallback(ctx contractshttp.Context) contractshttp.Response {
	return helpers.Success(ctx, nil)
}

// AlipayCallback handles Alipay callback notifications.
// Stub — will be implemented when payment license is obtained.
func (c *PaymentController) AlipayCallback(ctx contractshttp.Context) contractshttp.Response {
	return helpers.Success(ctx, nil)
}
```

- [ ] **Step 4: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
cd dx-api
git add app/http/requests/api/order_request.go app/http/controllers/api/order_controller.go app/http/controllers/api/payment_controller.go
git commit -m "feat: add order and payment controllers with request validation"
```

---

## Task 5: Routes + Expire Command + Bootstrap

**Files:**
- Modify: `dx-api/routes/api.go` (add order + payment routes)
- Create: `dx-api/app/console/commands/expire_stale_orders.go`
- Modify: `dx-api/bootstrap/app.go` (register command + schedule)

- [ ] **Step 1: Add routes**

In `dx-api/routes/api.go`, add within the `/api` prefix group:

After the public group routes (after the SSE endpoints around line 84), add the public payment callbacks:

```go
		// Public payment callbacks (no JWT required)
		paymentController := apicontrollers.NewPaymentController()
		router.Post("/payments/callback/wechat", paymentController.WechatCallback)
		router.Post("/payments/callback/alipay", paymentController.AlipayCallback)
```

Inside the `protected` group (after the redeems section around line 199), add:

```go
			// Orders
			orderController := apicontrollers.NewOrderController()
			protected.Post("/orders/membership", orderController.CreateMembershipOrder)
			protected.Post("/orders/beans", orderController.CreateBeansOrder)
			protected.Get("/orders/{id}", paymentController.GetOrder)
```

Note: `paymentController` is declared in the outer scope (public routes), so `GetOrder` can be referenced in the protected group.

Actually, `paymentController` is declared outside the protected closure. For it to be accessible inside, it needs to be declared before the protected group. Move the declaration to before the protected middleware group, or declare a new instance. The simplest approach: move the `paymentController` declaration to just before the protected group starts, after the public payment callback routes.

- [ ] **Step 2: Create expire command**

Create `dx-api/app/console/commands/expire_stale_orders.go`:

```go
package commands

import (
	"fmt"
	"time"

	services "dx-api/app/services/api"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type ExpireStaleOrders struct{}

func (c *ExpireStaleOrders) Signature() string {
	return "app:expire-stale-orders"
}

func (c *ExpireStaleOrders) Description() string {
	return "Expire unpaid orders past their 30-minute window"
}

func (c *ExpireStaleOrders) Extend() command.Extend {
	return command.Extend{}
}

func (c *ExpireStaleOrders) Handle(ctx console.Context) error {
	start := time.Now()

	count, err := services.ExpireStaleOrders()
	if err != nil {
		ctx.Error(fmt.Sprintf("[expire-stale-orders] failed: %v", err))
		return err
	}

	elapsed := time.Since(start)
	ctx.Info(fmt.Sprintf("[expire-stale-orders] done in %s — expired: %d", elapsed, count))
	return nil
}
```

- [ ] **Step 3: Register command and schedule**

In `dx-api/bootstrap/app.go`, add to the commands slice (after `&commands.ImportCourses{},` on line 30):

```go
				&commands.ExpireStaleOrders{},
```

Add to the schedule events slice (after the `update-play-streaks` entry on line 36):

```go
				facades.Schedule().Command("app:expire-stale-orders").EveryFiveMinutes().SkipIfStillRunning().Name("expire-stale-orders"),
```

- [ ] **Step 4: Verify build**

Run: `cd dx-api && go build ./...`
Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
cd dx-api
git add routes/api.go app/console/commands/expire_stale_orders.go bootstrap/app.go
git commit -m "feat: add order/payment routes and expire-stale-orders scheduled command"
```

---

## Task 6: Frontend Constants + Types + API Client

**Files:**
- Create: `dx-web/src/consts/order-type.ts`
- Create: `dx-web/src/consts/order-status.ts`
- Create: `dx-web/src/consts/payment-method.ts`
- Create: `dx-web/src/consts/bean-package.ts`
- Create: `dx-web/src/features/web/purchase/types/order.types.ts`
- Modify: `dx-web/src/consts/bean-slug.ts:15` (add purchase grant)
- Modify: `dx-web/src/consts/bean-reason.ts:14` (add purchase grant)
- Modify: `dx-web/src/lib/api-client.ts:797` (add orderApi)

- [ ] **Step 1: Create `order-type.ts`**

```ts
export const ORDER_TYPES = {
  MEMBERSHIP: "membership",
  BEANS: "beans",
} as const;

export type OrderType = (typeof ORDER_TYPES)[keyof typeof ORDER_TYPES];
```

- [ ] **Step 2: Create `order-status.ts`**

```ts
export const ORDER_STATUSES = {
  PENDING: "pending",
  PAID: "paid",
  FULFILLED: "fulfilled",
  EXPIRED: "expired",
  CANCELLED: "cancelled",
} as const;

export type OrderStatus = (typeof ORDER_STATUSES)[keyof typeof ORDER_STATUSES];

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  pending: "待支付",
  paid: "已支付",
  fulfilled: "已完成",
  expired: "已过期",
  cancelled: "已取消",
};
```

- [ ] **Step 3: Create `payment-method.ts`**

```ts
export const PAYMENT_METHODS = {
  WECHAT: "wechat",
  ALIPAY: "alipay",
} as const;

export type PaymentMethod =
  (typeof PAYMENT_METHODS)[keyof typeof PAYMENT_METHODS];

export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  wechat: "微信支付",
  alipay: "支付宝",
};
```

- [ ] **Step 4: Create `bean-package.ts`**

```ts
export type BeanPackage = {
  slug: string;
  beans: number;
  bonus: number;
  price: number; // fen
  tag?: string;
};

export const BEAN_PACKAGES: BeanPackage[] = [
  { slug: "beans-1", beans: 1000, bonus: 0, price: 100 },
  { slug: "beans-5", beans: 5000, bonus: 0, price: 500 },
  { slug: "beans-10", beans: 10000, bonus: 1000, price: 1000, tag: "超值推荐" },
  { slug: "beans-50", beans: 50000, bonus: 7500, price: 5000 },
  { slug: "beans-100", beans: 100000, bonus: 20000, price: 10000, tag: "最划算" },
];
```

- [ ] **Step 5: Create order types**

Create directory and file `dx-web/src/features/web/purchase/types/order.types.ts`:

```ts
import type { OrderType } from "@/consts/order-type";
import type { OrderStatus } from "@/consts/order-status";
import type { PaymentMethod } from "@/consts/payment-method";

export type Order = {
  id: string;
  type: OrderType;
  product: string;
  amount: number;
  status: OrderStatus;
  paymentMethod: PaymentMethod | null;
  expiresAt: string;
  createdAt: string;
};
```

- [ ] **Step 6: Add purchase grant to bean-slug.ts and bean-reason.ts**

In `dx-web/src/consts/bean-slug.ts`, add to `BEAN_SLUGS` object:

```ts
  PURCHASE_GRANT: "purchase-grant",
```

And to `BEAN_SLUG_LABELS`:

```ts
  "purchase-grant": "能量豆充值",
```

In `dx-web/src/consts/bean-reason.ts`, add to `BEAN_REASONS`:

```ts
  PURCHASE_GRANT: "能量豆充值",
```

- [ ] **Step 7: Add orderApi to api-client.ts**

At the end of `dx-web/src/lib/api-client.ts`, before the final `export type` line (line 799), add:

```ts
// Order API functions
export const orderApi = {
  async createMembershipOrder(data: { grade: string; paymentMethod: string }) {
    return apiClient.post<{
      id: string;
      type: string;
      product: string;
      amount: number;
      status: string;
      paymentMethod: string;
      expiresAt: string;
      createdAt: string;
    }>("/api/orders/membership", data);
  },
  async createBeansOrder(data: { package: string; paymentMethod: string }) {
    return apiClient.post<{
      id: string;
      type: string;
      product: string;
      amount: number;
      status: string;
      paymentMethod: string;
      expiresAt: string;
      createdAt: string;
    }>("/api/orders/beans", data);
  },
  async getOrder(id: string) {
    return apiClient.get<{
      id: string;
      type: string;
      product: string;
      amount: number;
      status: string;
      paymentMethod: string | null;
      expiresAt: string;
      createdAt: string;
    }>(`/api/orders/${id}`);
  },
};
```

- [ ] **Step 8: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: No errors.

- [ ] **Step 9: Commit**

```bash
cd dx-web
git add src/consts/order-type.ts src/consts/order-status.ts src/consts/payment-method.ts src/consts/bean-package.ts src/features/web/purchase/types/order.types.ts src/consts/bean-slug.ts src/consts/bean-reason.ts src/lib/api-client.ts
git commit -m "feat: add frontend order/payment constants, types, and API client"
```

---

## Task 7: Purchase Layout + Move Membership Components

**Files:**
- Create: `dx-web/src/app/(web)/purchase/layout.tsx`
- Create: `dx-web/src/features/web/purchase/components/pricing-card.tsx` (moved from auth)
- Create: `dx-web/src/features/web/purchase/components/pricing-grid.tsx` (moved from auth, with order creation)
- Create: `dx-web/src/features/web/purchase/components/testimonial-card.tsx` (moved from auth)
- Create: `dx-web/src/features/web/purchase/components/testimonials-grid.tsx` (moved from auth)
- Create: `dx-web/src/features/web/purchase/components/faq-section.tsx` (moved from auth)

- [ ] **Step 1: Create purchase layout**

Create `dx-web/src/app/(web)/purchase/layout.tsx`:

```tsx
"use client";

import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";

export default function PurchaseLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();

  return (
    <div className="flex min-h-screen w-full flex-col items-center bg-white">
      <header className="flex w-full max-w-[1200px] items-center px-4 pt-6 lg:px-8">
        <button
          type="button"
          onClick={() => router.back()}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-slate-100 hover:bg-slate-200"
        >
          <ArrowLeft className="h-5 w-5 text-slate-600" />
        </button>
      </header>
      <main className="flex w-full flex-1 flex-col items-center">
        {children}
      </main>
    </div>
  );
}
```

- [ ] **Step 2: Move pricing-card.tsx**

Create `dx-web/src/features/web/purchase/components/pricing-card.tsx`:

```tsx
"use client";

import { CircleCheck } from "lucide-react";

interface PricingCardProps {
  name: string;
  price: string;
  period: string;
  features: string[];
  bgColor: string;
  borderColor?: string;
  ctaText: string;
  highlight?: boolean;
  disabled?: boolean;
  onSubscribe?: () => void;
}

export function PricingCard({
  name,
  price,
  period,
  features,
  bgColor,
  borderColor,
  ctaText,
  highlight,
  disabled,
  onSubscribe,
}: PricingCardProps) {
  return (
    <div
      className={`flex flex-1 flex-col gap-3 rounded-[14px] p-5 ${bgColor}`}
      style={borderColor ? { border: `1px solid ${borderColor}` } : undefined}
    >
      <span className="text-base font-semibold text-white">{name}</span>
      <div className="flex items-end gap-1">
        <span className="text-4xl font-extrabold text-white">{price}</span>
        {period && <span className="mb-1 text-sm text-white/70">{period}</span>}
      </div>
      <button
        type="button"
        disabled={disabled}
        onClick={onSubscribe}
        className={`w-full rounded-[10px] py-2.5 text-center text-sm font-semibold ${
          disabled
            ? "cursor-default border border-white/25 bg-white/10 text-white/50"
            : highlight
              ? "bg-white text-teal-700 hover:bg-white/90"
              : "border border-white/25 bg-white/20 text-white hover:bg-white/30"
        }`}
      >
        {ctaText}
      </button>
      <div className="h-px w-full bg-white/30" />
      <div className="flex flex-1 flex-col gap-1.5">
        {features.map((f) => (
          <div key={f} className="flex items-center gap-2">
            <CircleCheck className="h-4 w-4 shrink-0 text-white/70" />
            <span className="text-[13px] text-white/80">{f}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Move pricing-grid.tsx with order creation**

Create `dx-web/src/features/web/purchase/components/pricing-grid.tsx`:

```tsx
"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { PricingCard } from "@/features/web/purchase/components/pricing-card";
import { USER_GRADE_PRICES } from "@/consts/user-grade";
import { PAYMENT_METHODS } from "@/consts/payment-method";
import { orderApi } from "@/lib/api-client";

const plans = [
  {
    grade: "free",
    name: "免费会员",
    price: `¥${USER_GRADE_PRICES.free}`,
    period: "",
    features: [
      "免费课程内容",
      "免费游戏试玩",
      "少量学习小组",
      "分享推广佣金",
    ],
    bgColor: "bg-slate-500",
    borderColor: "#475569",
    ctaText: "当前方案",
  },
  {
    grade: "month",
    name: "月度会员",
    price: `¥${USER_GRADE_PRICES.month}`,
    period: "/月",
    features: [
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "分享推广佣金",
    ],
    bgColor: "bg-blue-500",
    borderColor: "#2563EB",
    ctaText: "立即订阅",
  },
  {
    grade: "season",
    name: "季度会员",
    price: `¥${USER_GRADE_PRICES.season}`,
    period: "/季",
    features: [
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "分享推广佣金",
      "学习服务支持",
    ],
    bgColor: "bg-violet-500",
    borderColor: "#6D28D9",
    ctaText: "立即订阅",
  },
  {
    grade: "year",
    name: "年度会员",
    price: `¥${USER_GRADE_PRICES.year}`,
    period: "/年",
    features: [
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "更多辅助功能",
      "分享推广佣金",
      "高级服务支持",
    ],
    bgColor: "bg-gradient-to-b from-[#0D7369] to-[#0A5C53]",
    ctaText: "立即订阅",
    highlight: true,
  },
  {
    grade: "lifetime",
    name: "终身会员",
    price: `¥${USER_GRADE_PRICES.lifetime}`,
    period: "",
    features: [
      "解锁所有权益",
      "功能永久有效",
      "永久没有续费",
      "全部课程内容",
      "全部游戏畅玩",
      "AI - 智能助力",
      "高级音效发音",
      "更多学习小组",
      "更多辅助功能",
      "更多推广佣金",
      "专属服务支持",
    ],
    bgColor: "bg-[#ca9302]",
    borderColor: "#D97706",
    ctaText: "立即订阅",
  },
];

export function PricingGrid() {
  const router = useRouter();
  const [loading, setLoading] = useState<string | null>(null);

  async function handleSubscribe(grade: string) {
    if (loading) return;
    setLoading(grade);
    try {
      const res = await orderApi.createMembershipOrder({
        grade,
        paymentMethod: PAYMENT_METHODS.WECHAT,
      });
      if (res.code === 0 && res.data?.id) {
        router.push(`/purchase/payment/${res.data.id}`);
      }
    } finally {
      setLoading(null);
    }
  }

  return (
    <div className="grid w-full grid-cols-1 gap-4 md:grid-cols-3 lg:grid-cols-5">
      {plans.map((p) => (
        <PricingCard
          key={p.name}
          name={p.name}
          price={p.price}
          period={p.period}
          features={p.features}
          bgColor={p.bgColor}
          borderColor={p.borderColor}
          ctaText={loading === p.grade ? "创建订单..." : p.ctaText}
          highlight={p.highlight}
          disabled={p.grade === "free" || loading !== null}
          onSubscribe={() => handleSubscribe(p.grade)}
        />
      ))}
    </div>
  );
}
```

- [ ] **Step 4: Move testimonial-card.tsx, testimonials-grid.tsx, faq-section.tsx**

Copy these 3 files from `src/features/web/auth/components/` to `src/features/web/purchase/components/`, updating the import paths:

`dx-web/src/features/web/purchase/components/testimonial-card.tsx` — identical to `auth` version (no imports to change).

`dx-web/src/features/web/purchase/components/testimonials-grid.tsx` — change import:

```tsx
import { TestimonialCard } from "@/features/web/purchase/components/testimonial-card";
```

`dx-web/src/features/web/purchase/components/faq-section.tsx` — identical to `auth` version (no external feature imports).

- [ ] **Step 5: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: No errors.

- [ ] **Step 6: Commit**

```bash
cd dx-web
git add src/app/\(web\)/purchase/layout.tsx src/features/web/purchase/components/pricing-card.tsx src/features/web/purchase/components/pricing-grid.tsx src/features/web/purchase/components/testimonial-card.tsx src/features/web/purchase/components/testimonials-grid.tsx src/features/web/purchase/components/faq-section.tsx
git commit -m "feat: add purchase layout and move membership components from auth"
```

---

## Task 8: Membership Page + Delete Old Routes

**Files:**
- Create: `dx-web/src/app/(web)/purchase/membership/page.tsx`
- Delete: `dx-web/src/app/(web)/auth/membership/page.tsx`
- Delete: `dx-web/src/app/(web)/auth/membership/pay/confirm/page.tsx`
- Delete: `dx-web/src/app/(web)/auth/membership/pay/[method]/page.tsx`
- Delete: `dx-web/src/app/(web)/recharge/page.tsx`

- [ ] **Step 1: Create membership purchase page**

Create `dx-web/src/app/(web)/purchase/membership/page.tsx`:

```tsx
import { CircleCheck } from "lucide-react";
import { PricingGrid } from "@/features/web/purchase/components/pricing-grid";
import { TestimonialsGrid } from "@/features/web/purchase/components/testimonials-grid";
import { FaqSection } from "@/features/web/purchase/components/faq-section";

export default function MembershipPurchasePage() {
  return (
    <div className="flex w-full max-w-[1200px] flex-col items-center gap-5 px-4 py-8 lg:gap-6 lg:px-8 lg:py-10">
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-2xl font-bold text-slate-900 lg:text-[32px]">
          会员订阅套餐
        </h1>
        <p className="text-sm text-slate-500">
          选择适合您的会员方案，解锁更多学习功能
        </p>
      </div>

      <div className="flex items-center gap-1.5 rounded-full border border-teal-600 bg-teal-50 px-3 py-1.5">
        <CircleCheck className="h-3.5 w-3.5 text-teal-600" />
        <span className="text-xs font-medium text-teal-600">
          当前方案: 免费版
        </span>
      </div>

      <PricingGrid />

      <div className="flex w-full flex-col items-center gap-8 pt-12">
        <div className="flex w-full items-center gap-4">
          <div className="h-px flex-1 bg-slate-300" />
          <h2 className="text-xl font-extrabold text-slate-900 lg:text-[28px]">
            会员真实体验
          </h2>
          <div className="h-px flex-1 bg-slate-300" />
        </div>
        <TestimonialsGrid />
      </div>

      <FaqSection />
    </div>
  );
}
```

- [ ] **Step 2: Delete old routes**

Delete these files/directories:
- `dx-web/src/app/(web)/auth/membership/` (entire directory)
- `dx-web/src/app/(web)/recharge/` (entire directory)

```bash
cd dx-web
rm -rf src/app/\(web\)/auth/membership
rm -rf src/app/\(web\)/recharge
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: No errors.

- [ ] **Step 4: Commit**

```bash
cd dx-web
git add src/app/\(web\)/purchase/membership/page.tsx
git add -u src/app/\(web\)/auth/membership/ src/app/\(web\)/recharge/
git commit -m "feat: add /purchase/membership page, remove old /auth/membership and /recharge"
```

---

## Task 9: Beans Purchase Page

**Files:**
- Create: `dx-web/src/features/web/purchase/components/bean-package-card.tsx`
- Create: `dx-web/src/features/web/purchase/components/bean-package-grid.tsx`
- Create: `dx-web/src/app/(web)/purchase/beans/page.tsx`

- [ ] **Step 1: Create bean-package-card.tsx**

Create `dx-web/src/features/web/purchase/components/bean-package-card.tsx`:

```tsx
"use client";

import { Gem } from "lucide-react";

interface BeanPackageCardProps {
  beans: number;
  bonus: number;
  price: number; // fen
  tag?: string;
  highlight?: boolean;
  disabled?: boolean;
  onPurchase?: () => void;
}

export function BeanPackageCard({
  beans,
  bonus,
  price,
  tag,
  highlight,
  disabled,
  onPurchase,
}: BeanPackageCardProps) {
  const priceYuan = price / 100;
  const totalDisplay = (beans + bonus).toLocaleString();

  return (
    <div
      className={`relative flex min-w-[180px] flex-1 flex-col items-center gap-3 rounded-[14px] border p-5 ${
        highlight
          ? "border-orange-400 bg-slate-50 shadow-[0_0_0_1px_rgba(251,146,60,0.3)]"
          : "border-slate-200 bg-slate-50"
      }`}
    >
      {tag && (
        <span
          className={`absolute -top-0 right-3 rounded-b-lg px-2.5 py-1 text-[11px] font-bold text-white ${
            highlight
              ? "bg-gradient-to-r from-orange-400 to-red-500"
              : "bg-gradient-to-r from-violet-500 to-pink-500"
          }`}
        >
          {tag}
        </span>
      )}

      <Gem className="h-10 w-10 text-blue-400" />

      <div className="flex flex-col items-center gap-0.5">
        <span className="text-2xl font-extrabold text-slate-900">
          {totalDisplay}
        </span>
        {bonus > 0 && (
          <span className="text-xs text-green-600">
            +{bonus.toLocaleString()} 赠送
          </span>
        )}
        <span className="text-xs text-slate-500">能量豆</span>
      </div>

      <span className="text-xl font-bold text-orange-500">
        ¥{priceYuan}
      </span>

      <button
        type="button"
        disabled={disabled}
        onClick={onPurchase}
        className={`w-full rounded-[10px] py-2.5 text-center text-sm font-semibold ${
          highlight
            ? "bg-gradient-to-r from-orange-400 to-red-500 text-white hover:from-orange-500 hover:to-red-600"
            : "border border-slate-300 bg-white text-slate-700 hover:bg-slate-100"
        }`}
      >
        立即购买
      </button>
    </div>
  );
}
```

- [ ] **Step 2: Create bean-package-grid.tsx**

Create `dx-web/src/features/web/purchase/components/bean-package-grid.tsx`:

```tsx
"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { BeanPackageCard } from "@/features/web/purchase/components/bean-package-card";
import { BEAN_PACKAGES } from "@/consts/bean-package";
import { PAYMENT_METHODS } from "@/consts/payment-method";
import { orderApi } from "@/lib/api-client";

export function BeanPackageGrid() {
  const router = useRouter();
  const [loading, setLoading] = useState<string | null>(null);

  async function handlePurchase(slug: string) {
    if (loading) return;
    setLoading(slug);
    try {
      const res = await orderApi.createBeansOrder({
        package: slug,
        paymentMethod: PAYMENT_METHODS.WECHAT,
      });
      if (res.code === 0 && res.data?.id) {
        router.push(`/purchase/payment/${res.data.id}`);
      }
    } finally {
      setLoading(null);
    }
  }

  return (
    <div className="flex w-full gap-4 overflow-x-auto pb-2 lg:overflow-visible">
      {BEAN_PACKAGES.map((pkg) => (
        <BeanPackageCard
          key={pkg.slug}
          beans={pkg.beans}
          bonus={pkg.bonus}
          price={pkg.price}
          tag={pkg.tag}
          highlight={pkg.slug === "beans-10"}
          disabled={loading !== null}
          onPurchase={() => handlePurchase(pkg.slug)}
        />
      ))}
    </div>
  );
}
```

- [ ] **Step 3: Create beans page**

Create `dx-web/src/app/(web)/purchase/beans/page.tsx`:

```tsx
import { Lightbulb } from "lucide-react";
import { BeanPackageGrid } from "@/features/web/purchase/components/bean-package-grid";

export default function BeansPurchasePage() {
  return (
    <div className="flex w-full max-w-[1200px] flex-col items-center gap-6 px-4 py-8 lg:px-8 lg:py-10">
      <div className="flex flex-col items-center gap-2">
        <h1 className="text-2xl font-bold text-slate-900 lg:text-[32px]">
          能量豆充值
        </h1>
        <p className="text-sm text-slate-500">选择适合您的能量豆套餐</p>
      </div>

      <BeanPackageGrid />

      <div className="mt-6 w-full rounded-2xl border border-slate-200 bg-slate-50 p-6">
        <div className="mb-3 flex items-center gap-2">
          <Lightbulb className="h-5 w-5 text-amber-500" />
          <span className="text-base font-bold text-slate-900">
            能量豆用途
          </span>
        </div>
        <ul className="flex flex-col gap-1.5 text-sm text-slate-600">
          <li>
            &bull; 编辑端 AI 功能（课程生成、句子拆分、内容加工等）
          </li>
          <li>&bull; 高级学习辅助功能</li>
          <li>&bull; 更多即将推出的增值服务</li>
        </ul>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: No errors.

- [ ] **Step 5: Commit**

```bash
cd dx-web
git add src/features/web/purchase/components/bean-package-card.tsx src/features/web/purchase/components/bean-package-grid.tsx src/app/\(web\)/purchase/beans/page.tsx
git commit -m "feat: add /purchase/beans page with bean package grid"
```

---

## Task 10: Payment Page

**Files:**
- Create: `dx-web/src/features/web/purchase/components/order-payment.tsx`
- Create: `dx-web/src/app/(web)/purchase/payment/[orderId]/page.tsx`

- [ ] **Step 1: Create order-payment.tsx client component**

Create `dx-web/src/features/web/purchase/components/order-payment.tsx`:

```tsx
"use client";

import { useEffect, useState } from "react";
import { Check, ScanLine, Clock, AlertCircle } from "lucide-react";
import { orderApi } from "@/lib/api-client";
import { ORDER_STATUSES } from "@/consts/order-status";
import { ORDER_TYPES } from "@/consts/order-type";
import { USER_GRADE_LABELS, type UserGrade } from "@/consts/user-grade";
import { BEAN_PACKAGES } from "@/consts/bean-package";
import { PAYMENT_METHODS, PAYMENT_METHOD_LABELS, type PaymentMethod } from "@/consts/payment-method";
import type { Order } from "@/features/web/purchase/types/order.types";

function getProductLabel(order: Order): string {
  if (order.type === ORDER_TYPES.MEMBERSHIP) {
    return USER_GRADE_LABELS[order.product as UserGrade] ?? order.product;
  }
  const pkg = BEAN_PACKAGES.find((p) => p.slug === order.product);
  if (pkg) {
    const total = pkg.beans + pkg.bonus;
    return `${total.toLocaleString()} 能量豆`;
  }
  return order.product;
}

function formatAmount(fen: number): string {
  return `¥${(fen / 100).toFixed(2)}`;
}

function useCountdown(expiresAt: string): string {
  const [remaining, setRemaining] = useState("");

  useEffect(() => {
    function update() {
      const diff = new Date(expiresAt).getTime() - Date.now();
      if (diff <= 0) {
        setRemaining("已过期");
        return;
      }
      const mins = Math.floor(diff / 60000);
      const secs = Math.floor((diff % 60000) / 1000);
      setRemaining(`${mins}:${secs.toString().padStart(2, "0")}`);
    }
    update();
    const timer = setInterval(update, 1000);
    return () => clearInterval(timer);
  }, [expiresAt]);

  return remaining;
}

export function OrderPayment({ orderId }: { orderId: string }) {
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [agreed, setAgreed] = useState(false);
  const [selectedMethod, setSelectedMethod] = useState<PaymentMethod>(PAYMENT_METHODS.WECHAT);

  useEffect(() => {
    orderApi.getOrder(orderId).then((res) => {
      if (res.code === 0 && res.data) {
        setOrder(res.data as Order);
        if (res.data.paymentMethod) {
          setSelectedMethod(res.data.paymentMethod as PaymentMethod);
        }
      } else {
        setError("订单不存在");
      }
    });
  }, [orderId]);

  const countdown = useCountdown(order?.expiresAt ?? "");

  if (error) {
    return (
      <div className="flex flex-col items-center gap-3 py-20">
        <AlertCircle className="h-10 w-10 text-red-400" />
        <span className="text-base text-slate-600">{error}</span>
      </div>
    );
  }

  if (!order) {
    return (
      <div className="flex items-center justify-center py-20">
        <span className="text-sm text-slate-400">加载中...</span>
      </div>
    );
  }

  if (order.status !== ORDER_STATUSES.PENDING) {
    const statusMessages: Record<string, string> = {
      [ORDER_STATUSES.PAID]: "订单已支付，正在处理中...",
      [ORDER_STATUSES.FULFILLED]: "订单已完成",
      [ORDER_STATUSES.EXPIRED]: "订单已过期，请重新下单",
      [ORDER_STATUSES.CANCELLED]: "订单已取消",
    };
    return (
      <div className="flex flex-col items-center gap-3 py-20">
        <AlertCircle className="h-10 w-10 text-slate-400" />
        <span className="text-base text-slate-600">
          {statusMessages[order.status] ?? "订单状态异常"}
        </span>
      </div>
    );
  }

  const methods: { key: PaymentMethod; color: string; icon: string }[] = [
    { key: PAYMENT_METHODS.WECHAT, color: "bg-[#07C160]", icon: "W" },
    { key: PAYMENT_METHODS.ALIPAY, color: "bg-[#1677FF]", icon: "A" },
  ];

  return (
    <div className="flex w-full max-w-[520px] flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-[0_8px_32px_rgba(15,23,42,0.1)]">
      {/* Order summary */}
      <div className="flex flex-col gap-3.5 px-7 py-6">
        <div className="flex items-center justify-between">
          <span className="text-base font-bold text-slate-900">
            {getProductLabel(order)}
          </span>
          <div className="flex items-center gap-1 text-sm text-slate-500">
            <Clock className="h-3.5 w-3.5" />
            <span>{countdown}</span>
          </div>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-xs text-slate-400">订单编号</span>
          <span className="text-xs text-slate-500">{order.id.slice(0, 18)}</span>
        </div>
        <div className="flex items-center gap-0.5">
          <span className="text-4xl font-bold text-teal-600">
            {formatAmount(order.amount)}
          </span>
        </div>
      </div>

      <div className="h-px w-full bg-slate-200" />

      {/* Agreement */}
      <div className="px-7 py-4">
        <label className="flex cursor-pointer gap-2.5">
          <button
            type="button"
            onClick={() => setAgreed(!agreed)}
            className={`flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border-[1.5px] ${
              agreed
                ? "border-teal-600 bg-teal-600"
                : "border-slate-300 bg-white"
            }`}
          >
            {agreed && <Check className="h-2.5 w-2.5 text-white" />}
          </button>
          <div className="flex flex-col gap-1">
            <span className="text-sm text-slate-700">
              我已阅读并同意以下协议
            </span>
            <span className="text-xs text-teal-600">《斗学服务协议》</span>
          </div>
        </label>
      </div>

      <div className="h-px w-full bg-slate-200" />

      {/* Payment method */}
      <div className="flex flex-col gap-4 px-7 py-4">
        {methods.map((m) => (
          <button
            key={m.key}
            type="button"
            onClick={() => setSelectedMethod(m.key)}
            className="flex items-center gap-2.5"
          >
            <div
              className={`flex h-[18px] w-[18px] items-center justify-center rounded-full ${
                selectedMethod === m.key
                  ? "bg-teal-600"
                  : "border-[1.5px] border-slate-300 bg-white"
              }`}
            >
              {selectedMethod === m.key && (
                <Check className="h-2.5 w-2.5 text-white" />
              )}
            </div>
            <div
              className={`flex h-[22px] w-[22px] items-center justify-center rounded-[5px] ${m.color}`}
            >
              <span className="text-[9px] font-bold text-white">{m.icon}</span>
            </div>
            <span
              className={`text-sm ${
                selectedMethod === m.key
                  ? "font-medium text-slate-900"
                  : "text-slate-600"
              }`}
            >
              {PAYMENT_METHOD_LABELS[m.key]}
            </span>
          </button>
        ))}
      </div>

      <div className="h-px w-full bg-slate-200" />

      {/* QR code placeholder */}
      <div className="flex flex-col items-center gap-3 px-7 py-6">
        <div className="flex h-[180px] w-[180px] items-center justify-center rounded-lg border border-slate-200 bg-slate-50">
          <div className="flex flex-col items-center gap-2">
            <ScanLine className="h-8 w-8 text-slate-300" />
            <span className="text-xs text-slate-400">支付功能即将开放</span>
          </div>
        </div>
        <span className="text-[13px] text-slate-500">
          {PAYMENT_METHOD_LABELS[selectedMethod]}
          {" — 扫码支付"}
        </span>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create payment page**

Create `dx-web/src/app/(web)/purchase/payment/[orderId]/page.tsx`:

```tsx
import { OrderPayment } from "@/features/web/purchase/components/order-payment";

export default async function PaymentPage({
  params,
}: {
  params: Promise<{ orderId: string }>;
}) {
  const { orderId } = await params;

  return (
    <div className="flex w-full flex-1 items-center justify-center px-4 py-10">
      <OrderPayment orderId={orderId} />
    </div>
  );
}
```

- [ ] **Step 3: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: No errors.

- [ ] **Step 4: Commit**

```bash
cd dx-web
git add src/features/web/purchase/components/order-payment.tsx src/app/\(web\)/purchase/payment/\[orderId\]/page.tsx
git commit -m "feat: add /purchase/payment/[orderId] page with order summary and QR placeholder"
```

---

## Task 11: Update References + Cleanup

**Files:**
- Modify: `dx-web/src/components/in/insufficient-beans-dialog.tsx:55`
- Modify: `dx-web/src/features/web/me/components/membership-block.tsx:19`
- Modify: `dx-web/src/features/web/hall/components/ad-cards-row.tsx:21`
- Modify: `dx-web/src/features/web/hall/components/hall-sidebar.tsx:114`
- Modify: `dx-web/src/features/web/auth/components/user-profile-menu.tsx:46`

- [ ] **Step 1: Update insufficient-beans-dialog.tsx**

Line 55: Change `"/recharge"` to `"/purchase/beans"`:

```tsx
            onClick={() => router.push("/purchase/beans")}
```

- [ ] **Step 2: Update membership-block.tsx**

Line 19: Change `"/auth/membership"` to `"/purchase/membership"`:

```tsx
          href="/purchase/membership"
```

- [ ] **Step 3: Update ad-cards-row.tsx**

Line 21: Change `"/auth/membership"` to `"/purchase/membership"`:

```tsx
          href="/purchase/membership"
```

- [ ] **Step 4: Update hall-sidebar.tsx**

Line 114: Change `"/auth/membership"` to `"/purchase/membership"`:

```tsx
    href: "/purchase/membership",
```

- [ ] **Step 5: Update user-profile-menu.tsx**

Line 46: Change `"/auth/membership"` to `"/purchase/membership"`:

```tsx
      { label: "升级会员", icon: Crown, href: "/purchase/membership" },
```

- [ ] **Step 6: Verify lint**

Run: `cd dx-web && npx next lint`
Expected: No errors.

- [ ] **Step 7: Verify build**

Run: `cd dx-web && npm run build`
Expected: Build succeeds with no errors.

- [ ] **Step 8: Commit**

```bash
cd dx-web
git add src/components/in/insufficient-beans-dialog.tsx src/features/web/me/components/membership-block.tsx src/features/web/hall/components/ad-cards-row.tsx src/features/web/hall/components/hall-sidebar.tsx src/features/web/auth/components/user-profile-menu.tsx
git commit -m "fix: update all route references from /auth/membership and /recharge to /purchase/*"
```

---

## Task 12: Final Verification

- [ ] **Step 1: Backend full build**

Run: `cd dx-api && go build ./...`
Expected: Clean build, no errors.

- [ ] **Step 2: Frontend lint**

Run: `cd dx-web && npx next lint`
Expected: No lint errors.

- [ ] **Step 3: Frontend build**

Run: `cd dx-web && npm run build`
Expected: Clean build, all pages compile.

- [ ] **Step 4: Verify no broken references**

Run from project root:
```bash
grep -r "auth/membership" dx-web/src/ --include="*.tsx" --include="*.ts"
grep -r '"/recharge"' dx-web/src/ --include="*.tsx" --include="*.ts"
```
Expected: No results for either command.

- [ ] **Step 5: Verify new routes exist**

```bash
ls dx-web/src/app/\(web\)/purchase/layout.tsx
ls dx-web/src/app/\(web\)/purchase/membership/page.tsx
ls dx-web/src/app/\(web\)/purchase/beans/page.tsx
ls dx-web/src/app/\(web\)/purchase/payment/\[orderId\]/page.tsx
```
Expected: All 4 files exist.

- [ ] **Step 6: Verify old routes removed**

```bash
ls dx-web/src/app/\(web\)/auth/membership/ 2>&1
ls dx-web/src/app/\(web\)/recharge/ 2>&1
```
Expected: Both directories do not exist.
