# Orvibo CRM CLI MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI (`crm-cli`) that lets AI Agents operate the Orvibo CRM system via Claude Code Skills.

**Architecture:** Cobra-based CLI with a shared HTTP client layer that handles Admin-Token injection and RequestParam auto-detection. Entity commands are generic — a single `entity.go` dispatches to the correct API prefix via a registry. Default output is JSON for AI consumption.

**Tech Stack:** Go 1.22+, cobra, net/http, tablewriter

**Project directory:** `/Users/apple/orvibocli`

---

## File Map

| File | Responsibility |
|---|---|
| `main.go` | Entry point, wire cobra commands |
| `cmd/root.go` | Root command, persistent flags (--env, --format) |
| `cmd/auth.go` | `auth login`, `auth whoami` |
| `cmd/entity.go` | `customer list/detail`, `leads list/detail`, etc. |
| `cmd/activity.go` | `activity list --type --id` |
| `cmd/api.go` | `api <METHOD> <PATH> --data/--query` |
| `internal/client/config.go` | Read/write `~/.crm-cli/config.json` |
| `internal/client/http.go` | HTTP client: token injection, RequestParam auto-detect, response handling |
| `internal/api/types.go` | Shared types: `APIResponse`, `SearchBO`, `SearchItem`, `BasePage` |
| `internal/api/registry.go` | Entity registry: name → {apiPrefix, label, crmEnum} |
| `internal/api/param_mode.go` | Hardcoded set of RequestParam endpoint paths |
| `internal/filter/parser.go` | `--filter` DSL → `[]SearchItem` |
| `internal/output/json.go` | JSON output (default) |
| `internal/output/table.go` | Table output (`--format table`) |
| `skills/crm-shared/SKILL.md` | Shared skill: auth, CrmEnum, error codes |
| `skills/crm-customer/SKILL.md` | Customer skill: commands, tips |

---

## Task 1: Go Module + Cobra Scaffold

**Files:**
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `go.mod`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/apple/orvibocli
go mod init github.com/orvibo/crm-cli
```

- [ ] **Step 2: Install cobra dependency**

```bash
cd /Users/apple/orvibocli
go get github.com/spf13/cobra@latest
```

- [ ] **Step 3: Create `cmd/root.go`**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	env    string
	format string
)

var rootCmd = &cobra.Command{
	Use:   "crm-cli",
	Short: "Orvibo CRM command-line tool",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&env, "env", "prod", "Environment: prod, test, local")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "Output format: json, table")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create `main.go`**

```go
package main

import "github.com/orvibo/crm-cli/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 5: Verify build**

```bash
cd /Users/apple/orvibocli
go build -o crm-cli . && ./crm-cli --help
```

Expected: prints help text with `--env` and `--format` flags.

- [ ] **Step 6: Commit**

```bash
cd /Users/apple/orvibocli
git add main.go cmd/root.go go.mod go.sum
git commit -m "feat: cobra scaffold with root command and persistent flags"
```

---

## Task 2: Config Layer

**Files:**
- Create: `internal/client/config.go`

- [ ] **Step 1: Create config type and read/write functions**

