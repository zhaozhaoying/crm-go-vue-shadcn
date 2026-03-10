package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/util"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	service       service.CustomerService
	importService service.CustomerImportService
}

type TransferCustomerRequest struct {
	ToOwnerUserID int64 `json:"toOwnerUserId" binding:"required"`
}

type CustomerPhoneInputRequest struct {
	Phone      string `json:"phone" binding:"required"`
	PhoneLabel string `json:"phoneLabel"`
	IsPrimary  bool   `json:"isPrimary"`
}

type CreateCustomerRequest struct {
	Name          string                      `json:"name" binding:"required"`
	LegalName     string                      `json:"legalName"`
	ContactName   string                      `json:"contactName"`
	Weixin        string                      `json:"weixin"`
	Email         string                      `json:"email"`
	Province      int                         `json:"province"`
	City          int                         `json:"city"`
	Area          int                         `json:"area"`
	DetailAddress string                      `json:"detailAddress"`
	Remark        string                      `json:"remark"`
	Status        string                      `json:"status"`
	OwnerUserID   *int64                      `json:"ownerUserId"`
	Phones        []CustomerPhoneInputRequest `json:"phones"`
}

type UpdateCustomerRequest struct {
	Name          string                      `json:"name" binding:"required"`
	LegalName     string                      `json:"legalName"`
	ContactName   string                      `json:"contactName"`
	Weixin        string                      `json:"weixin"`
	Email         string                      `json:"email"`
	Province      int                         `json:"province"`
	City          int                         `json:"city"`
	Area          int                         `json:"area"`
	DetailAddress string                      `json:"detailAddress"`
	Remark        string                      `json:"remark"`
	Phones        []CustomerPhoneInputRequest `json:"phones"`
}

type CheckCustomerUniqueRequest struct {
	ExcludeCustomerID *int64   `json:"excludeCustomerId"`
	Name              string   `json:"name"`
	LegalName         string   `json:"legalName"`
	Weixin            string   `json:"weixin"`
	Phones            []string `json:"phones"`
}

type AddPhoneRequest struct {
	Phone      string `json:"phone" binding:"required"`
	PhoneLabel string `json:"phoneLabel"`
	IsPrimary  bool   `json:"isPrimary"`
}

type UpdatePhoneRequest struct {
	Phone      string `json:"phone" binding:"required"`
	PhoneLabel string `json:"phoneLabel"`
	IsPrimary  bool   `json:"isPrimary"`
}

type CreateStatusLogRequest struct {
	ToStatus int    `json:"toStatus" binding:"required"`
	Reason   string `json:"reason"`
}

func NewCustomerHandler(
	service service.CustomerService,
	importService service.CustomerImportService,
) *CustomerHandler {
	return &CustomerHandler{
		service:       service,
		importService: importService,
	}
}

// List godoc
// @Summary     获取客户列表
// @Description 获取客户信息（需要认证）
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       name query string false "客户名称"
// @Param       contactName query string false "联系人"
// @Param       phone query string false "手机号"
// @Param       weixin query string false "微信"
// @Param       ownerUserName query string false "负责人"
// @Param       province query int false "省编码"
// @Param       city query int false "市编码"
// @Param       area query int false "区编码"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/customers [get]
func (h *CustomerHandler) List(c *gin.Context) {
	h.listByCategory(c, c.Query("category"))
}

// ListMy godoc
// @Summary     获取我的客户列表
// @Description 获取当前用户权限范围内的我的客户
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       ownershipScope query string false "查看范围(all/mine/subordinates/sales)"
// @Param       keyword query string false "关键词"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/customers/my [get]
func (h *CustomerHandler) ListMy(c *gin.Context) {
	h.listByCategory(c, "my")
}

// ListPool godoc
// @Summary     获取公海客户列表
// @Description 获取公海客户信息
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       keyword query string false "关键词"
// @Param       sortBy query string false "排序字段(dropTime/followTime/updatedAt)"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/customers/pool [get]
func (h *CustomerHandler) ListPool(c *gin.Context) {
	h.listByCategory(c, "pool")
}

// ListPotential godoc
// @Summary     获取潜在客户列表
// @Description 获取当前用户历史接触过的公海潜在客户
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/customers/potential [get]
func (h *CustomerHandler) ListPotential(c *gin.Context) {
	h.listByCategory(c, "potential")
}

// ListPartner godoc
// @Summary     获取合作客户列表
// @Description 获取当前用户已成交客户
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/customers/partner [get]
func (h *CustomerHandler) ListPartner(c *gin.Context) {
	h.listByCategory(c, "partner")
}

