package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContractHandler struct {
	service service.ContractService
}

type ContractCreateRequest struct {
	ContractImage        string  `json:"contractImage"`
	PaymentImage         string  `json:"paymentImage"`
	PaymentStatus        string  `json:"paymentStatus"`
	Remark               string  `json:"remark"`
	CustomerID           int64   `json:"customerId" binding:"required"`
	CooperationType      string  `json:"cooperationType"`
	ContractNumber       string  `json:"contractNumber" binding:"required"`
	ContractNumberSuffix string  `json:"contractNumberSuffix"`
	ContractName         string  `json:"contractName" binding:"required"`
	ContractAmount       float64 `json:"contractAmount"`
	PaymentAmount        float64 `json:"paymentAmount"`
	CooperationYears     int     `json:"cooperationYears"`
	NodeCount            int     `json:"nodeCount"`
	ServiceUserID        *int64  `json:"serviceUserId"`
	WebsiteName          string  `json:"websiteName"`
	WebsiteURL           string  `json:"websiteUrl"`
	WebsiteUsername      string  `json:"websiteUsername"`
	IsOnline             bool    `json:"isOnline"`
	StartDate            *int64  `json:"startDate"`
	EndDate              *int64  `json:"endDate"`
	AuditStatus          string  `json:"auditStatus"`
	ExpiryHandlingStatus string  `json:"expiryHandlingStatus"`
}

type ContractUpdateRequest struct {
	ContractImage        string  `json:"contractImage"`
	PaymentImage         string  `json:"paymentImage"`
	PaymentStatus        string  `json:"paymentStatus"`
	Remark               string  `json:"remark"`
	CustomerID           int64   `json:"customerId" binding:"required"`
	CooperationType      string  `json:"cooperationType"`
	ContractNumber       string  `json:"contractNumber" binding:"required"`
	ContractNumberSuffix string  `json:"contractNumberSuffix"`
	ContractName         string  `json:"contractName" binding:"required"`
	ContractAmount       float64 `json:"contractAmount"`
	PaymentAmount        float64 `json:"paymentAmount"`
	CooperationYears     int     `json:"cooperationYears"`
	NodeCount            int     `json:"nodeCount"`
	ServiceUserID        *int64  `json:"serviceUserId"`
	WebsiteName          string  `json:"websiteName"`
	WebsiteURL           string  `json:"websiteUrl"`
	WebsiteUsername      string  `json:"websiteUsername"`
	IsOnline             bool    `json:"isOnline"`
	StartDate            *int64  `json:"startDate"`
	EndDate              *int64  `json:"endDate"`
	AuditStatus          string  `json:"auditStatus"`
	ExpiryHandlingStatus string  `json:"expiryHandlingStatus"`
}

type ContractAuditRequest struct {
	ContractImage        string  `json:"contractImage"`
	PaymentImage         string  `json:"paymentImage"`
	PaymentStatus        string  `json:"paymentStatus"`
	Remark               string  `json:"remark"`
	CustomerID           int64   `json:"customerId" binding:"required"`
	CooperationType      string  `json:"cooperationType"`
	ContractNumber       string  `json:"contractNumber" binding:"required"`
	ContractNumberSuffix string  `json:"contractNumberSuffix"`
	ContractName         string  `json:"contractName" binding:"required"`
	ContractAmount       float64 `json:"contractAmount"`
	PaymentAmount        float64 `json:"paymentAmount"`
	CooperationYears     int     `json:"cooperationYears"`
	NodeCount            int     `json:"nodeCount"`
	ServiceUserID        *int64  `json:"serviceUserId"`
	WebsiteName          string  `json:"websiteName"`
	WebsiteURL           string  `json:"websiteUrl"`
	WebsiteUsername      string  `json:"websiteUsername"`
	IsOnline             bool    `json:"isOnline"`
	StartDate            *int64  `json:"startDate"`
	EndDate              *int64  `json:"endDate"`
	AuditStatus          string  `json:"auditStatus" binding:"required"`
	AuditComment         string  `json:"auditComment"`
	ExpiryHandlingStatus string  `json:"expiryHandlingStatus"`
}