```go
package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Token  string `json:"token"`
	Env    string `json:"env"`
	Format string `json:"format"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home dir: %w", err)
	}
	return filepath.Join(home, ".crm-cli", "config.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Env: "prod", Format: "json"}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

Expected: compiles with no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/client/config.go
git commit -m "feat: config read/write for ~/.crm-cli/config.json"
```

---

## Task 3: HTTP Client

**Files:**
- Create: `internal/api/types.go`
- Create: `internal/api/param_mode.go`
- Create: `internal/client/http.go`

- [ ] **Step 1: Create shared API types**

```go
package api

// APIResponse is the universal response wrapper.
type APIResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// SearchItem is one entry in CrmSearchBO.searchList.
type SearchItem struct {
	Name     string   `json:"name"`
	FormType string   `json:"formType"`
	Type     int      `json:"type"`
	Values   []string `json:"values"`
}

// SearchBO is CrmSearchBO — the universal list query body.
type SearchBO struct {
	Page     int          `json:"page"`
	Limit    int          `json:"limit"`
	PageType int          `json:"pageType"`
	Search   string       `json:"search,omitempty"`
	PoolID   int64        `json:"poolId,omitempty"`
	SceneID  int64        `json:"sceneId,omitempty"`
	Label    int          `json:"label,omitempty"`
	SubLabel int          `json:"subLabel,omitempty"`
	SortField string      `json:"sortField,omitempty"`
	Order    int          `json:"order,omitempty"`
	SearchList []SearchItem `json:"searchList,omitempty"`
}
```

Note: needs `import "encoding/json"` for `json.RawMessage`.

- [ ] **Step 2: Create RequestParam endpoint set**

```go
package api

import "strings"

// requestParamPaths contains all API paths that use @RequestParam
// instead of @RequestBody. When a path matches, the client sends
// parameters as URL query string instead of JSON body.
var requestParamPaths = map[string]bool{
	// crmActionRecord
	"crmActionRecord/queryRecordList": true,
	// crmCustomer
	"crmCustomer/field":                  true,
	"crmCustomer/lock":                   true,
	"crmCustomer/setDealStatus":          true,
	"crmCustomer/queryCustomerSetting":   true,
	"crmCustomer/deleteCustomerSetting":  true,
	"crmCustomer/nearbyCustomer":         true,
	"crmCustomer/queryCustomerName":      true,
	"crmCustomer/hasFollowTask":          true,
	"crmCustomer/num":                    true,
	"crmCustomer/queryFileList":          true,
	// crmActivity
	"crmActivity/deleteOutworkSign":      true,
	"crmActivity/setPictureSetting":      true,
	"crmActivity/queryOutworkStats":      true,
}

// IsRequestParam returns true if the given API path uses @RequestParam.
func IsRequestParam(path string) bool {
	return requestParamPaths[path]
}

// NormalizePath strips leading/trailing slashes for lookup.
func NormalizePath(path string) string {
	return strings.Trim(path, "/")
}
```

- [ ] **Step 3: Create HTTP client**

```go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/orvibo/crm-cli/internal/api"
)

var envURLs = map[string]string{
	"prod":  "https://crm.orvibo.com/api/",
	"test":  "https://crm-test.orvibo.com/test/api/",
	"local": "http://localhost:8090/api/",
}

// DoRequest sends a request to the CRM API.
// method: HTTP method (POST, GET)
// path: API path (e.g. "crmCustomer/queryPageList")
// body: request body object (will be marshaled to JSON or query string)
// token: Admin-Token value
// envName: "prod", "test", or "local"
func DoRequest(method, path string, body interface{}, token, envName string) (*api.APIResponse, error) {
	baseURL, ok := envURLs[envName]
	if !ok {
		return nil, fmt.Errorf("unknown env: %s", envName)
	}

	fullURL := baseURL + path
	var req *http.Request
	var err error

	normalized := api.NormalizePath(path)

	if api.IsRequestParam(normalized) {
		// Send params as query string
		if body != nil {
			params := toQueryString(body)
			if strings.Contains(fullURL, "?") {
				fullURL += "&" + params
			} else {
				fullURL += "?" + params
			}
		}
		req, err = http.NewRequest(method, fullURL, nil)
	} else {
		// Send as JSON body
		var bodyReader io.Reader
		if body != nil {
			jsonData, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				return nil, fmt.Errorf("marshal body: %w", marshalErr)
			}
			bodyReader = bytes.NewReader(jsonData)
		}
		req, err = http.NewRequest(method, fullURL, bodyReader)
		if err == nil && body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Admin-Token", token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp api.APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %s", string(data))
	}

	return &apiResp, nil
}

// toQueryString converts a struct to URL query parameters using JSON tags.
func toQueryString(v interface{}) string {
	if v == nil {
		return ""
	}
	// Marshal to JSON then unmarshal to map, then encode as query
	jsonData, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		return ""
	}
	params := url.Values{}
	for key, val := range m {
		if val == nil {
			continue
		}
		switch v := val.(type) {
		case string:
			if v != "" {
				params.Set(key, v)
			}
		case float64:
			params.Set(key, fmt.Sprintf("%.0f", v))
		default:
			params.Set(key, fmt.Sprintf("%v", v))
		}
	}
	return params.Encode()
}

// CheckResponse returns an error for non-success API responses.
func CheckResponse(resp *api.APIResponse) error {
	switch resp.Code {
	case 0, 200:
		return nil
	case 302:
		return fmt.Errorf("token expired, run: crm-cli auth login")
	case 401:
		return fmt.Errorf("unauthorized: %s", resp.Msg)
	default:
		return fmt.Errorf("API error (code %d): %s", resp.Code, resp.Msg)
	}
}
```

- [ ] **Step 4: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 5: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/api/types.go internal/api/param_mode.go internal/client/http.go
git commit -m "feat: HTTP client with Admin-Token injection and RequestParam auto-detect"
```