// ListSearch godoc
// @Summary     获取查找客户列表
// @Description 获取客户搜索结果，联系方式会按权限掩码
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       keyword query string false "关键词"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/customers/search [get]
func (h *CustomerHandler) ListSearch(c *gin.Context) {
	h.listByCategory(c, "search")
}

func (h *CustomerHandler) listByCategory(c *gin.Context, category string) {
	viewerID, hasViewer := currentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	filter := model.CustomerListFilter{
		Category:       strings.TrimSpace(category),
		OwnershipScope: parseOwnershipScopeQuery(c),
		Keyword:        strings.TrimSpace(c.Query("keyword")),
		Name:           strings.TrimSpace(c.Query("name")),
		ContactName:    strings.TrimSpace(c.Query("contactName")),
		Phone:          strings.TrimSpace(c.Query("phone")),
		Weixin:         strings.TrimSpace(c.Query("weixin")),
		OwnerUserName:  strings.TrimSpace(c.Query("ownerUserName")),
		Province:       parseCodeQuery(c.Query("province")),
		City:           parseCodeQuery(c.Query("city")),
		Area:           parseCodeQuery(c.Query("area")),
		ExcludePool:    parseBoolQuery(c.Query("excludePool")),
		SortBy:         strings.TrimSpace(c.Query("sortBy")),
		Page:           page,
		PageSize:       pageSize,
		ViewerID:       viewerID,
		HasViewer:      hasViewer,
		ActorRole:      currentUserRole(c),
	}

	result, err := h.service.ListCustomers(c.Request.Context(), filter)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10001, "failed to list customers")
		return
	}

	// Apply masking for search category
	if category == "search" {
		result.Items = maskCustomerData(result.Items)
	}

	Success(c, result)
}

func parseOwnershipScopeQuery(c *gin.Context) string {
	scope := strings.TrimSpace(c.Query("ownershipScope"))
	if scope != "" {
		return scope
	}
	// Backward compatibility for clients that still send `scope`.
	return strings.TrimSpace(c.Query("scope"))
}

func parseCodeQuery(raw string) int {
	value := strings.TrimSpace(raw)
	if value == "" || value == "all" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseBoolQuery(raw string) bool {
	value := strings.TrimSpace(strings.ToLower(raw))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

// Create godoc
// @Summary     创建客户
// @Description 创建一个新客户
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body CreateCustomerRequest true "客户信息"
// @Success     200 {object} APIResponse{data=model.Customer}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     409 {object} APIResponse
// @Router      /api/v1/customers [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10031, "invalid request body")
		return
	}

	customer, err := h.service.CreateCustomer(c.Request.Context(), model.CustomerCreateInput{
		Name:           req.Name,
		LegalName:      req.LegalName,
		ContactName:    req.ContactName,
		Weixin:         req.Weixin,
		Email:          req.Email,
		Province:       req.Province,
		City:           req.City,
		Area:           req.Area,
		DetailAddress:  req.DetailAddress,
		Remark:         req.Remark,
		Status:         req.Status,
		OwnerUserID:    req.OwnerUserID,
		OperatorUserID: userID,
		Phones:         toCustomerPhoneInputs(req.Phones),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerNameRequired):
			Error(c, http.StatusBadRequest, 10032, "customer name is required")
		case errors.Is(err, service.ErrInvalidPhoneFormat):
			Error(c, http.StatusBadRequest, 10021, "invalid phone format")
		case errors.Is(err, service.ErrCustomerNameExists):
			Error(c, http.StatusConflict, 10033, "customer name already exists")
		case errors.Is(err, service.ErrCustomerLegalExists):
			Error(c, http.StatusConflict, 10034, "customer legal name already exists")
		case errors.Is(err, service.ErrCustomerWeixinExists):
			Error(c, http.StatusConflict, 10035, "customer weixin already exists")
		case errors.Is(err, service.ErrCustomerPhoneExists):
			Error(c, http.StatusConflict, 10022, "phone already exists for this customer")
		case errors.Is(err, service.ErrCustomerLimitExceeded):
			Error(c, http.StatusConflict, 10038, "个人客户池已达上限，已成交客户不计入")
		default:
			Error(c, http.StatusInternalServerError, 10036, "failed to create customer")
		}
		return
	}

	Success(c, customer)
}

