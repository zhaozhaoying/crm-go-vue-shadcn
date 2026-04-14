package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/util"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	service       service.CustomerService
	importService service.CustomerImportService
}

type TransferCustomerRequest struct {
	ToOwnerUserID int64  `json:"toOwnerUserId" binding:"required"`
	Note          string `json:"note"`
}

type BatchRankedReassignCustomersRequest struct {
	CustomerIDs []int64 `json:"customerIds" binding:"required"`
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
	ContactName       string   `json:"contactName"`
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
// @Param       ownerUserId query int false "负责人ID"
// @Param       ownerUserName query string false "负责人"
// @Param       province query int false "省编码"
// @Param       city query int false "市编码"
// @Param       area query int false "区编码"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse "未登录或登录已失效"
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
// @Param       ownershipScope query string false "查看范围（全部/我的/下属/销售）"
// @Param       ownerUserId query int false "负责人ID"
// @Param       keyword query string false "关键词"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse "未登录或登录已失效"
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
// @Param       sortBy query string false "排序字段（掉库时间/跟进时间/更新时间）"
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerListResult}
// @Failure     401 {object} APIResponse "未登录或登录已失效"
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
// @Failure     401 {object} APIResponse "未登录或登录已失效"
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
// @Failure     401 {object} APIResponse "未登录或登录已失效"
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
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Router      /api/v1/customers/search [get]
func (h *CustomerHandler) ListSearch(c *gin.Context) {
	h.listByCategory(c, "search")
}