---

## Task 4: Entity Registry

**Files:**
- Create: `internal/api/registry.go`

- [ ] **Step 1: Create entity registry**

```go
package api

import "fmt"

type EntityDef struct {
	Name      string
	Label     int
	APIPrefix string
}

var entities = map[string]EntityDef{
	"customer":    {Name: "customer", Label: 2, APIPrefix: "crmCustomer"},
	"leads":       {Name: "leads", Label: 1, APIPrefix: "crmLeads"},
	"contacts":    {Name: "contacts", Label: 3, APIPrefix: "crmContacts"},
	"business":    {Name: "business", Label: 5, APIPrefix: "crmBusiness"},
	"contract":    {Name: "contract", Label: 6, APIPrefix: "crmContract"},
	"receivables": {Name: "receivables", Label: 7, APIPrefix: "crmReceivables"},
	"plan":        {Name: "plan", Label: 8, APIPrefix: "crmReceivablesPlan"},
	"product":     {Name: "product", Label: 4, APIPrefix: "crmProduct"},
	"pool":        {Name: "pool", Label: 9, APIPrefix: "crmCustomerPool"},
	"visit":       {Name: "visit", Label: 17, APIPrefix: "crmReturnVisit"},
	"invoice":     {Name: "invoice", Label: 18, APIPrefix: "crmInvoice"},
}

func GetEntity(name string) (EntityDef, error) {
	e, ok := entities[name]
	if !ok {
		return EntityDef{}, fmt.Errorf("unknown entity: %s (valid: customer, leads, contacts, business, contract, receivables, plan, product, pool, visit, invoice)", name)
	}
	return e, nil
}

// EntityNames returns all registered entity names (for command registration).
func EntityNames() []string {
	names := make([]string, 0, len(entities))
	for k := range entities {
		names = append(names, k)
	}
	return names
}
```

- [ ] **Step 2: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 3: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/api/registry.go
git commit -m "feat: entity registry mapping names to API prefixes and CrmEnum labels"
```

---

## Task 5: Filter Parser

**Files:**
- Create: `internal/filter/parser.go`

- [ ] **Step 1: Create filter parser**

```go
package filter

import (
	"fmt"
	"strings"

	"github.com/orvibo/crm-cli/internal/api"
)

var opToEnum = map[string]int{
	"eq":          1,
	"ne":          2,
	"contains":    3,
	"notcontains": 4,
	"empty":       5,
	"notempty":    6,
	"gt":          7,
	"gte":         8,
	"lt":          9,
	"lte":         10,
	"prefix":      12,
	"suffix":      13,
	"range":       14,
}

// Default formType mapping for common CRM fields.
var fieldFormTypes = map[string]string{
	"customerName":   "text",
	"mobile":         "mobile",
	"telephone":      "mobile",
	"email":          "email",
	"website":        "text",
	"address":        "text",
	"dealStatus":     "select",
	"customerLevel":  "select",
	"industry":       "select",
	"source":         "select",
	"remark":         "text",
	"createTime":     "datetime",
	"updateTime":     "datetime",
	"ownerUserName":  "text",
	"lastContent":    "text",
	"nextTime":       "datetime",
	"followup":       "datetime",
	"name":           "text",
	"leadsName":      "text",
	"contactsName":   "text",
	"businessName":   "text",
	"contractName":   "text",
	"money":          "number",
	"contractMoney":  "number",
}

// ParseFilter parses a filter string "field:op:value" into a SearchItem.
func ParseFilter(raw string) (api.SearchItem, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 {
		return api.SearchItem{}, fmt.Errorf("invalid filter format, expected field:op[:value], got: %s", raw)
	}

	fieldName := parts[0]
	op := parts[1]
	var value string
	if len(parts) == 3 {
		value = parts[2]
	}

	enumVal, ok := opToEnum[op]
	if !ok {
		return api.SearchItem{}, fmt.Errorf("unknown operator: %s", op)
	}

	var values []string
	switch op {
	case "empty", "notempty":
		values = []string{}
	case "range":
		values = strings.Split(value, ",")
	default:
		if value != "" {
			values = []string{value}
		}
	}

	formType, ok := fieldFormTypes[fieldName]
	if !ok {
		formType = "text" // default fallback
	}

	return api.SearchItem{
		Name:     fieldName,
		FormType: formType,
		Type:     enumVal,
		Values:   values,
	}, nil
}

