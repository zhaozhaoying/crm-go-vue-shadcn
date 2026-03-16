package errmsg

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	validatorPattern          = regexp.MustCompile(`(?i)(?:key:\s*'[^']+'\s*error:)?field validation for '([^']+)' failed on the '([^']+)' tag`)
	jsonUnmarshalFieldPattern = regexp.MustCompile(`(?i)json: cannot unmarshal ([^ ]+) into go struct field [^.]+\.(\w+) of type (.+)`)
	duplicateEntryPattern     = regexp.MustCompile(`(?i)duplicate entry '([^']+)' for key '([^']+)'`)
	unsupportedFieldPattern   = regexp.MustCompile(`(?i)unsupported field:\s*(.+)`)
	requiredColumnsPattern    = regexp.MustCompile(`(?i)required columns:\s*(.+)`)
	unsupportedDatetimeType   = regexp.MustCompile(`(?i)unsupported datetime type:\s*(.+)`)
	invalidDatetimeValue      = regexp.MustCompile(`(?i)invalid datetime value:\s*(.+)`)
	invalidSubClaimType       = regexp.MustCompile(`(?i)invalid sub claim type:\s*(.+)`)
	providerNotConfigured     = regexp.MustCompile(`(?i)provider for platform\s+(\d+)\s+is not configured`)
	baiduStatusPattern        = regexp.MustCompile(`(?i)baidu status=(\d+)\s+message=(.+)`)
	chinesePattern            = regexp.MustCompile(`[\p{Han}]`)
)

