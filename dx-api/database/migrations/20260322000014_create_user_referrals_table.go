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
			table.String("id")
			table.Primary("id")
			table.String("referrer_id")
			table.String("invitee_id").Nullable()
			table.String("status").Default("")
			table.Double("reward_amount").Default(0)
			table.TimestampTz("rewarded_at").Nullable()
			table.TimestampsTz()
		})
	}
	return nil
}

func (r *M20260322000014CreateUserReferralsTable) Down() error {
	return facades.Schema().DropIfExists("user_referrals")
}