// ParseFilters parses multiple filter strings.
func ParseFilters(filters []string) ([]api.SearchItem, error) {
	var items []api.SearchItem
	for _, f := range filters {
		item, err := ParseFilter(f)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
```

- [ ] **Step 2: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 3: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/filter/parser.go
git commit -m "feat: --filter DSL parser converting field:op:value to SearchItem"
```

---

## Task 6: Output Layer

**Files:**
- Create: `internal/output/json.go`
- Create: `internal/output/table.go`

- [ ] **Step 1: Install tablewriter**

```bash
cd /Users/apple/orvibocli
go get github.com/olekukonez/tablewriter@latest
```

- [ ] **Step 2: Create JSON output**

```go
package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}

func PrintRawJSON(raw json.RawMessage) error {
	fmt.Println(string(raw))
	return nil
}
```

- [ ] **Step 3: Create table output**

```go
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonez/tablewriter"
)

// PrintTable renders a list of maps as a table.
// fields is comma-separated field names; if empty, uses all keys from the first row.
func PrintTable(rows []map[string]interface{}, fields string) error {
	if len(rows) == 0 {
		fmt.Println("No results.")
		return nil
	}

	var headers []string
	if fields != "" {
		headers = strings.Split(fields, ",")
	} else {
		for k := range rows[0] {
			headers = append(headers, k)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)

	for _, row := range rows {
		var vals []string
		for _, h := range headers {
			vals = append(vals, fmtVal(row[h]))
		}
		table.Append(vals)
	}

	table.Render()
	return nil
}

func fmtVal(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%f", val)
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
```

- [ ] **Step 4: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 5: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/output/json.go internal/output/table.go
git commit -m "feat: JSON and table output formatters"
```

---

## Task 7: Auth Command

**Files:**
- Create: `cmd/auth.go`

- [ ] **Step 1: Create auth command**

```go
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/orvibo/crm-cli/internal/client"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication management",
}

var loginUsername string
var loginPassword string

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with username and password",
	RunE: func(cmd *cobra.Command, args []string) error {
		username := loginUsername
		password := loginPassword

		if username == "" {
			fmt.Print("Username: ")
			reader := bufio.NewReader(os.Stdin)
			var err error
			username, err = reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read username: %w", err)
			}
			username = strings.TrimSpace(username)
		}

		if password == "" {
			fmt.Print("Password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			password = string(bytePassword)
		}

		loginBody := map[string]interface{}{
			"username":  username,
			"password":  password,
			"loginType": 1,
			"type":      1,
		}

		resp, err := client.DoRequest("POST", "login", loginBody, "", env)
		if err != nil {
			return fmt.Errorf("login request failed: %w", err)
		}
		if err := client.CheckResponse(resp); err != nil {
			return err
		}

		// Extract token from response data
		var loginResult struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(resp.Data, &loginResult); err != nil {
			// If token field not found, the token may be in a different structure.
			// Try to find it generically.
			var dataMap map[string]interface{}
			if err2 := json.Unmarshal(resp.Data, &dataMap); err2 == nil {
				if t, ok := dataMap["token"].(string); ok {
					loginResult.Token = t
				}
			}
		}

		if loginResult.Token == "" {
			return fmt.Errorf("no token in login response, raw data: %s", string(resp.Data))
		}

		cfg, _ := client.LoadConfig()
		cfg.Token = loginResult.Token
		cfg.Env = env
		if err := client.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Printf("Login successful. Token saved to ~/.crm-cli/config.json\n")
		return nil
	},
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Verify current token and show user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.LoadConfig()
		if err != nil {
			return err
		}
		if cfg.Token == "" {
			return fmt.Errorf("not logged in. Run: crm-cli auth login")
		}

		resp, err := client.DoRequest("POST", "adminUser/queryLoginUser", nil, cfg.Token, env)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		if err := client.CheckResponse(resp); err != nil {
			return err
		}

		fmt.Println(string(resp.Data))
		return nil
	},
}