var exactTranslations = map[string]string{
	"upload service not configured":                         "上传服务未配置",
	"invalid image type":                                    "图片格式不正确",
	"image too large":                                       "图片大小超出限制",
	"image upload failed":                                   "图片上传失败",
	"role not found":                                        "角色不存在",
	"role already exists":                                   "角色已存在",
	"user not found":                                        "用户不存在",
	"invalid role":                                          "角色无效",
	"invalid password":                                      "密码无效",
	"invalid user ids":                                      "用户ID无效",
	"resource pool item not found":                          "资源池线索不存在",
	"resource pool item already converted":                  "资源池线索已转为客户",
	"role name already exists":                              "角色名称已存在",
	"captcha invalid":                                       "验证码错误",
	"captcha expired":                                       "验证码已过期",
	"captcha too many attempts":                             "验证码错误次数过多",
	"invalid customer import file":                          "客户导入文件无效",
	"invalid customer import header":                        "客户导入表头无效",
	"username already exists":                               "用户名已存在",
	"invalid username or password":                          "用户名或密码错误",
	"user is disabled":                                      "用户已被禁用",
	"invalid refresh token":                                 "刷新令牌无效",
	"token revoked":                                         "令牌已失效",
	"external company search keyword is required":           "抓取关键词不能为空",
	"external company search platform is required":          "抓取平台不能为空",
	"external company search platform is unsupported":       "抓取平台不支持",
	"external company search task forbidden":                "无权访问该抓取任务",
	"external company search task not found":                "抓取任务不存在",
	"external company search target reached":                "已达到抓取目标数量",
	"customer not found":                                    "客户不存在",
	"customer not in pool":                                  "客户不在公海中",
	"customer already in pool":                              "客户已在公海中",
	"customer not owned":                                    "当前用户不是该客户负责人",
	"customer name already exists":                          "客户名称已存在",
	"customer legal name already exists":                    "法人名称已存在",
	"customer weixin already exists":                        "微信号已存在",
	"customer phone already exists":                         "客户手机号已存在",
	"customer name is required":                             "客户名称不能为空",
	"customer limit exceeded":                               "个人客户池已达上限",
	"same department customer cannot be claimed":            "同部门客户不可领取",
	"no outside sales available":                            "当前团队下暂无可分配的销售负责人",
	"no assignable sales owner available":                   "当前团队下暂无可分配的销售负责人",
	"phone not found":                                       "电话不存在",
	"phone already exists for this customer":                "手机号已存在",
	"invalid phone format":                                  "手机号格式不正确",
	"resource pool invalid input":                           "资源池请求参数无效",
	"resource pool provider not configured":                 "资源池地图服务未配置",
	"resource pool location not found":                      "未找到查询位置",
	"resource pool search failed":                           "资源池检索失败",
	"resource pool no convertible phone":                    "地图资源电话不可用于创建客户",
	"resource pool convert failed":                          "地图资源转客户失败",
	"contract not found":                                    "合同不存在",
	"contract number already exists":                        "合同编号已存在",
	"invalid user":                                          "用户无效",
	"invalid customer":                                      "客户无效",
	"invalid service user":                                  "客服人员无效",
	"contract access forbidden":                             "无权访问该合同",
	"contract number is required":                           "合同编号不能为空",
	"contract name is required":                             "合同名称不能为空",
	"invalid cooperation type":                              "合作类型无效",
	"invalid payment status":                                "付款状态无效",
	"invalid audit status":                                  "审核状态无效",
	"invalid expiry handling status":                        "到期处理状态无效",
	"invalid amount":                                        "金额无效",
	"payment amount exceeds contract amount":                "回款金额不能大于合同金额",
	"end date cannot be earlier than start date":            "结束日期不能早于开始日期",
	"only admin or finance manager can update audit status": "仅管理员或财务经理可以修改审核状态",
	"only admin can update contract number":                 "仅管理员可以修改合同编号",
	"audited contract is readonly":                          "已审核合同不允许修改",
	"company_no is required":                                "企业编号不能为空",
	"record not found":                                      "记录不存在",
	"sql: no rows in result set":                            "记录不存在",
	"empty datetime string":                                 "时间字符串为空",
	"negative value":                                        "数值不能为负数",
	"duplicate name in database":                            "数据库中已存在同名客户",
	"duplicate legalname in database":                       "数据库中已存在相同法人名称",
	"duplicate weixin in database":                          "数据库中已存在相同微信号",
	"duplicate phone in database":                           "数据库中已存在相同手机号",
	"duplicate record in database":                          "数据库中已存在重复客户记录",
	"duplicate name in import file":                         "导入文件中客户名称重复",
	"duplicate legalname in import file":                    "导入文件中法人名称重复",
	"duplicate weixin in import file":                       "导入文件中微信号重复",
	"duplicate phone in import file":                        "导入文件中手机号重复",
	"name is required":                                      "客户名称不能为空",
	"phone is invalid":                                      "手机号格式不正确",
	"province is invalid":                                   "省编码无效",
	"city is invalid":                                       "市编码无效",
	"area is invalid":                                       "区编码无效",
	"customerlevelid is invalid":                            "客户级别ID无效",
	"customersourceid is invalid":                           "客户来源ID无效",
	"owneruserid is invalid":                                "负责人ID无效",
	"operator user id is required":                          "操作人ID不能为空",
	"task created":                                          "任务已创建",
	"task canceled":                                         "任务已取消",
	"task started":                                          "任务开始执行",
	"task progress updated":                                 "任务进度已更新",
	"result saved":                                          "结果已保存",
	"task completed":                                        "任务已完成",
}

