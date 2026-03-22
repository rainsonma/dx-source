package commands

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"dx-api/app/constants"
	"github.com/goravel/framework/facades"
	"dx-api/app/models"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/oklog/ulid/v2"
)

type ResetEnergyBeans struct {
}

// Signature returns the unique signature of the command.
func (c *ResetEnergyBeans) Signature() string {
	return "app:reset-energy-beans"
}

// Description returns the console command description.
func (c *ResetEnergyBeans) Description() string {
	return "Reset energy beans for paid members on their monthly anniversary"
}

// Extend returns the command extend options.
func (c *ResetEnergyBeans) Extend() command.Extend {
	return command.Extend{}
}

// Handle executes the console command.
func (c *ResetEnergyBeans) Handle(ctx console.Context) error {
	start := time.Now()

	now := time.Now()
	todayDay := now.Day()
	lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()

	// Find users whose membership grant anniversary matches today.
	// For day 29/30/31 in shorter months, match on last day of month.
	type eligibleUser struct {
		ID           string     `gorm:"column:id"`
		Beans        int        `gorm:"column:beans"`
		GrantedBeans int        `gorm:"column:granted_beans"`
		Grade        string     `gorm:"column:grade"`
		VipDueAt     *time.Time `gorm:"column:vip_due_at"`
		GrantDay     int        `gorm:"column:grant_day"`
	}

	var users []eligibleUser
	if err := facades.Orm().Query().Raw(`
		SELECT
			u.id,
			u.beans,
			u.granted_beans,
			u.grade,
			u.vip_due_at,
			EXTRACT(DAY FROM ub.first_grant)::int AS grant_day
		FROM users u
		INNER JOIN (
			SELECT user_id, MIN(created_at) AS first_grant
			FROM user_beans
			WHERE slug = ?
			GROUP BY user_id
		) ub ON u.id = ub.user_id
		WHERE u.grade != 'free'
		  AND u.is_active = true
		  AND (
			EXTRACT(DAY FROM ub.first_grant)::int = ?
			OR (
				EXTRACT(DAY FROM ub.first_grant)::int > ?
				AND ? = ?
			)
		  )
	`, constants.BeanSlugMembershipGrant, todayDay, lastDayOfMonth, todayDay, lastDayOfMonth).Scan(&users); err != nil {
		ctx.Error(fmt.Sprintf("failed to query eligible users: %v", err))
		return err
	}

	var debited, credited, skipped int

	for _, user := range users {
		isLifetime := user.Grade == "lifetime"
		isExpired := !isLifetime && (user.VipDueAt == nil || user.VipDueAt.Before(now))
		hasGrantedBeans := user.GrantedBeans > 0

		// Skip: expired + no granted beans left
		if isExpired && !hasGrantedBeans {
			skipped++
			continue
		}

		grantAmount := 10000
		if isLifetime {
			grantAmount = 15000
		}

		if err := resetBeansForUser(user.ID, user.Beans, user.GrantedBeans, user.Grade, grantAmount, hasGrantedBeans, isExpired); err != nil {
			ctx.Error(fmt.Sprintf("failed to reset beans for user %s: %v", user.ID, err))
			continue
		}

		if hasGrantedBeans {
			debited++
		}
		if !isExpired {
			credited++
		}
	}

	elapsed := time.Since(start)
	ctx.Info(fmt.Sprintf(
		"[reset-energy-beans] done in %s — eligible: %d, debited: %d, credited: %d, skipped: %d",
		elapsed, len(users), debited, credited, skipped,
	))
	return nil
}

func resetBeansForUser(userID string, beans, grantedBeans int, grade string, grantAmount int, hasGrantedBeans, isExpired bool) error {
	tx, err := facades.Orm().Query().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	currentBeans := beans
	currentGranted := grantedBeans

	// Debit remaining granted beans
	if hasGrantedBeans {
		debitAmount := -currentGranted
		debitResult := currentBeans + debitAmount
		dataJSON := marshalJSON(map[string]any{"grantedBeansCleared": currentGranted})

		ledger := models.UserBean{
			ID:     newID(),
			UserID: userID,
			Beans:  debitAmount,
			Origin: currentBeans,
			Result: debitResult,
			Slug:   constants.BeanSlugMonthlyResetDebit,
			Reason: constants.BeanReasonMonthlyResetDebit,
			Data:   dataJSON,
		}
		if err := tx.Create(&ledger); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to create debit ledger: %w", err)
		}

		currentBeans = debitResult
		currentGranted = 0
	}

	// Credit new grant (only if membership is active)
	if !isExpired {
		creditResult := currentBeans + grantAmount
		dataJSON := marshalJSON(map[string]any{"gradeAtTime": grade, "grantAmount": grantAmount})

		ledger := models.UserBean{
			ID:     newID(),
			UserID: userID,
			Beans:  grantAmount,
			Origin: currentBeans,
			Result: creditResult,
			Slug:   constants.BeanSlugMonthlyResetCredit,
			Reason: constants.BeanReasonMonthlyResetCredit,
			Data:   dataJSON,
		}
		if err := tx.Create(&ledger); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to create credit ledger: %w", err)
		}

		currentBeans = creditResult
		currentGranted = grantAmount
	}

	// Atomic user balance update
	if _, err := tx.Model(&models.User{}).Where("id", userID).
		Update("beans", currentBeans); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to update user beans: %w", err)
	}
	if _, err := tx.Model(&models.User{}).Where("id", userID).
		Update("granted_beans", currentGranted); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to update user granted_beans: %w", err)
	}

	return tx.Commit()
}

func newID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}

func marshalJSON(v any) *string {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