func init() {
	authLoginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username")
	authLoginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authWhoamiCmd)
	rootCmd.AddCommand(authCmd)
}
```

Note: needs `go get golang.org/x/term` for password masking.

- [ ] **Step 2: Install term dependency**

```bash
cd /Users/apple/orvibocli
go get golang.org/x/term@latest
```

- [ ] **Step 3: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 4: Test manually**

```bash
cd /Users/apple/orvibocli
go build -o crm-cli . && ./crm-cli auth login -u testuser -p testpass
```

Expected: Either "Login successful" or an auth error from the API (not a Go panic).

- [ ] **Step 5: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/auth.go go.mod go.sum
git commit -m "feat: auth login (username+password) and whoami commands"
```

---

## Task 8: Entity Command (list / detail)

**Files:**
- Create: `cmd/entity.go`

- [ ] **Step 1: Create entity commands**

```go
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/orvibo/crm-cli/internal/api"
	"github.com/orvibo/crm-cli/internal/client"
	"github.com/orvibo/crm-cli/internal/filter"
	"github.com/orvibo/crm-cli/internal/output"
)

var (
	searchStr string
	filterStrs []string
	pageNum   int
	limitNum  int
	entityID  string
	fieldsStr string
	dryRun    bool
)

func init() {
	for _, name := range api.EntityNames() {
		entityCmd := &cobra.Command{
			Use:   name,
			Short: fmt.Sprintf("Operate on %s", name),
		}

		// list sub-command
		listCmd := &cobra.Command{
			Use:   "list",
			Short: fmt.Sprintf("List %s records", name),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runList(cmd.Root().Name())
			},
		}
		listCmd.Flags().StringVarP(&searchStr, "search", "s", "", "Search keyword")
		listCmd.Flags().StringArrayVar(&filterStrs, "filter", nil, "Filter: field:op:value (repeatable)")
		listCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page number")
		listCmd.Flags().IntVarP(&limitNum, "limit", "l", 15, "Page size")
		listCmd.Flags().StringVar(&fieldsStr, "fields", "", "Output fields (comma-separated)")

		// detail sub-command
		detailCmd := &cobra.Command{
			Use:   "detail",
			Short: fmt.Sprintf("Get %s detail", name),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runDetail(cmd.Root().Name())
			},
		}
		detailCmd.Flags().StringVar(&entityID, "id", "", "Entity ID (required)")
		detailCmd.MarkFlagRequired("id")

		entityCmd.AddCommand(listCmd)
		entityCmd.AddCommand(detailCmd)
		rootCmd.AddCommand(entityCmd)
	}
}

func runList(entityName string) error {
	ent, err := api.GetEntity(entityName)
	if err != nil {
		return err
	}

	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("not logged in. Run: crm-cli auth login")
	}

	searchItems, err := filter.ParseFilters(filterStrs)
	if err != nil {
		return err
	}

	body := api.SearchBO{
		Page:       pageNum,
		Limit:      limitNum,
		PageType:   1,
		Search:     searchStr,
		Label:      ent.Label,
		SearchList: searchItems,
	}

	path := ent.APIPrefix + "/queryPageList"

	if dryRun {
		jsonData, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("DRY RUN: POST %s\n%s\n", path, string(jsonData))
		return nil
	}

	resp, err := client.DoRequest("POST", path, body, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}

func runDetail(entityName string) error {
	ent, err := api.GetEntity(entityName)
	if err != nil {
		return err
	}

	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("not logged in. Run: crm-cli auth login")
	}

	path := fmt.Sprintf("%s/queryById/%s", ent.APIPrefix, entityID)

	if dryRun {
		fmt.Printf("DRY RUN: POST %s\n", path)
		return nil
	}

	resp, err := client.DoRequest("POST", path, nil, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}

func resolveEnv() string {
	if env != "" {
		return env
	}
	cfg, _ := client.LoadConfig()
	if cfg.Env != "" {
		return cfg.Env
	}
	return "prod"
}

func printOutput(data json.RawMessage) error {
	if format == "table" {
		// Try to extract list for table rendering
		var pageResult struct {
			List []map[string]interface{} `json:"list"`
		}
		if err := json.Unmarshal(data, &pageResult); err == nil && len(pageResult.List) > 0 {
			return output.PrintTable(pageResult.List, fieldsStr)
		}
		// Single object — render as key-value
		var obj map[string]interface{}
		if err := json.Unmarshal(data, &obj); err == nil {
			rows := []map[string]interface{}{obj}
			return output.PrintTable(rows, fieldsStr)
		}
	}
	return output.PrintRawJSON(data)
}
```

