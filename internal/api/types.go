package api

import "encoding/json"

type APIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type SearchItem struct {
	Name     string   `json:"name"`
	FormType string   `json:"formType"`
	Type     int      `json:"type"`
	Values   []string `json:"values"`
}

type SearchBO struct {
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	PageType   int          `json:"pageType"`
	Search     string       `json:"search,omitempty"`
	PoolID     int64        `json:"poolId,omitempty"`
	SceneID    int64        `json:"sceneId,omitempty"`
	Label      int          `json:"label,omitempty"`
	SubLabel   int          `json:"subLabel,omitempty"`
	SortField  string       `json:"sortField,omitempty"`
	Order      int          `json:"order,omitempty"`
	SearchList []SearchItem `json:"searchList,omitempty"`
}