func NewContractHandler(service service.ContractService) *ContractHandler {
	return &ContractHandler{service: service}
}

// CheckNumber godoc
// @Summary     校验合同编号是否可用
// @Tags        contracts
// @Produce     json
// @Security    BearerAuth
// @Param       contractNumber query string true "合同编号"
// @Param       excludeId query int false "排除的合同ID"
// @Success     200 {object} APIResponse
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Router      /api/v1/contracts/check-number [get]
func (h *ContractHandler) CheckNumber(c *gin.Context) {
	contractNumber := strings.TrimSpace(c.Query("contractNumber"))
	excludeID, _ := strconv.ParseInt(strings.TrimSpace(c.Query("excludeId")), 10, 64)

	available, err := h.service.IsContractNumberAvailable(c.Request.Context(), contractNumber, excludeID)
	if err != nil {
		if errors.Is(err, service.ErrContractContractNumberRequired) {
			Error(c, http.StatusBadRequest, 60007, "contract number is required")
			return
		}
		Error(c, http.StatusInternalServerError, 60001, "failed to check contract number")
		return
	}

	Success(c, gin.H{
		"available": available,
	})
}

// List godoc
// @Summary     获取合同列表
// @Tags        contracts
// @Produce     json
// @Security    BearerAuth
// @Param       keyword query string false "关键词"
// @Param       paymentStatus query string false "付款状态"
// @Param       cooperationType query string false "合作类型"
// @Param       auditStatus query string false "审核状态"
// @Param       expiryHandlingStatus query string false "到期处理状态"
// @Param       userId query int false "签单人ID"
// @Param       customerId query int false "客户ID"
// @Param       page query int false "页码"
// @Param       pageSize query int false "每页条数"
// @Success     200 {object} APIResponse{data=model.ContractListResult}
// @Failure     401 {object} APIResponse
// @Router      /api/v1/contracts [get]
func (h *ContractHandler) List(c *gin.Context) {
	actorUserID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	userID, _ := strconv.ParseInt(strings.TrimSpace(c.Query("userId")), 10, 64)
	customerID, _ := strconv.ParseInt(strings.TrimSpace(c.Query("customerId")), 10, 64)

	result, err := h.service.ListContracts(c.Request.Context(), model.ContractListFilter{
		Keyword:              strings.TrimSpace(c.Query("keyword")),
		PaymentStatus:        strings.TrimSpace(c.Query("paymentStatus")),
		CooperationType:      strings.TrimSpace(c.Query("cooperationType")),
		AuditStatus:          strings.TrimSpace(c.Query("auditStatus")),
		ExpiryHandlingStatus: strings.TrimSpace(c.Query("expiryHandlingStatus")),
		UserID:               userID,
		CustomerID:           customerID,
		ActorUserID:          actorUserID,
		ActorRole:            currentUserRole(c),
		Page:                 page,
		PageSize:             pageSize,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, 60001, "failed to list contracts")
		return
	}
	Success(c, result)
}

// GetByID godoc
// @Summary     获取合同详情
// @Tags        contracts
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "合同ID"
// @Success     200 {object} APIResponse{data=model.Contract}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/contracts/{id} [get]
func (h *ContractHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 60002, "invalid contract id")
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	contract, err := h.service.GetContractByID(c.Request.Context(), id, userID, currentUserRole(c))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, contract)
}