// ListAssignments godoc
// @Summary     获取客户分配列表
// @Description 获取电销分配给销售的客户记录
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       page query int false "页码，从1开始"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.CustomerAssignmentListResult}
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Router      /api/v1/customers/assignments [get]
func (h *CustomerHandler) ListAssignments(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	if !canViewCustomerAssignments(currentUserRole(c)) {
		Error(c, http.StatusForbidden, 10031, "仅管理员或财务经理可以查看客户分配")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	result, err := h.service.ListCustomerAssignments(c.Request.Context(), model.CustomerAssignmentListFilter{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10001, "加载客户分配列表失败", err)
		return
	}

	Success(c, result)
}

func (h *CustomerHandler) BatchRankedReassign(c *gin.Context) {
	operatorUserID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	role := strings.TrimSpace(strings.ToLower(currentUserRole(c)))
	if role != "admin" && role != "管理员" {
		Error(c, http.StatusForbidden, 10031, "仅管理员可以重新分配客户")
		return
	}

	var req BatchRankedReassignCustomersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 40001, "请求参数错误", err)
		return
	}

	result, err := h.service.ReassignCustomersByYesterdayRanking(c.Request.Context(), model.CustomerBatchRankedReassignInput{
		CustomerIDs:    req.CustomerIDs,
		OperatorUserID: operatorUserID,
	})
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 40002, "批量重新分配客户失败", err)
		return
	}

	Success(c, result)
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
		OwnerUserID:    parseInt64Query(c.Query("ownerUserId")),
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
		ErrorWithDetail(c, http.StatusInternalServerError, 10001, "加载客户列表失败", err)
		return
	}

	// Apply masking for search category
	if category == "search" {
		result.Items = maskCustomerData(result.Items, viewerID, hasViewer, filter.ActorRole)
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

func parseInt64Query(raw string) int64 {
	value := strings.TrimSpace(raw)
	if value == "" || value == "all" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
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
// @Failure     400 {object} APIResponse "请求参数错误"
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Failure     409 {object} APIResponse "请求冲突"
// @Router      /api/v1/customers [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10031, "请求参数错误", err)
		return
	}
	req.LegalName = strings.TrimSpace(req.LegalName)
	req.ContactName = strings.TrimSpace(req.ContactName)

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
			Error(c, http.StatusBadRequest, 10032, "客户名称不能为空")
		case errors.Is(err, service.ErrCustomerLegalNameRequired):
			Error(c, http.StatusBadRequest, 10048, "法人不能为空")
		case errors.Is(err, service.ErrCustomerContactNameRequired):
			Error(c, http.StatusBadRequest, 10049, "联系人不能为空")
		case errors.Is(err, service.ErrCustomerLegalNameTooShort):
			Error(c, http.StatusBadRequest, 10050, "法人至少需要2个字")
		case errors.Is(err, service.ErrCustomerContactNameTooShort):
			Error(c, http.StatusBadRequest, 10051, "联系人至少需要2个字")
		case errors.Is(err, service.ErrInvalidPhoneFormat):
			Error(c, http.StatusBadRequest, 10021, "联系电话格式不正确，请输入手机号或座机号")
		case errors.Is(err, service.ErrCustomerNameExists):
			Error(c, http.StatusConflict, 10033, "客户名称已存在")
		case errors.Is(err, service.ErrCustomerWeixinExists):
			Error(c, http.StatusConflict, 10035, "微信号已存在")
		case errors.Is(err, service.ErrCustomerPhoneExists):
			Error(c, http.StatusConflict, 10022, "联系电话已存在")
		case errors.Is(err, service.ErrCustomerNoOutsideSalesAvailable):
			Error(c, http.StatusConflict, 10039, "当前团队下暂无可分配的销售负责人")
		case errors.Is(err, service.ErrCustomerLimitExceeded):
			Error(c, http.StatusConflict, 10038, "个人客户池已达上限，已成交客户不计入")
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 10036, "创建客户失败", err)
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
// @Failure     400 {object} APIResponse "请求参数错误"
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Failure     404 {object} APIResponse "资源不存在"
// @Router      /api/v1/customers/{id} [put]
func (h *CustomerHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10031, "请求参数错误", err)
		return
	}
	req.LegalName = strings.TrimSpace(req.LegalName)
	req.ContactName = strings.TrimSpace(req.ContactName)

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
			Error(c, http.StatusNotFound, 10003, "客户不存在")
		case errors.Is(err, service.ErrCustomerNameRequired):
			Error(c, http.StatusBadRequest, 10032, "客户名称不能为空")
		case errors.Is(err, service.ErrCustomerLegalNameRequired):
			Error(c, http.StatusBadRequest, 10048, "法人不能为空")
		case errors.Is(err, service.ErrCustomerContactNameRequired):
			Error(c, http.StatusBadRequest, 10049, "联系人不能为空")
		case errors.Is(err, service.ErrCustomerLegalNameTooShort):
			Error(c, http.StatusBadRequest, 10050, "法人至少需要2个字")
		case errors.Is(err, service.ErrCustomerContactNameTooShort):
			Error(c, http.StatusBadRequest, 10051, "联系人至少需要2个字")
		case errors.Is(err, service.ErrInvalidPhoneFormat):
			Error(c, http.StatusBadRequest, 10021, "联系电话格式不正确，请输入手机号或座机号")
		case errors.Is(err, service.ErrCustomerNameExists):
			Error(c, http.StatusConflict, 10033, "客户名称已存在")
		case errors.Is(err, service.ErrCustomerWeixinExists):
			Error(c, http.StatusConflict, 10035, "微信号已存在")
		case errors.Is(err, service.ErrCustomerPhoneExists):
			Error(c, http.StatusConflict, 10022, "联系电话已存在")
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 10037, "更新客户失败", err)
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
// @Failure     400 {object} APIResponse "请求参数错误"
// @Failure     401 {object} APIResponse "未登录或登录已失效"
// @Router      /api/v1/customers/validate-unique [post]
func (h *CustomerHandler) CheckUnique(c *gin.Context) {
	var req CheckCustomerUniqueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10031, "请求参数错误", err)
		return
	}

	result, err := h.service.CheckUnique(c.Request.Context(), model.CustomerUniqueCheckInput{
		ExcludeCustomerID: req.ExcludeCustomerID,
		Name:              req.Name,
		LegalName:         req.LegalName,
		ContactName:       req.ContactName,
		Weixin:            req.Weixin,
		Phones:            req.Phones,
	})
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10046, "校验客户唯一性失败", err)
		return
	}

	Success(c, result)
}