var fieldTranslations = map[string]string{
	"username":           "用户名",
	"password":           "密码",
	"nickname":           "昵称",
	"email":              "邮箱",
	"mobile":             "手机号",
	"avatar":             "头像",
	"roleid":             "角色ID",
	"role_id":            "角色ID",
	"parentid":           "上级用户ID",
	"parent_id":          "上级用户ID",
	"userids":            "用户ID列表",
	"user_ids":           "用户ID列表",
	"name":               "名称",
	"label":              "标签",
	"sort":               "排序",
	"customerid":         "客户ID",
	"customer_id":        "客户ID",
	"phone":              "手机号",
	"phonelabel":         "电话标签",
	"phone_label":        "电话标签",
	"contactname":        "联系人",
	"contact_name":       "联系人",
	"legalname":          "法人名称",
	"legal_name":         "法人名称",
	"weixin":             "微信号",
	"status":             "状态",
	"owneruserid":        "负责人ID",
	"owner_user_id":      "负责人ID",
	"toowneruserid":      "转移负责人ID",
	"to_owner_user_id":   "转移负责人ID",
	"content":            "跟进内容",
	"followmethodid":     "跟进方式ID",
	"follow_method_id":   "跟进方式ID",
	"contractnumber":     "合同编号",
	"contract_number":    "合同编号",
	"contractname":       "合同名称",
	"contract_name":      "合同名称",
	"auditstatus":        "审核状态",
	"audit_status":       "审核状态",
	"refreshtoken":       "刷新令牌",
	"refresh_token":      "刷新令牌",
	"captchaid":          "验证码ID",
	"captcha_id":         "验证码ID",
	"captchacode":        "验证码",
	"captcha_code":       "验证码",
	"keyword":            "关键词",
	"platforms":          "抓取平台",
	"page":               "页码",
	"pagesize":           "每页条数",
	"page_size":          "每页条数",
	"batchsize":          "批大小",
	"batch_size":         "批大小",
	"defaultstatus":      "默认客户状态",
	"default_status":     "默认客户状态",
	"maxerrors":          "最大错误数",
	"max_errors":         "最大错误数",
	"excludeid":          "排除ID",
	"exclude_id":         "排除ID",
	"id":                 "ID",
	"serviceuserid":      "客服人员ID",
	"service_user_id":    "客服人员ID",
	"customerlevelid":    "客户级别ID",
	"customer_level_id":  "客户级别ID",
	"customersourceid":   "客户来源ID",
	"customer_source_id": "客户来源ID",
	"nextfollowtime":     "下次跟进时间",
	"next_follow_time":   "下次跟进时间",
	"appointmenttime":    "约见时间",
	"appointment_time":   "约见时间",
	"shootingtime":       "拍摄时间",
	"shooting_time":      "拍摄时间",
	"ownershipscope":     "查看范围",
	"ownership_scope":    "查看范围",
	"searchoptions":      "搜索配置",
	"search_options":     "搜索配置",
}

func Normalize(detail string) string {
	detail = strings.Join(strings.Fields(strings.TrimSpace(detail)), " ")
	if detail == "" {
		return ""
	}

	lower := strings.ToLower(detail)
	if translated, ok := exactTranslations[lower]; ok {
		return translated
	}

	if translated := normalizeValidatorError(detail); translated != "" {
		return translated
	}
	if translated := normalizeJSONError(detail); translated != "" {
		return translated
	}
	if translated := normalizeByPrefix(detail, lower); translated != "" {
		return translated
	}
	if translated := normalizeByPattern(detail); translated != "" {
		return translated
	}
	if translated := normalizeDatabaseError(detail, lower); translated != "" {
		return translated
	}
	if chinesePattern.MatchString(detail) {
		return detail
	}
	if len(detail) > 300 {
		return detail[:300] + "..."
	}
	return detail
}

func normalizeByPrefix(detail, lower string) string {
	switch {
	case strings.HasPrefix(lower, "csv parse error:"):
		return "CSV 解析失败：" + Normalize(strings.TrimSpace(detail[len("csv parse error:"):]))
	case strings.HasPrefix(lower, "insert customer failed:"):
		return "新增客户失败：" + Normalize(strings.TrimSpace(detail[len("insert customer failed:"):]))
	case strings.HasPrefix(lower, "insert customer phone failed:"):
		return "新增客户电话失败：" + Normalize(strings.TrimSpace(detail[len("insert customer phone failed:"):]))
	case strings.HasPrefix(lower, "insert owner log failed:"):
		return "新增客户归属日志失败：" + Normalize(strings.TrimSpace(detail[len("insert owner log failed:"):]))
	case strings.HasPrefix(lower, "invalid customer import header:"):
		return "客户导入表头无效：" + Normalize(strings.TrimSpace(detail[len("invalid customer import header:"):]))
	}
	return ""
}

