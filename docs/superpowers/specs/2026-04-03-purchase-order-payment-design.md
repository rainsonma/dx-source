# Purchase, Order & Payment System Design

## Overview

Add a purchase flow for membership subscriptions and energy bean packages. Users select a product, create an order, and pay via a shared payment page. Payment callbacks (WeChat Pay / Alipay) mark orders as paid and fulfill them (upgrade membership or credit beans).

Payment controller stubs only — no payment license yet.

## Decisions

- Single `orders` table with `type` discriminator (`membership` / `beans`)
- Prices stored in fen (integer) to avoid float math
- Code-level FK constraints (no DB-level FKs, per project convention for PostgreSQL partitions)
- UUID-based IDs
- 30-minute order expiry for unpaid orders
- Order lifecycle: `pending` -> `paid` -> `fulfilled` (also `expired`, `cancelled`)
- Single shared `/purchase/payment/[orderId]` page for both order types
- Light theme for all purchase pages, new layout with back button

## Backend

### Constants (`dx-api/app/consts/`)

**`order_type.go`**

```go
const (
    OrderTypeMembership = "membership"
    OrderTypeBeans      = "beans"
)
```

**`order_status.go`**

```go
const (
    OrderStatusPending   = "pending"
    OrderStatusPaid      = "paid"
    OrderStatusFulfilled = "fulfilled"
    OrderStatusExpired   = "expired"
    OrderStatusCancelled = "cancelled"
)
```

**`payment_method.go`**

```go
const (
    PaymentMethodWechat = "wechat"
    PaymentMethodAlipay = "alipay"
)
```

**`bean_package.go`**

```go
type BeanPackage struct {
    Price int // fen
    Beans int
    Bonus int
}

var BeanPackages = map[string]BeanPackage{
    "beans-1":   {Price: 100, Beans: 1000, Bonus: 0},
    "beans-5":   {Price: 500, Beans: 5000, Bonus: 0},
    "beans-10":  {Price: 1000, Beans: 10000, Bonus: 1000},
    "beans-50":  {Price: 5000, Beans: 50000, Bonus: 7500},
    "beans-100": {Price: 10000, Beans: 100000, Bonus: 20000},
}
```

**New error codes in `error_code.go`**

```go
CodeOrderNotFound   = 40411
CodeOrderNotPending = 40013
CodeInvalidProduct  = 40014
```

### Order Model (`app/models/order.go`)

```go
type Order struct {
    orm.Timestamps
    ID            string           `gorm:"column:id;primaryKey"`
    UserID        string           `gorm:"column:user_id"`
    Type          string           `gorm:"column:type"`
    Product       string           `gorm:"column:product"`
    Amount        int              `gorm:"column:amount"`
    Status        string           `gorm:"column:status"`
    PaymentMethod *string          `gorm:"column:payment_method"`
    PaymentNo     *string          `gorm:"column:payment_no"`
    PaidAt        *carbon.DateTime `gorm:"column:paid_at"`
    FulfilledAt   *carbon.DateTime `gorm:"column:fulfilled_at"`
    ExpiresAt     carbon.DateTime  `gorm:"column:expires_at"`
}
```

Table name: `orders`

### Migration (`database/migrations/YYYYMMDD_create_orders_table.go`)

```
orders:
    id              uuid PK
    user_id         uuid NOT NULL
    type            text NOT NULL DEFAULT ''
    product         text NOT NULL DEFAULT ''
    amount          integer NOT NULL DEFAULT 0
    status          text NOT NULL DEFAULT 'pending'
    payment_method  text NULL
    payment_no      text NULL
    paid_at         timestamptz NULL
    fulfilled_at    timestamptz NULL
    expires_at      timestamptz NOT NULL
    created_at      timestamptz
    updated_at      timestamptz

Indexes:
    - user_id
    - (user_id, status)
    - (status, expires_at)
```