// ImportCSV godoc
// @Summary     批量导入客户（CSV）
// @Description 上传 CSV 文件导入客户，支持十万级数据批量处理；支持仅校验不入库预演
// @Tags        customers
// @Accept      multipart/form-data
// @Produce     json
// @Security    BearerAuth
// @Param       file formData file true "CSV 文件（需包含表头，至少包含 name、phone 列）"
// @Param       batchSize formData int false "批大小，默认1000，最大5000"
// @Param       dryRun formData bool false "是否仅校验不入库，默认否"
// @Param       defaultStatus formData string false "默认客户状态（传值使用 owned/pool，对应私有/公海，默认 owned）"
// @Param       maxErrors formData int false "最多返回多少条错误明细，默认200"
// @Success     200 {object} APIResponse{data=model.CustomerImportResult}
// @Router      /api/v1/customers/import/csv [post]
func (h *CustomerHandler) ImportCSV(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	if h.importService == nil {
		Error(c, http.StatusNotImplemented, 10045, "客户导入服务未配置")
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 10040, "请上传导入文件")
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
			Error(c, http.StatusBadRequest, 10041, "CSV 文件无效")
		case errors.Is(err, service.ErrCustomerImportInvalidHeader):
			Error(c, http.StatusBadRequest, 10042, "CSV 表头无效，必须包含 name、phone")
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 10043, "导入客户失败", err)
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
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	customer, err := h.service.ClaimCustomer(c.Request.Context(), id, userID)
	if err != nil {
		var freezeErr *service.CustomerClaimFreezeError
		if errors.As(err, &freezeErr) {
			Error(c, http.StatusForbidden, 10044, buildClaimFreezeMessage(freezeErr))
			return
		}
		if errors.Is(err, service.ErrCustomerNotFound) {
			Error(c, http.StatusNotFound, 10003, "客户不存在")
			return
		}
		if errors.Is(err, service.ErrCustomerNotInPool) {
			Error(c, http.StatusConflict, 10004, "该客户不在公海中")
			return
		}
		if errors.Is(err, service.ErrCustomerLimitExceeded) {
			Error(c, http.StatusConflict, 10038, "个人客户池已达上限，已成交客户不计入")
			return
		}
		if errors.Is(err, service.ErrCustomerSameDepartmentClaimForbidden) {
			Error(c, http.StatusForbidden, 10044, "该客户历史上已被你所在销售团队放弃，禁止再次领取，冷冻期结束后可再领取")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 10005, "领取客户失败", err)
		return
	}
	Success(c, customer)
}