// Update godoc
// @Summary     更新客户
// @Description 更新指定客户的基础信息
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Param       body body UpdateCustomerRequest true "客户信息"
// @Success     200 {object} APIResponse{data=model.Customer}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/customers/{id} [put]
func (h *CustomerHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10031, "invalid request body")
		return
	}

	customer, err := h.service.UpdateCustomer(c.Request.Context(), customerID, model.CustomerUpdateInput{
		Name:           req.Name,
		LegalName:      req.LegalName,
		ContactName:    req.ContactName,
		Weixin:         req.Weixin,
		Email:          req.Email,
		Province:       req.Province,
		City:           req.City,
		Area:           req.Area,
		DetailAddress:  req.DetailAddress,
		Remark:         req.Remark,
		OperatorUserID: userID,
		Phones:         toCustomerPhoneInputs(req.Phones),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerNotFound):
			Error(c, http.StatusNotFound, 10003, "customer not found")
		case errors.Is(err, service.ErrCustomerNameRequired):
			Error(c, http.StatusBadRequest, 10032, "customer name is required")
		case errors.Is(err, service.ErrInvalidPhoneFormat):
			Error(c, http.StatusBadRequest, 10021, "invalid phone format")
		case errors.Is(err, service.ErrCustomerNameExists):
			Error(c, http.StatusConflict, 10033, "customer name already exists")
		case errors.Is(err, service.ErrCustomerLegalExists):
			Error(c, http.StatusConflict, 10034, "customer legal name already exists")
		case errors.Is(err, service.ErrCustomerWeixinExists):
			Error(c, http.StatusConflict, 10035, "customer weixin already exists")
		case errors.Is(err, service.ErrCustomerPhoneExists):
			Error(c, http.StatusConflict, 10022, "phone already exists for this customer")
		default:
			Error(c, http.StatusInternalServerError, 10037, "failed to update customer")
		}
		return
	}

	Success(c, customer)
}

// CheckUnique godoc
// @Summary     校验客户唯一性
// @Description 校验客户名称、法人、微信和电话是否重复
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body CheckCustomerUniqueRequest true "唯一性校验条件"
// @Success     200 {object} APIResponse{data=model.CustomerUniqueCheckResult}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/customers/validate-unique [post]
func (h *CustomerHandler) CheckUnique(c *gin.Context) {
	var req CheckCustomerUniqueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10031, "invalid request body")
		return
	}

	result, err := h.service.CheckUnique(c.Request.Context(), model.CustomerUniqueCheckInput{
		ExcludeCustomerID: req.ExcludeCustomerID,
		Name:              req.Name,
		LegalName:         req.LegalName,
		Weixin:            req.Weixin,
		Phones:            req.Phones,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, 10038, "failed to check uniqueness")
		return
	}

	Success(c, result)
}

// ImportCSV godoc
// @Summary     批量导入客户(CSV)
// @Description 上传CSV文件导入客户，支持十万级数据批量处理；支持dryRun预演
// @Tags        customers
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       file formData file true "CSV文件(需包含表头，至少: name, phone)"
// @Param       batchSize formData int false "批大小，默认1000，最大5000"
// @Param       dryRun formData bool false "是否仅校验不入库，默认false"
// @Param       defaultStatus formData string false "默认客户状态(owned/pool)，默认owned"
// @Param       maxErrors formData int false "最多返回多少条错误明细，默认200"
// @Success     200 {object} APIResponse{data=model.CustomerImportResult}
// @Router      /api/v1/customers/import/csv [post]
func (h *CustomerHandler) ImportCSV(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	if h.importService == nil {
		Error(c, http.StatusNotImplemented, 10039, "customer import service is not configured")
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 10040, "file is required")
		return
	}
	defer file.Close()

	batchSize, _ := strconv.Atoi(strings.TrimSpace(c.DefaultPostForm("batchSize", "1000")))
	maxErrors, _ := strconv.Atoi(strings.TrimSpace(c.DefaultPostForm("maxErrors", "200")))
	defaultStatus := strings.TrimSpace(c.DefaultPostForm("defaultStatus", model.CustomerStatusOwned))
	dryRun := parseBoolQuery(c.DefaultPostForm("dryRun", "false"))

	report, err := h.importService.ImportCSV(c.Request.Context(), file, service.CustomerCSVImportInput{
		OperatorUserID: userID,
		BatchSize:      batchSize,
		DryRun:         dryRun,
		DefaultStatus:  defaultStatus,
		MaxErrors:      maxErrors,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerImportInvalidFile):
			Error(c, http.StatusBadRequest, 10041, "invalid csv file")
		case errors.Is(err, service.ErrCustomerImportInvalidHeader):
			Error(c, http.StatusBadRequest, 10042, "invalid csv header, required columns: name, phone")
		default:
			Error(c, http.StatusInternalServerError, 10043, "failed to import customers: "+err.Error())
		}
		return
	}

	Success(c, report)
}

