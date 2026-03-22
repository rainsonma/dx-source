package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"github.com/goravel/framework/facades"
)

type M20260322000014CreateUserReferralsTable struct{}

func (r *M20260322000014CreateUserReferralsTable) Signature() string {
	return "20260322000014_create_user_referrals_table"
}

func (r *M20260322000014CreateUserReferralsTable) Up() error {
	if !facades.Schema().HasTable("user_referrals") {
		return facades.Schema().Create("user_referrals", func(table schema.Blueprint) {
			table.Text("id")
			table.Primary("id")
			table.Text("referrer_id")
			table.Text("invitee_id").Nullable()
			table.Text("status").Default("")
			table.Double("reward_amount").Default(0)
			table.TimestampTz("rewarded_at").Nullable()
			table.TimestampsTz()
			table.Index("referrer_id")
			table.Index("invitee_id")
			table.Index("status")
			table.Index("created_at")
		})
	}
	return nil
}

func (r *M20260322000014CreateUserReferralsTable) Down() error {
	return facades.Schema().DropIfExists("user_referrals")
}
