package repository

import (
	"backend/internal/model"
	"strings"

	"gorm.io/gorm"
)

type CustomerVisitRepository struct {
	db *gorm.DB
}

func NewCustomerVisitRepository(db *gorm.DB) *CustomerVisitRepository {
	return &CustomerVisitRepository{db: db}
}

// Create 创建上门拜访记录
func (r *CustomerVisitRepository) Create(input model.CustomerVisitCreateInput) (int64, error) {
	visit := model.CustomerVisit{
		OperatorUserID: input.OperatorUserID,
		CustomerName:   input.CustomerName,
		Inviter:        input.Inviter,
		CheckInIP:      input.CheckInIP,
		CheckInLat:     input.CheckInLat,
		CheckInLng:     input.CheckInLng,
		Province:       input.Province,
		City:           input.City,
		Area:           input.Area,
		DetailAddress:  input.DetailAddress,
		Images:         input.Images,
		VisitPurpose:   input.VisitPurpose,
		Remark:         input.Remark,
		VisitDate:      input.VisitDate,
	}

	if err := r.db.Create(&visit).Error; err != nil {
		return 0, err
	}
	return visit.ID, nil
}

func (r *CustomerVisitRepository) ExistsSameDayVisitByUserCompanyIP(
	operatorUserID int64,
	customerName, checkInIP, visitDate string,
) (bool, error) {
	customerName = strings.TrimSpace(customerName)
	checkInIP = strings.TrimSpace(checkInIP)
	visitDate = strings.TrimSpace(visitDate)
	if operatorUserID == 0 || customerName == "" || checkInIP == "" || visitDate == "" {
		return false, nil
	}

	var count int64
	err := r.db.Table("customer_visits").
		Where("operator_user_id = ? AND customer_name = ? AND check_in_ip = ? AND visit_date = ?",
			operatorUserID, customerName, checkInIP, visitDate).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// List 获取上门拜访列表
func (r *CustomerVisitRepository) List(filter model.CustomerVisitListFilter) (model.CustomerVisitListResult, error) {
	var result model.CustomerVisitListResult

	query := r.db.Model(&model.CustomerVisit{})

	// 没有全量查看权限时只能看自己的记录
	if !filter.CanViewAll {
		query = query.Where("operator_user_id = ?", filter.OperatorUserID)
	}

	// 关键词搜索
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("customer_name LIKE ? OR detail_address LIKE ? OR visit_purpose LIKE ? OR remark LIKE ?",
			keyword, keyword, keyword, keyword)
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	// 获取总数
	if err := query.Count(&result.Total).Error; err != nil {
		return result, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	var visits []model.CustomerVisit
	if err := query.
		Preload("Operator").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&visits).Error; err != nil {
		return result, err
	}

	// 填充操作人名称
	for i := range visits {
		if visits[i].Operator != nil {
			visits[i].OperatorUserName = visits[i].Operator.Nickname
			if visits[i].OperatorUserName == "" {
				visits[i].OperatorUserName = visits[i].Operator.Username
			}
		}
	}

	result.Items = visits
	result.Page = page
	result.PageSize = pageSize

	return result, nil
}
