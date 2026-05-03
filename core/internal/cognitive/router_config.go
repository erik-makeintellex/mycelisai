package cognitive

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
)

func loadFromDB(db *sql.DB, config *BrainConfig) error {
	rows, err := db.Query("SELECT id, driver, base_url, api_key_env_var, config FROM llm_providers")
	if err != nil {
		return err
	}
	defer rows.Close()

	if config.Providers == nil {
		config.Providers = make(map[string]ProviderConfig)
	}

	for rows.Next() {
		var id, driver, baseURL string
		var envVar sql.NullString
		var configJSON []byte

		if err := rows.Scan(&id, &driver, &baseURL, &envVar, &configJSON); err != nil {
			log.Printf("WARN: Skipping bad provider row: %v", err)
			continue
		}

		pConfig := config.Providers[id]
		pConfig.Driver = driver
		if strings.TrimSpace(pConfig.Type) == "" {
			pConfig.Type = driver
		}
		if strings.TrimSpace(baseURL) != "" {
			pConfig.Endpoint = baseURL
		}
		if envVar.Valid && strings.TrimSpace(envVar.String) != "" {
			pConfig.AuthKeyEnv = envVar.String
		}
		if len(configJSON) > 0 {
			var extra struct {
				ModelID string `json:"model_id"`
			}
			if err := json.Unmarshal(configJSON, &extra); err == nil && strings.TrimSpace(extra.ModelID) != "" {
				pConfig.ModelID = extra.ModelID
			}
		}
		config.Providers[id] = pConfig
	}

	rows2, err := db.Query("SELECT key, value FROM system_config WHERE key LIKE 'role.%'")
	if err != nil {
		return err
	}
	defer rows2.Close()

	if config.Profiles == nil {
		config.Profiles = make(map[string]string)
	}

	for rows2.Next() {
		var key, providerID string
		if err := rows2.Scan(&key, &providerID); err != nil {
			continue
		}
		profileName := strings.TrimPrefix(key, "role.")
		config.Profiles[profileName] = providerID
	}

	return nil
}
