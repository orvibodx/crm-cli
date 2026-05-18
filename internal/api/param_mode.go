package api

import "strings"

var requestParamPaths = map[string]bool{
	"crmActionRecord/queryRecordList": true,
	"crmCustomer/field":                true,
	"crmCustomer/lock":                 true,
	"crmCustomer/setDealStatus":        true,
	"crmCustomer/queryCustomerSetting": true,
	"crmCustomer/deleteCustomerSetting": true,
	"crmCustomer/nearbyCustomer":       true,
	"crmCustomer/queryCustomerName":    true,
	"crmCustomer/hasFollowTask":        true,
	"crmCustomer/num":                  true,
	"crmCustomer/queryFileList":        true,
	"crmActivity/deleteOutworkSign":    true,
	"crmActivity/setPictureSetting":    true,
	"crmActivity/queryOutworkStats":    true,
}

func IsRequestParam(path string) bool {
	return requestParamPaths[path]
}

func NormalizePath(path string) string {
	return strings.Trim(path, "/")
}