### Order Service (`app/services/api/order_service.go`)

- `CreateOrder(userID, orderType, product, amount, paymentMethod) (*Order, error)` — creates pending order with expires_at = now + 30min
- `GetOrder(orderID, userID) (*OrderDetail, error)` — fetch with ownership check
- `MarkPaid(orderID, paymentNo) error` — sets status=paid, paid_at=now, calls FulfillOrder
- `FulfillOrder(orderID) error` — transactional:
  - Membership: update user grade + vip_due_at + grant beans (reuses calcVipDueAt logic from redeem_service)
  - Beans: credit (base + bonus) beans to user balance + ledger entry
  - Sets status=fulfilled, fulfilled_at=now
- `ExpireStaleOrders() error` — bulk update pending orders past expires_at

New bean slug + reason for purchased beans:
- `BeanSlugPurchaseGrant = "purchase-grant"` in `bean_slug.go`
- `BeanReasonPurchaseGrant = "能量豆充值"` in `bean_reason.go`

### Error Sentinels (`app/services/api/errors.go`)

```go
ErrOrderNotFound   = errors.New("订单不存在")
ErrOrderNotPending = errors.New("订单状态异常")
ErrInvalidProduct  = errors.New("无效的商品")
```

### Order Controller (`app/http/controllers/api/order_controller.go`)

- `CreateMembershipOrder` `POST /api/orders/membership`
  - Input: `{ grade, paymentMethod }`
  - Validates grade is not "free", validates payment method
  - Price from `consts.UserGradePrices` (converted to fen)
  - Creates order, returns `{ id, type, product, amount, status, expiresAt }`

- `CreateBeansOrder` `POST /api/orders/beans`
  - Input: `{ package, paymentMethod }`
  - Validates package exists in `consts.BeanPackages`
  - Creates order, returns same shape

### Payment Controller (`app/http/controllers/api/payment_controller.go`)

- `GetOrder` `GET /api/orders/{id}` (protected)
  - Returns order details with ownership check

- `WechatCallback` `POST /api/payments/callback/wechat` (public)
  - Stub — returns success

- `AlipayCallback` `POST /api/payments/callback/alipay` (public)
  - Stub — returns success

### Request Validation (`app/http/requests/api/`)

**`order_request.go`**

- `CreateMembershipOrderRequest` — validates grade (enum: month/season/year/lifetime), paymentMethod (enum: wechat/alipay)
- `CreateBeansOrderRequest` — validates package (enum: beans-1/beans-5/beans-10/beans-50/beans-100), paymentMethod (enum: wechat/alipay)

### Routes (`routes/api.go`)

```go
// Public payment callbacks (no JWT)
router.Post("/payments/callback/wechat", paymentController.WechatCallback)
router.Post("/payments/callback/alipay", paymentController.AlipayCallback)

// Protected order routes
protected.Post("/orders/membership", orderController.CreateMembershipOrder)
protected.Post("/orders/beans", orderController.CreateBeansOrder)
protected.Get("/orders/{id}", paymentController.GetOrder)
```

### Scheduled Task

`app:expire-stale-orders` — runs every 5 minutes via `bootstrap/app.go`, calls `order_service.ExpireStaleOrders()`.

## Frontend

### New Feature (`src/features/web/purchase/`)

```
src/features/web/purchase/
├── components/
│   ├── pricing-grid.tsx        — moved from auth feature
│   ├── pricing-card.tsx        — moved from auth feature
│   ├── bean-package-grid.tsx   — 5 bean tier cards
│   ├── bean-package-card.tsx   — single bean card
│   ├── order-payment.tsx       — order summary + agreement + method + QR
│   ├── testimonials-grid.tsx   — moved from auth feature
│   ├── testimonial-card.tsx    — moved from auth feature
│   └── faq-section.tsx         — moved from auth feature
├── actions/
│   ├── create-order.ts         — server actions: createMembershipOrder, createBeansOrder
│   └── get-order.ts            — server action: getOrder
├── types/
│   └── order.types.ts          — Order, OrderStatus, OrderType, PaymentMethod types
└── helpers/
    └── bean-packages.ts        — bean package data mirroring backend
```