// Convert godoc
// @Summary     转化客户
// @Description 将待转化客户按销售分配规则转化给负责人
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "客户ID"
// @Success     200 {object} APIResponse{data=model.Customer}
// @Router      /api/v1/customers/{id}/convert [post]
func (h *CustomerHandler) Convert(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	customer, err := h.service.ConvertCustomer(c.Request.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCustomerNotFound):
			Error(c, http.StatusNotFound, 10003, "客户不存在")
		case errors.Is(err, service.ErrCustomerNotInPool):
			Error(c, http.StatusConflict, 10006, "客户不在公海中，无法转化")
		case errors.Is(err, service.ErrCustomerConvertForbidden):
			Error(c, http.StatusForbidden, 10052, "当前客户不允许转化")
		case errors.Is(err, service.ErrCustomerNoOutsideSalesAvailable):
			Error(c, http.StatusConflict, 10039, "当前团队下暂无可分配的销售负责人")
		case errors.Is(err, service.ErrCustomerLimitExceeded):
			Error(c, http.StatusConflict, 10038, "个人客户池已达上限，已成交客户不计入")
		default:
			ErrorWithDetail(c, http.StatusInternalServerError, 10053, "转化客户失败", err)
		}
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
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	customer, err := h.service.ReleaseCustomer(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			Error(c, http.StatusNotFound, 10003, "客户不存在")
			return
		}
		if errors.Is(err, service.ErrCustomerAlreadyInPool) {
			Error(c, http.StatusConflict, 10006, "客户已在公海中")
			return
		}
		if errors.Is(err, service.ErrCustomerNotOwned) {
			Error(c, http.StatusForbidden, 10007, "当前用户不是该客户负责人")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 10008, "释放客户失败", err)
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
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	var req TransferCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10009, "请求参数错误", err)
		return
	}

	customer, err := h.service.TransferCustomer(c.Request.Context(), model.CustomerTransferInput{
		CustomerID:     id,
		ToOwnerUserID:  req.ToOwnerUserID,
		OperatorUserID: userID,
		Note:           strings.TrimSpace(req.Note),
	})
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			Error(c, http.StatusNotFound, 10003, "客户不存在")
			return
		}
		if errors.Is(err, service.ErrCustomerNotOwned) {
			Error(c, http.StatusForbidden, 10007, "当前用户不是该客户负责人")
			return
		}
		if errors.Is(err, service.ErrCustomerLimitExceeded) {
			Error(c, http.StatusConflict, 10038, "个人客户池已达上限，已成交客户不计入")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 10010, "转移客户失败", err)
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
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	var req AddPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10020, "请求参数错误", err)
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
			Error(c, http.StatusBadRequest, 10021, "联系电话格式不正确，请输入手机号或座机号")
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			Error(c, http.StatusConflict, 10022, "联系电话已存在")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 10023, "新增客户电话失败", err)
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
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	phones, err := h.service.ListPhones(c.Request.Context(), customerID)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10024, "加载客户电话失败", err)
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
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}
	phoneID, err := strconv.ParseInt(c.Param("phoneId"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10025, "无效的电话ID")
		return
	}

	var req UpdatePhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10020, "请求参数错误", err)
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
			Error(c, http.StatusNotFound, 10026, "电话不存在")
			return
		}
		if errors.Is(err, service.ErrInvalidPhoneFormat) {
			Error(c, http.StatusBadRequest, 10021, "联系电话格式不正确，请输入手机号或座机号")
			return
		}
		if errors.Is(err, service.ErrPhoneAlreadyExists) {
			Error(c, http.StatusConflict, 10022, "联系电话已存在")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 10027, "更新客户电话失败", err)
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
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}
	phoneID, err := strconv.ParseInt(c.Param("phoneId"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10025, "无效的电话ID")
		return
	}

	if err := h.service.DeletePhone(c.Request.Context(), customerID, phoneID); err != nil {
		if errors.Is(err, service.ErrPhoneNotFound) {
			Error(c, http.StatusNotFound, 10026, "电话不存在")
			return
		}
		ErrorWithDetail(c, http.StatusInternalServerError, 10028, "删除客户电话失败", err)
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
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, err := h.service.ListStatusLogs(c.Request.Context(), customerID, page, pageSize)
	if err != nil {
		ErrorWithDetail(c, http.StatusInternalServerError, 10029, "加载客户状态日志失败", err)
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
		Error(c, http.StatusUnauthorized, 30004, "登录信息无效")
		return
	}
	customerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 10002, "无效的客户ID")
		return
	}

	var req CreateStatusLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, 10020, "请求参数错误", err)
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
		ErrorWithDetail(c, http.StatusInternalServerError, 10030, "创建客户状态日志失败", err)
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

func maskCustomerData(customers []model.Customer, viewerID int64, hasViewer bool, actorRole string) []model.Customer {
	if len(customers) == 0 {
		return customers
	}
	if isMaskBypassRole(actorRole) {
		return customers
	}

	if !isSalesOrOperationRole(actorRole) || !hasViewer || viewerID <= 0 {
		return maskAllCustomers(customers)
	}

	masked := make([]model.Customer, len(customers))
	for i, customer := range customers {
		masked[i] = customer
		if shouldMaskCustomerForViewer(customer, viewerID) {
			applyCustomerMask(&masked[i])
		}
	}
	return masked
}