// Claim godoc
// @Summary     领取公海客户
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Success     200 {object} APIResponse{data=model.Customer}
// @Router      /api/v1/customers/{id}/claim [post]
func (h *CustomerHandler) Claim(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	customer, err := h.service.ClaimCustomer(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			Error(c, http.StatusNotFound, 10003, "customer not found")
			return
		}
		if errors.Is(err, service.ErrCustomerNotInPool) {
			Error(c, http.StatusConflict, 10004, "customer is not in pool")
			return
		}
		if errors.Is(err, service.ErrCustomerLimitExceeded) {
			Error(c, http.StatusConflict, 10038, "个人客户池已达上限，已成交客户不计入")
			return
		}
		if errors.Is(err, service.ErrCustomerSameDepartmentClaimForbidden) {
			Error(c, http.StatusForbidden, 10044, "同部门客户不可领取")
			return
		}
		Error(c, http.StatusInternalServerError, 10005, "failed to claim customer")
		return
	}
	Success(c, customer)
}

// Release godoc
// @Summary     释放客户到公海
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Success     200 {object} APIResponse{data=model.Customer}
// @Router      /api/v1/customers/{id}/release [post]
func (h *CustomerHandler) Release(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	customer, err := h.service.ReleaseCustomer(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			Error(c, http.StatusNotFound, 10003, "customer not found")
			return
		}
		if errors.Is(err, service.ErrCustomerAlreadyInPool) {
			Error(c, http.StatusConflict, 10006, "customer is already in pool")
			return
		}
		if errors.Is(err, service.ErrCustomerNotOwned) {
			Error(c, http.StatusForbidden, 10007, "customer is not owned by current user")
			return
		}
		Error(c, http.StatusInternalServerError, 10008, "failed to release customer")
		return
	}
	Success(c, customer)
}

// Transfer godoc
// @Summary     转移客户负责人
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Param       body body TransferCustomerRequest true "转移信息"
// @Success     200 {object} APIResponse{data=model.Customer}
// @Router      /api/v1/customers/{id}/transfer [post]
func (h *CustomerHandler) Transfer(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	var req TransferCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10009, "invalid request body")
		return
	}

	customer, err := h.service.TransferCustomer(c.Request.Context(), model.CustomerTransferInput{
		CustomerID:     id,
		ToOwnerUserID:  req.ToOwnerUserID,
		OperatorUserID: userID,
	})
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			Error(c, http.StatusNotFound, 10003, "customer not found")
			return
		}
		if errors.Is(err, service.ErrCustomerNotOwned) {
			Error(c, http.StatusForbidden, 10007, "customer is not owned by current user")
			return
		}
		if errors.Is(err, service.ErrCustomerLimitExceeded) {
			Error(c, http.StatusConflict, 10038, "个人客户池已达上限，已成交客户不计入")
			return
		}
		Error(c, http.StatusInternalServerError, 10010, "failed to transfer customer")
		return
	}
	Success(c, customer)
}

// AddPhone godoc
// @Summary     添加客户电话
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Param       body body AddPhoneRequest true "电话信息"
// @Success     200 {object} APIResponse{data=model.CustomerPhone}
// @Router      /api/v1/customers/{id}/phones [post]
func (h *CustomerHandler) AddPhone(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	var req AddPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10020, "invalid request body")
		return
	}

	phone := &model.CustomerPhone{
		CustomerID: customerID,
		Phone:      req.Phone,
		PhoneLabel: req.PhoneLabel,
		IsPrimary:  req.IsPrimary,
	}

	if err := h.service.AddPhone(c.Request.Context(), phone); err != nil {
		if errors.Is(err, service.ErrInvalidPhoneFormat) {
			Error(c, http.StatusBadRequest, 10021, "invalid phone format")
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			Error(c, http.StatusConflict, 10022, "phone already exists for this customer")
			return
		}
		Error(c, http.StatusInternalServerError, 10023, "failed to add phone")
		return
	}

	Success(c, phone)
}

// ListPhones godoc
// @Summary     获取客户电话列表
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Success     200 {object} APIResponse{data=[]model.CustomerPhone}
// @Router      /api/v1/customers/{id}/phones [get]
func (h *CustomerHandler) ListPhones(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	phones, err := h.service.ListPhones(c.Request.Context(), customerID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10024, "failed to list phones")
		return
	}

	Success(c, phones)
}