- [ ] **Step 2: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 3: Test with real API**

```bash
cd /Users/apple/orvibocli
go build -o crm-cli .
# First login, then:
./crm-cli customer list --search "华为" --limit 5
./crm-cli customer detail --id 2043193821684215808
./crm-cli leads list --limit 3
```

Expected: JSON output of customer/leads data.

- [ ] **Step 4: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/entity.go
git commit -m "feat: entity list/detail commands for all registered entities"
```

---

## Task 9: Activity Command

**Files:**
- Create: `cmd/activity.go`

- [ ] **Step 1: Create activity command**

```go
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/orvibo/crm-cli/internal/api"
	"github.com/orvibo/crm-cli/internal/client"
)

var (
	activityType string
	activityID   string
)

func init() {
	activityCmd := &cobra.Command{
		Use:   "activity",
		Short: "Activity / follow-up records",
	}

	activityListCmd := &cobra.Command{
		Use:   "list",
		Short: "List activity records for an entity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runActivityList()
		},
	}
	activityListCmd.Flags().StringVar(&activityType, "type", "", "Entity type: customer, leads, contacts, business, etc. (required)")
	activityListCmd.Flags().StringVar(&activityID, "id", "", "Entity ID (required)")
	activityListCmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page number")
	activityListCmd.Flags().IntVarP(&limitNum, "limit", "l", 15, "Page size")
	activityListCmd.MarkFlagRequired("type")
	activityListCmd.MarkFlagRequired("id")

	activityCmd.AddCommand(activityListCmd)
	rootCmd.AddCommand(activityCmd)
}

func runActivityList() error {
	ent, err := api.GetEntity(activityType)
	if err != nil {
		return err
	}

	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("not logged in. Run: crm-cli auth login")
	}

	// CrmActivityQueryBO fields
	body := map[string]interface{}{
		"page":           pageNum,
		"limit":          limitNum,
		"pageType":       1,
		"activityType":   ent.Label,
		"activityTypeId": activityID,
	}

	resp, err := client.DoRequest("POST", "crmActivity/getCrmActivityPageList", body, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}
```

- [ ] **Step 2: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 3: Test with real API**

```bash
cd /Users/apple/orvibocli
./crm-cli activity list --type customer --id 2043193821684215808
```

Expected: JSON output with activity/follow-up records.

- [ ] **Step 4: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/activity.go
git commit -m "feat: activity list command for follow-up records"
```

---

## Task 10: API Raw Command

**Files:**
- Create: `cmd/api.go`

- [ ] **Step 1: Create api command**

```go
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/orvibo/crm-cli/internal/client"
)

var (
	apiData  string
	apiQuery string
)

func init() {
	apiCmd := &cobra.Command{
		Use:   "api [METHOD] [PATH]",
		Short: "Make raw API calls",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPI(args[0], args[1])
		},
	}
	apiCmd.Flags().StringVar(&apiData, "data", "", "JSON body")
	apiCmd.Flags().StringVar(&apiQuery, "query", "", "URL query string (for @RequestParam endpoints)")
	apiCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print request without sending")

	rootCmd.AddCommand(apiCmd)
}

func runAPI(method, path string) error {
	cfg, err := client.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.Token == "" {
		return fmt.Errorf("not logged in. Run: crm-cli auth login")
	}

	var body interface{}
	if apiData != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(apiData), &parsed); err != nil {
			return fmt.Errorf("invalid JSON in --data: %w", err)
		}
		body = parsed
	}

	// If --query is provided, append to URL and send with no body
	if apiQuery != "" {
		path = path + "?" + apiQuery
		body = nil
	}

	if dryRun {
		fmt.Printf("DRY RUN: %s %s\n", method, path)
		if body != nil {
			jsonData, _ := json.MarshalIndent(body, "", "  ")
			fmt.Println(string(jsonData))
		}
		return nil
	}

	resp, err := client.DoRequest(method, path, body, cfg.Token, resolveEnv())
	if err != nil {
		return err
	}
	if err := client.CheckResponse(resp); err != nil {
		return err
	}

	return printOutput(resp.Data)
}
```

- [ ] **Step 2: Verify build**

```bash
cd /Users/apple/orvibocli
go build ./...
```

- [ ] **Step 3: Test with real API**