func maskAllCustomers(customers []model.Customer) []model.Customer {
	masked := make([]model.Customer, len(customers))
	for i, customer := range customers {
		masked[i] = customer
		applyCustomerMask(&masked[i])
	}
	return masked
}

func applyCustomerMask(customer *model.Customer) {
	// Mask company name and personal names
	customer.Name = util.MaskCompanyName(customer.Name)
	customer.LegalName = util.MaskPersonName(customer.LegalName)
	customer.ContactName = util.MaskPersonName(customer.ContactName)
	// Mask email and owner fields
	customer.Email = util.MaskEmail(customer.Email)
	customer.OwnerUserName = "*"
	customer.OwnerUserID = nil

	// Mask phone numbers
	if len(customer.Phones) > 0 {
		maskedPhones := make([]model.CustomerPhone, len(customer.Phones))
		for j, phone := range customer.Phones {
			maskedPhones[j] = phone
			maskedPhones[j].Phone = util.MaskPhone(phone.Phone)
		}
		customer.Phones = maskedPhones
	}
}

func shouldMaskCustomerForViewer(customer model.Customer, viewerID int64) bool {
	if customer.IsInPool || customer.Status == "pool" || customer.Status == "公海" {
		return false
	}
	if customer.OwnerUserID == nil {
		return false
	}
	return *customer.OwnerUserID != viewerID
}

func isMaskBypassRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}

func canViewCustomerAssignments(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin", "管理员", "finance", "finance_manager", "财务", "财务经理":
		return true
	default:
		return false
	}
}

func isSalesOrOperationRole(role string) bool {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "sales_director", "sales_manager", "sales_staff", "sales_inside", "sales_outside", "销售总监", "销售经理", "销售员工", "销售", "inside销售", "outside销售":
		return true
	case "ops_manager", "operation_manager", "ops_staff", "operation_staff", "运营经理", "运营员工", "运营":
		return true
	default:
		return false
	}
}

func buildClaimFreezeMessage(freezeErr *service.CustomerClaimFreezeError) string {
	if freezeErr == nil {
		return "当前客户对你处于回捡冷冻期，暂不可领取"
	}

	if freezeErr.BlockType == "department" {
		return "该客户历史上已被你所在销售团队放弃，当前仍处于团队禁领冷冻期。冷冻时长为" +
			strconv.Itoa(freezeErr.FreezeDays) +
			"天，剩余" + formatClaimFreezeRemain(freezeErr.Remaining) +
			"，冷冻结束时间为" + freezeErr.FrozenUntil.Local().Format("2006-01-02 15:04:05") +
			"，冷冻期结束后可再次领取。"
	}

	return "当前客户刚从你名下进入公海，现处于回捡冷冻期。冷冻时长为" +
		strconv.Itoa(freezeErr.FreezeDays) +
		"天，剩余" + formatClaimFreezeRemain(freezeErr.Remaining) +
		"，冷冻结束时间为" + freezeErr.FrozenUntil.Local().Format("2006-01-02 15:04:05") +
		"，冷冻期内暂不可领取。"
}

func formatClaimFreezeRemain(remaining time.Duration) string {
	if remaining <= 0 {
		return "0小时"
	}

	totalMinutes := int((remaining + time.Minute - 1) / time.Minute)
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60

	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, strconv.Itoa(days)+"天")
	}
	if hours > 0 {
		parts = append(parts, strconv.Itoa(hours)+"小时")
	}
	if minutes > 0 && days == 0 {
		parts = append(parts, strconv.Itoa(minutes)+"分钟")
	}
	if len(parts) == 0 {
		return "不足1分钟"
	}
	return strings.Join(parts, "")
}