// Create godoc
// @Summary     创建合同
// @Tags        contracts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body ContractCreateRequest true "合同信息"
// @Success     200 {object} APIResponse{data=model.Contract}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     409 {object} APIResponse
// @Router      /api/v1/contracts [post]
func (h *ContractHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	var req ContractCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 60005, "invalid request body")
		return
	}
	contractNumber := strings.TrimSpace(req.ContractNumber)
	if strings.TrimSpace(req.ContractNumberSuffix) != "" {
		contractNumber = strings.TrimSpace(req.ContractNumberSuffix)
	}
	contract, err := h.service.CreateContract(c.Request.Context(), model.ContractCreateInput{
		ContractImage:        req.ContractImage,
		PaymentImage:         req.PaymentImage,
		PaymentStatus:        req.PaymentStatus,
		Remark:               req.Remark,
		UserID:               userID,
		CustomerID:           req.CustomerID,
		CooperationType:      req.CooperationType,
		ContractNumber:       contractNumber,
		ContractName:         req.ContractName,
		ContractAmount:       req.ContractAmount,
		PaymentAmount:        req.PaymentAmount,
		CooperationYears:     req.CooperationYears,
		NodeCount:            req.NodeCount,
		ServiceUserID:        req.ServiceUserID,
		WebsiteName:          req.WebsiteName,
		WebsiteURL:           req.WebsiteURL,
		WebsiteUsername:      req.WebsiteUsername,
		IsOnline:             req.IsOnline,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		AuditStatus:          req.AuditStatus,
		ExpiryHandlingStatus: req.ExpiryHandlingStatus,
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, contract)
}

// Update godoc
// @Summary     更新合同
// @Tags        contracts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "合同ID"
// @Param       body body ContractUpdateRequest true "合同信息"
// @Success     200 {object} APIResponse{data=model.Contract}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/contracts/{id} [put]
func (h *ContractHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 60002, "invalid contract id")
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	var req ContractUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 60005, "invalid request body")
		return
	}
	contractNumber := strings.TrimSpace(req.ContractNumber)
	if strings.TrimSpace(req.ContractNumberSuffix) != "" {
		contractNumber = strings.TrimSpace(req.ContractNumberSuffix)
	}
	contract, err := h.service.UpdateContract(c.Request.Context(), id, model.ContractUpdateInput{
		ContractImage:        req.ContractImage,
		PaymentImage:         req.PaymentImage,
		PaymentStatus:        req.PaymentStatus,
		Remark:               req.Remark,
		UserID:               userID,
		CustomerID:           req.CustomerID,
		CooperationType:      req.CooperationType,
		ContractNumber:       contractNumber,
		ContractName:         req.ContractName,
		ContractAmount:       req.ContractAmount,
		PaymentAmount:        req.PaymentAmount,
		CooperationYears:     req.CooperationYears,
		NodeCount:            req.NodeCount,
		ServiceUserID:        req.ServiceUserID,
		WebsiteName:          req.WebsiteName,
		WebsiteURL:           req.WebsiteURL,
		WebsiteUsername:      req.WebsiteUsername,
		IsOnline:             req.IsOnline,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		AuditStatus:          req.AuditStatus,
		ExpiryHandlingStatus: req.ExpiryHandlingStatus,
	}, userID, currentUserRole(c))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, contract)
}

// Audit godoc
// @Summary     审核合同
// @Tags        contracts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "合同ID"
// @Param       body body ContractAuditRequest true "审核信息"
// @Success     200 {object} APIResponse{data=model.Contract}
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/contracts/{id}/audit [post]
func (h *ContractHandler) Audit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 60002, "invalid contract id")
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	var req ContractAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 60005, "invalid request body")
		return
	}
	contractNumber := strings.TrimSpace(req.ContractNumber)
	if strings.TrimSpace(req.ContractNumberSuffix) != "" {
		contractNumber = strings.TrimSpace(req.ContractNumberSuffix)
	}
	contract, err := h.service.AuditContract(c.Request.Context(), id, model.ContractUpdateInput{
		ContractImage:        req.ContractImage,
		PaymentImage:         req.PaymentImage,
		PaymentStatus:        req.PaymentStatus,
		Remark:               req.Remark,
		UserID:               userID,
		CustomerID:           req.CustomerID,
		CooperationType:      req.CooperationType,
		ContractNumber:       contractNumber,
		ContractName:         req.ContractName,
		ContractAmount:       req.ContractAmount,
		PaymentAmount:        req.PaymentAmount,
		CooperationYears:     req.CooperationYears,
		NodeCount:            req.NodeCount,
		ServiceUserID:        req.ServiceUserID,
		WebsiteName:          req.WebsiteName,
		WebsiteURL:           req.WebsiteURL,
		WebsiteUsername:      req.WebsiteUsername,
		IsOnline:             req.IsOnline,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		AuditStatus:          req.AuditStatus,
		AuditComment:         req.AuditComment,
		ExpiryHandlingStatus: req.ExpiryHandlingStatus,
	}, userID, currentUserRole(c))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, contract)
}