```bash
cd /Users/apple/orvibocli
./crm-cli api POST crmCustomer/queryPageList --data '{"page":1,"limit":3,"label":2}'
./crm-cli api POST crmActionRecord/queryRecordList --query 'actionId=2043193821684215808&types=2'
```

Expected: JSON output from the raw API call.

- [ ] **Step 4: Commit**

```bash
cd /Users/apple/orvibocli
git add cmd/api.go
git commit -m "feat: api raw command for direct HTTP calls"
```

---

## Task 11: Skills

**Files:**
- Create: `skills/crm-shared/SKILL.md`
- Create: `skills/crm-customer/SKILL.md`

- [ ] **Step 1: Create crm-shared SKILL.md**

```markdown
---
name: crm-shared
description: 欧瑞博 CRM 全局约定和认证配置。所有 crm-* skill 的前置依赖。
requires:
  bins: ["crm-cli"]
---

## 认证

使用 `crm-cli auth` 管理登录状态：

```bash
# 首次使用
crm-cli auth login -u <用户名> -p <密码>

# 验证登录状态
crm-cli auth whoami
```

Token 过期时会提示 `token expired, run: crm-cli auth login`。

## 全局约定

- 所有命令默认输出 JSON（AI 消费），`--format table` 可选
- 环境切换：`--env prod|test|local`（默认 prod）
- 分页：`--page 1 --limit 15`（默认值）
- 搜索：`--search "关键词"` 进行全局搜索
- 高级筛选：`--filter 'field:op:value'`（可重复）
- Dry run：`--dry-run` 打印请求不发送（写操作）

## 实体类型 (CrmEnum)

| entity | CrmEnum | 说明 |
|---|---|---|
| customer | 2 | 客户 |
| leads | 1 | 线索 |
| contacts | 3 | 联系人 |
| business | 5 | 商机 |
| contract | 6 | 合同 |
| receivables | 7 | 回款 |
| plan | 8 | 回款计划 |
| product | 4 | 产品 |
| pool | 9 | 公海 |
| visit | 17 | 回访 |
| invoice | 18 | 发票 |

## 错误码

| code | 含义 |
|---|---|
| 0, 200 | 成功 |
| 302 | 未登录（Token 过期） |
| 400 | 参数错误 |
| 401 | 无权操作 |
| 500 | 服务器错误 |

## 兜底命令

当快捷命令不够用时，直接调原始 API：

```bash
# JSON body 接口
crm-cli api POST crmCustomer/queryPageList --data '{"page":1,"limit":10}'

# RequestParam 接口（参数拼在 URL 上）
crm-cli api POST crmActionRecord/queryRecordList --query 'actionId=123&types=2'
```
```

- [ ] **Step 2: Create crm-customer SKILL.md**

```markdown
---
name: crm-customer
description: 查询、搜索欧瑞博 CRM 客户数据。当用户需要查找客户、查看客户详情、查看客户跟进记录时使用。
requires:
  bins: ["crm-cli"]
  skills: ["crm-shared"]
---

## 前置条件

先确认认证状态（见 ../crm-shared/SKILL.md）。如果 `crm-cli auth whoami` 报错，先登录。

## 常用命令

### 按关键词搜索客户

```bash
crm-cli customer list --search "华为"
crm-cli customer list --search "13126987062" --limit 5
```

### 高级筛选

```bash
# 按成交状态筛选
crm-cli customer list --filter 'dealStatus:eq:1'

# 按客户级别筛选
crm-cli customer list --filter 'customerLevel:eq:A级'

# 组合筛选
crm-cli customer list --filter 'dealStatus:eq:1' --filter 'customerLevel:eq:A级'
```

### 客户详情

```bash
crm-cli customer detail --id <customerId>
```

### 客户跟进记录

```bash
crm-cli activity list --type customer --id <customerId>
```

### 客户操作记录（字段变更）

```bash
crm-cli api POST crmActionRecord/queryRecordList --query 'actionId=<customerId>&types=2'
```

### 客户下关联数据

```bash
# 联系人
crm-cli api POST crmCustomer/queryContacts --data '{"customerId":<id>,"page":1,"limit":15}'

# 商机
crm-cli api POST crmCustomer/queryBusiness --data '{"customerId":<id>,"page":1,"limit":15}'

