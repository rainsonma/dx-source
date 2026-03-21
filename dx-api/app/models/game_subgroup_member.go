package models

import "github.com/goravel/framework/database/orm"

type GameSubgroupMember struct {
	orm.Timestamps
	ID             string `gorm:"column:id;primaryKey" json:"id"`
	GameSubgroupID string `gorm:"column:game_subgroup_id" json:"game_subgroup_id"`
	UserID         string `gorm:"column:user_id" json:"user_id"`
	Role           string `gorm:"column:role" json:"role"`
}

func (g *GameSubgroupMember) TableName() string {
	return "game_subgroup_members"
}