func normalizeByPattern(detail string) string {
	if matches := unsupportedFieldPattern.FindStringSubmatch(detail); len(matches) == 2 {
		return fmt.Sprintf("不支持的字段：%s", strings.TrimSpace(matches[1]))
	}
	if matches := requiredColumnsPattern.FindStringSubmatch(detail); len(matches) == 2 {
		return fmt.Sprintf("必需列：%s", normalizeColumnList(matches[1]))
	}
	if matches := unsupportedDatetimeType.FindStringSubmatch(detail); len(matches) == 2 {
		return fmt.Sprintf("不支持的时间类型：%s", strings.TrimSpace(matches[1]))
	}
	if matches := invalidDatetimeValue.FindStringSubmatch(detail); len(matches) == 2 {
		return fmt.Sprintf("无效的时间值：%s", strings.TrimSpace(matches[1]))
	}
	if matches := invalidSubClaimType.FindStringSubmatch(detail); len(matches) == 2 {
		return fmt.Sprintf("sub 声明类型无效：%s", strings.TrimSpace(matches[1]))
	}
	if matches := providerNotConfigured.FindStringSubmatch(detail); len(matches) == 2 {
		return fmt.Sprintf("平台 %s 的抓取服务未配置", strings.TrimSpace(matches[1]))
	}
	if matches := baiduStatusPattern.FindStringSubmatch(detail); len(matches) == 3 {
		return fmt.Sprintf("百度地图返回异常（状态码：%s，消息：%s）", strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]))
	}
	return ""
}

func normalizeDatabaseError(detail, lower string) string {
	if matches := duplicateEntryPattern.FindStringSubmatch(detail); len(matches) == 3 {
		return fmt.Sprintf("数据重复，违反唯一约束（%s）", strings.TrimSpace(matches[2]))
	}
	if strings.Contains(lower, "unique constraint failed") {
		return "数据重复，违反唯一约束"
	}
	if strings.Contains(lower, "foreign key constraint fails") {
		return "关联数据不存在或仍被引用，无法完成当前操作"
	}
	if strings.Contains(lower, "duplicated key not allowed") {
		return "数据重复，违反唯一约束"
	}
	return ""
}

func normalizeValidatorError(detail string) string {
	matches := validatorPattern.FindStringSubmatch(detail)
	if len(matches) != 3 {
		return ""
	}

	field := translateFieldName(matches[1])
	tag := strings.ToLower(strings.TrimSpace(matches[2]))
	switch tag {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "min":
		return fmt.Sprintf("%s未达到最小限制", field)
	case "max":
		return fmt.Sprintf("%s超过最大限制", field)
	case "oneof":
		return fmt.Sprintf("%s取值不合法", field)
	case "email":
		return fmt.Sprintf("%s格式不正确", field)
	default:
		return fmt.Sprintf("%s校验失败（%s）", field, tag)
	}
}

func normalizeJSONError(detail string) string {
	switch detail {
	case "EOF":
		return "请求体不能为空"
	case "unexpected EOF", "unexpected end of JSON input":
		return "JSON 格式不完整"
	}

	lower := strings.ToLower(detail)
	if strings.HasPrefix(lower, "invalid character ") && strings.Contains(lower, "looking for beginning of object key string") {
		return "JSON 格式错误"
	}

	matches := jsonUnmarshalFieldPattern.FindStringSubmatch(detail)
	if len(matches) == 4 {
		field := translateFieldName(matches[2])
		expectedType := translateTypeName(matches[3])
		return fmt.Sprintf("%s类型错误，应为%s", field, expectedType)
	}
	return ""
}

func translateFieldName(field string) string {
	key := strings.ToLower(strings.TrimSpace(field))
	if translated, ok := fieldTranslations[key]; ok {
		return translated
	}
	return strings.TrimSpace(field)
}

func translateTypeName(typeName string) string {
	switch strings.ToLower(strings.TrimSpace(typeName)) {
	case "string":
		return "字符串"
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return "整数"
	case "float32", "float64":
		return "数字"
	case "bool":
		return "布尔值"
	case "[]string":
		return "字符串数组"
	case "[]int64", "[]int", "[]uint64":
		return "整数数组"
	default:
		return strings.TrimSpace(typeName)
	}
}

func normalizeColumnList(raw string) string {
	parts := strings.Split(raw, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return strings.Join(normalized, "、")
}
