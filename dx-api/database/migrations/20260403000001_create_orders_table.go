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
