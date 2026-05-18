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

		var loginResult struct {
			Token      string `json:"token"`
			AdminToken string `json:"adminToken"`
		}
		if err := json.Unmarshal(resp.Data, &loginResult); err != nil {
			var dataMap map[string]interface{}
			if err2 := json.Unmarshal(resp.Data, &dataMap); err2 == nil {
				if t, ok := dataMap["adminToken"].(string); ok {
					loginResult.AdminToken = t
				} else if t, ok := dataMap["token"].(string); ok {
					loginResult.Token = t
				}
			}
		}

		token := loginResult.AdminToken
		if token == "" {
			token = loginResult.Token
		}
		if token == "" {
			return fmt.Errorf("no token in login response, raw data: %s", string(resp.Data))
		}

		cfg, _ := client.LoadConfig()
		cfg.Token = token
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
