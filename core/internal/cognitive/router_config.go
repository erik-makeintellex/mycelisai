package cognitive

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"gopkg.in/yaml.v3"
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
				APIKey  string `json:"api_key"`
			}
			if err := json.Unmarshal(configJSON, &extra); err == nil {
				if strings.TrimSpace(extra.ModelID) != "" {
					pConfig.ModelID = extra.ModelID
				}
				// extra.APIKey is intentionally never assigned to
				// pConfig.AuthKey: a literal key in the DB overlay's config
				// JSON is just as unusable as one in cognitive.yaml (AuthKey
				// is yaml:"-" and never persisted/loaded from a tracked
				// source). Warn once, naming the provider only, and flag it
				// so adapter init returns explicit guidance instead of a
				// generic missing-key error.
				if strings.TrimSpace(extra.APIKey) != "" {
					log.Printf("WARN: provider %q has a literal api_key in the DB llm_providers.config overlay; it is ignored. Configure api_key_env with a secret reference instead.", id)
					pConfig.LiteralAPIKeyIgnored = true
				}
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
		// A blank role.* value is treated as if the row were absent: it
		// never deletes the YAML default and never records an override.
		if strings.TrimSpace(providerID) == "" {
			continue
		}
		recordProfileOverride(config, profileName, strings.TrimSpace(providerID), ProfileOriginDB)
	}

	return rows2.Err()
}

// detectLiteralProviderAPIKeys scans raw cognitive.yaml bytes for a
// non-empty literal `api_key` per provider. ProviderConfig.AuthKey is
// yaml:"-" (never loaded from a tracked file, by design — see the L-slice
// security fix), so a literal key in the file is otherwise silently
// unusable. This reports which provider IDs have one, without ever
// inspecting or returning the value itself, so NewRouter can warn the
// operator and flag those providers for an explicit adapter-init error.
func detectLiteralProviderAPIKeys(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	var probe struct {
		Providers map[string]struct {
			AuthKey string `yaml:"api_key"`
		} `yaml:"providers"`
	}
	if err := yaml.Unmarshal(data, &probe); err != nil {
		return nil
	}
	var found []string
	for id, p := range probe.Providers {
		if strings.TrimSpace(p.AuthKey) != "" {
			found = append(found, id)
		}
	}
	return found
}