### Route Structure

```
src/app/(web)/purchase/
├── layout.tsx                  — back button + light background + centered max-w container
├── membership/
│   └── page.tsx                — PricingGrid + TestimonialsGrid + FaqSection
├── beans/
│   └── page.tsx                — BeanPackageGrid + usage info section
└── payment/
    └── [orderId]/
        └── page.tsx            — OrderPayment component
```

### Updated Constants (`src/consts/`)

**`bean-slug.ts`** — add `PURCHASE_GRANT: "purchase-grant"`

**`bean-reason.ts`** — add `PURCHASE_GRANT: "能量豆充值"`

### New Constants (`src/consts/`)

**`bean-package.ts`** — 5 packages with slug, beans, bonus, price (fen), optional tag

**`order-type.ts`** — `ORDER_TYPES.MEMBERSHIP`, `ORDER_TYPES.BEANS`

**`order-status.ts`** — `ORDER_STATUSES.PENDING`, `.PAID`, `.FULFILLED`, `.EXPIRED`, `.CANCELLED`

**`payment-method.ts`** — `PAYMENT_METHODS.WECHAT`, `.ALIPAY`

### Deleted Routes

- `src/app/(web)/auth/membership/` — entire directory (page.tsx, pay/confirm, pay/[method])
- `src/app/(web)/recharge/` — replaced by /purchase/beans

### Updated References

| File | Old | New |
|------|-----|-----|
| `src/components/in/insufficient-beans-dialog.tsx` | `/recharge` | `/purchase/beans` |
| `src/features/web/me/components/membership-block.tsx` | `/auth/membership` | `/purchase/membership` |
| `src/features/web/hall/components/ad-cards-row.tsx` | `/auth/membership` | `/purchase/membership` |
| `src/features/web/hall/components/hall-sidebar.tsx` | `/auth/membership` | `/purchase/membership` |
| `src/features/web/auth/components/user-profile-menu.tsx` | `/auth/membership` | `/purchase/membership` |

### Purchase Layout

- Back button (top-left) via `router.back()`
- Light background, no gradient
- Centered content, max-width container
- No AuthHeader

### Membership Page (`/purchase/membership`)

- Same as current `/auth/membership`: title, current plan badge, PricingGrid, TestimonialsGrid, FaqSection
- "立即订阅" button on each paid tier creates a membership order via server action
- On success, `router.push("/purchase/payment/[orderId]")`
- Free tier button disabled, shows "当前方案"

### Beans Page (`/purchase/beans`)

- Title: "能量豆充值", subtitle: "选择适合您的能量豆套餐"
- BeanPackageGrid: 5 cards in horizontal scroll (mobile) / flex row (desktop)
- Each card: bean icon, amount, bonus line if applicable, price, "立即购买" button
- Tags: "超值推荐" on beans-10, "最划算" on beans-100
- Below grid: "能量豆用途" info section
- "立即购买" creates beans order via server action, redirects to payment page

### Payment Page (`/purchase/payment/[orderId]`)

Single-page view:
- Order summary: order number, product description, amount
- Agreement checkbox: "我已阅读并同意《斗学服务协议》"
- Payment method: WeChat / Alipay radio selection
- QR code area: placeholder until payment integration
- Expiry countdown from `expiresAt`
- Status handling: expired/cancelled/fulfilled shows appropriate message instead of payment UI

### API Client Updates (`src/lib/api-client.ts`)

Add `orderApi` with:
- `createMembershipOrder(data)` — POST `/api/orders/membership`
- `createBeansOrder(data)` — POST `/api/orders/beans`
- `getOrder(id)` — GET `/api/orders/{id}`