// UpdatePhone godoc
// @Summary     更新客户电话
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Param       phoneId path int true "电话ID"
// @Param       body body UpdatePhoneRequest true "电话信息"
// @Success     200 {object} APIResponse{data=model.CustomerPhone}
// @Router      /api/v1/customers/{id}/phones/{phoneId} [put]
func (h *CustomerHandler) UpdatePhone(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}
	phoneID, err := strconv.ParseInt(c.Param("phoneId"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10025, "invalid phone id")
		return
	}

	var req UpdatePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10020, "invalid request body")
		return
	}

	phone := &model.CustomerPhone{
		ID:         phoneID,
		CustomerID: customerID,
		Phone:      req.Phone,
		PhoneLabel: req.PhoneLabel,
		IsPrimary:  req.IsPrimary,
	}

	if err := h.service.UpdatePhone(c.Request.Context(), phone); err != nil {
		if errors.Is(err, service.ErrPhoneNotFound) {
			Error(c, http.StatusNotFound, 10026, "phone not found")
			return
		}
		if errors.Is(err, service.ErrInvalidPhoneFormat) {
			Error(c, http.StatusBadRequest, 10021, "invalid phone format")
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			Error(c, http.StatusConflict, 10022, "phone already exists for this customer")
			return
		}
		Error(c, http.StatusInternalServerError, 10027, "failed to update phone")
		return
	}

	Success(c, phone)
}

// DeletePhone godoc
// @Summary     删除客户电话
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Param       phoneId path int true "电话ID"
// @Success     200 {object} APIResponse
// @Router      /api/v1/customers/{id}/phones/{phoneId} [delete]
func (h *CustomerHandler) DeletePhone(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}
	phoneID, err := strconv.ParseInt(c.Param("phoneId"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10025, "invalid phone id")
		return
	}

	if err := h.service.DeletePhone(c.Request.Context(), customerID, phoneID); err != nil {
		if errors.Is(err, service.ErrPhoneNotFound) {
			Error(c, http.StatusNotFound, 10026, "phone not found")
			return
		}
		Error(c, http.StatusInternalServerError, 10028, "failed to delete phone")
		return
	}

	Success(c, nil)
}

// ListStatusLogs godoc
// @Summary     获取客户状态日志
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Param       page query int false "页码"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=[]model.CustomerStatusLog}
// @Router      /api/v1/customers/{id}/status-logs [get]
func (h *CustomerHandler) ListStatusLogs(c *gin.Context) {
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, err := h.service.ListStatusLogs(c.Request.Context(), customerID, page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 10029, "failed to list status logs")
		return
	}

	Success(c, logs)
}

// CreateStatusLog godoc
// @Summary     创建客户状态日志
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Param       body body CreateStatusLogRequest true "状态日志信息"
// @Success     200 {object} APIResponse{data=model.CustomerStatusLog}
// @Router      /api/v1/customers/{id}/status-logs [post]
func (h *CustomerHandler) CreateStatusLog(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "invalid customer id")
		return
	}

	var req CreateStatusLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 10020, "invalid request body")
		return
	}

	log := &model.CustomerStatusLog{
		CustomerID:     customerID,
		FromStatus:     0,
		ToStatus:       req.ToStatus,
		TriggerType:    0,
		Reason:         req.Reason,
		OperatorUserID: &userID,
	}

	if err := h.service.CreateStatusLog(c.Request.Context(), log); err != nil {
		Error(c, http.StatusInternalServerError, 10030, "failed to create status log")
		return
	}

	Success(c, log)
}

func currentUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}

func toCustomerPhoneInputs(phones []CustomerPhoneInputRequest) []model.CustomerPhoneInput {
	if len(phones) == 0 {
		return nil
	}
	items := make([]model.CustomerPhoneInput, 0, len(phones))
	for _, phone := range phones {
		items = append(items, model.CustomerPhoneInput{
			Phone:      phone.Phone,
			PhoneLabel: phone.PhoneLabel,
			IsPrimary:  phone.IsPrimary,
		})
	}
	return items
}

func maskCustomerData(customers []model.Customer) []model.Customer {
	masked := make([]model.Customer, len(customers))
	for i, customer := range customers {
		masked[i] = customer
		// Mask company name
		masked[i].Name = util.MaskCompanyName(customer.Name)
		// Mask email and owner fields
		masked[i].Email = util.MaskEmail(customer.Email)
		masked[i].OwnerUserName = "*"
		masked[i].OwnerUserID = nil

		// Mask phone numbers
		if len(customer.Phones) > 0 {
			maskedPhones := make([]model.CustomerPhone, len(customer.Phones))
			for j, phone := range customer.Phones {
				maskedPhones[j] = phone
				maskedPhones[j].Phone = util.MaskPhone(phone.Phone)
			}
			masked[i].Phones = maskedPhones
		}
	}
	return masked
}