// Delete godoc
// @Summary     删除合同
// @Tags        contracts
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "合同ID"
// @Success     200 {object} APIResponse
// @Failure     400 {object} APIResponse
// @Failure     401 {object} APIResponse
// @Failure     404 {object} APIResponse
// @Router      /api/v1/contracts/{id} [delete]
func (h *ContractHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		Error(c, http.StatusBadRequest, 60002, "invalid contract id")
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		Error(c, http.StatusUnauthorized, 30004, "invalid token claims")
		return
	}
	if err := h.service.DeleteContract(c.Request.Context(), id, userID, currentUserRole(c)); err != nil {
		h.handleServiceError(c, err)
		return
	}
	Success(c, nil)
}

func (h *ContractHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrContractNotFound):
		Error(c, http.StatusNotFound, 60003, "contract not found")
	case errors.Is(err, service.ErrContractForbidden):
		Error(c, http.StatusForbidden, 60004, "contract access forbidden")
	case errors.Is(err, service.ErrContractContractNumberRequired):
		Error(c, http.StatusBadRequest, 60007, "contract number is required")
	case errors.Is(err, service.ErrContractNameRequired):
		Error(c, http.StatusBadRequest, 60008, "contract name is required")
	case errors.Is(err, service.ErrContractInvalidCooperationType):
		Error(c, http.StatusBadRequest, 60009, "invalid cooperation type")
	case errors.Is(err, service.ErrContractInvalidPaymentStatus):
		Error(c, http.StatusBadRequest, 60010, "invalid payment status")
	case errors.Is(err, service.ErrContractInvalidAuditStatus):
		Error(c, http.StatusBadRequest, 60011, "invalid audit status")
	case errors.Is(err, service.ErrContractInvalidExpiryHandlingStatus):
		Error(c, http.StatusBadRequest, 60012, "invalid expiry handling status")
	case errors.Is(err, service.ErrContractInvalidAmount):
		Error(c, http.StatusBadRequest, 60013, "invalid contract or payment amount")
	case errors.Is(err, service.ErrContractPaymentExceedsContract):
		Error(c, http.StatusBadRequest, 60014, "payment amount cannot exceed contract amount")
	case errors.Is(err, service.ErrContractInvalidDateRange):
		Error(c, http.StatusBadRequest, 60015, "end date cannot be earlier than start date")
	case errors.Is(err, service.ErrContractNumberExists):
		Error(c, http.StatusConflict, 60016, "contract number already exists")
	case errors.Is(err, service.ErrContractInvalidUser):
		Error(c, http.StatusBadRequest, 60017, "invalid user")
	case errors.Is(err, service.ErrContractInvalidCustomer):
		Error(c, http.StatusBadRequest, 60018, "invalid customer")
	case errors.Is(err, service.ErrContractInvalidServiceUser):
		Error(c, http.StatusBadRequest, 60019, "invalid service user")
	case errors.Is(err, service.ErrContractAuditStatusForbidden):
		Error(c, http.StatusForbidden, 60020, "only admin or finance manager can update audit status")
	case errors.Is(err, service.ErrContractNumberForbidden):
		Error(c, http.StatusForbidden, 60021, "only admin can update contract number")
	case errors.Is(err, service.ErrContractAuditedReadonly):
		Error(c, http.StatusForbidden, 60022, "audited contract cannot be modified")
	default:
		Error(c, http.StatusInternalServerError, 60099, "contract operation failed")
	}
}

func currentUserRole(c *gin.Context) string {
	value, exists := c.Get("role")
	if !exists {
		return ""
	}
	role, _ := value.(string)
	return role
}
