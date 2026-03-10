package model

import (
	"fmt"
	"time"
)

type ActivityLog struct {
	ID         int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID     int64     `json:"userId" gorm:"column:user_id;not null"`
	Action     string    `json:"action" gorm:"column:action;not null"`
	TargetType string    `json:"targetType" gorm:"column:target_type;not null;default:''"`
	TargetID   int64     `json:"targetId" gorm:"column:target_id;not null;default:0"`
	TargetName string    `json:"targetName" gorm:"column:target_name;not null;default:''"`
	Content    string    `json:"content" gorm:"column:content;not null;default:''"`
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at"`
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}

const (
	ActionCreateContract   = "create_contract"
	ActionAuditContract    = "audit_contract"
	ActionCreateCustomer   = "create_customer"
	ActionImportCustomer   = "import_customer"
	ActionClaimCustomer    = "claim_customer"
	ActionReleaseCustomer  = "release_customer"
	ActionTransferCustomer = "transfer_customer"
	ActionSalesFollow      = "sales_follow"
	ActionOperationFollow  = "operation_follow"

	TargetTypeContract = "contract"
	TargetTypeCustomer = "customer"
)

func ActivityLogNotificationKey(log ActivityLog) string {
	return fmt.Sprintf("activity-%d", log.ID)
}