# 合同
crm-cli api POST crmCustomer/queryContract --data '{"customerId":<id>,"page":1,"limit":15}'
```

## 注意事项

- `customerId` 是长整型 ID（如 `2043193821684215808`），不是客户名称
- `dealStatus`：0=未成交，1=已成交，2=已成交（历史）
- 搜索关键词会同时匹配客户名、手机号、电话等字段
- 默认返回 15 条，用 `--limit` 调整（最大 1000）
- 表格模式加 `--format table`
```

- [ ] **Step 3: Commit**

```bash
cd /Users/apple/orvibocli
git add skills/
git commit -m "feat: crm-shared and crm-customer Claude Code skills"
```

---

## Task 12: Integration Test

**Files:**
- Create: `internal/filter/parser_test.go`

- [ ] **Step 1: Write filter parser tests**

```go
package filter

import (
	"testing"
)

func TestParseFilter_Eq(t *testing.T) {
	item, err := ParseFilter("mobile:eq:13800138000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Name != "mobile" {
		t.Errorf("expected name=mobile, got %s", item.Name)
	}
	if item.Type != 1 {
		t.Errorf("expected type=1, got %d", item.Type)
	}
	if item.FormType != "mobile" {
		t.Errorf("expected formType=mobile, got %s", item.FormType)
	}
	if len(item.Values) != 1 || item.Values[0] != "13800138000" {
		t.Errorf("expected values=[13800138000], got %v", item.Values)
	}
}

func TestParseFilter_Contains(t *testing.T) {
	item, err := ParseFilter("customerName:contains:华为")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != 3 {
		t.Errorf("expected type=3, got %d", item.Type)
	}
	if item.FormType != "text" {
		t.Errorf("expected formType=text, got %s", item.FormType)
	}
}

func TestParseFilter_Empty(t *testing.T) {
	item, err := ParseFilter("mobile:empty:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != 5 {
		t.Errorf("expected type=5, got %d", item.Type)
	}
	if len(item.Values) != 0 {
		t.Errorf("expected empty values, got %v", item.Values)
	}
}

func TestParseFilter_Range(t *testing.T) {
	item, err := ParseFilter("money:range:100,500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Type != 14 {
		t.Errorf("expected type=14, got %d", item.Type)
	}
	if len(item.Values) != 2 || item.Values[0] != "100" || item.Values[1] != "500" {
		t.Errorf("expected values=[100,500], got %v", item.Values)
	}
}

func TestParseFilter_InvalidOp(t *testing.T) {
	_, err := ParseFilter("mobile:badop:value")
	if err == nil {
		t.Fatal("expected error for invalid operator")
	}
}

func TestParseFilter_InvalidFormat(t *testing.T) {
	_, err := ParseFilter("nocolon")
	if err == nil {
		t.Fatal("expected error for missing colon")
	}
}

func TestParseFilters(t *testing.T) {
	items, err := ParseFilters([]string{"mobile:eq:123", "customerName:contains:华为"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestParseFilter_DefaultFormType(t *testing.T) {
	item, err := ParseFilter("unknownField:eq:val")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.FormType != "text" {
		t.Errorf("expected default formType=text, got %s", item.FormType)
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd /Users/apple/orvibocli
go test ./internal/filter/ -v
```

Expected: All 7 tests pass.

- [ ] **Step 3: Commit**

```bash
cd /Users/apple/orvibocli
git add internal/filter/parser_test.go
git commit -m "test: filter parser unit tests"
```

---

## Task 13: End-to-End Smoke Test

No new files. Manual verification.

- [ ] **Step 1: Build binary**

```bash
cd /Users/apple/orvibocli
go build -o crm-cli .
```

- [ ] **Step 2: Test auth**

```bash
./crm-cli auth login -u <username> -p <password>
./crm-cli auth whoami
```

- [ ] **Step 3: Test customer list**

```bash
./crm-cli customer list --limit 3
./crm-cli customer list --search "华为" --format table
```

- [ ] **Step 4: Test customer detail**

```bash
./crm-cli customer detail --id <some_customer_id>
```

- [ ] **Step 5: Test leads**

```bash
./crm-cli leads list --limit 3
```

- [ ] **Step 6: Test activity**

```bash
./crm-cli activity list --type customer --id <customer_id>
```

- [ ] **Step 7: Test api raw**

```bash
./crm-cli api POST crmCustomer/queryPageList --data '{"page":1,"limit":2,"label":2}'
```

- [ ] **Step 8: Final commit**

```bash
cd /Users/apple/orvibocli
git add -A
git commit -m "chore: MVP complete, e2e smoke test passed"
```
