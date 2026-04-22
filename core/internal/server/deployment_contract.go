package server

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	deploymentContractPathEnv      = "MYCELIS_DEPLOYMENT_CONTRACT_PATH"
	accessManagementTierEnv        = "MYCELIS_ACCESS_MANAGEMENT_TIER"
	productEditionEnv              = "MYCELIS_PRODUCT_EDITION"
	identityModeEnv                = "MYCELIS_IDENTITY_MODE"
	sharedAgentSpecificityOwnerEnv = "MYCELIS_SHARED_AGENT_SPECIFICITY_OWNER"
)

// DeploymentContract is the deploy-owned edition/auth posture surface exposed to
// runtime handlers and startup validation.
type DeploymentContract struct {
	AccessManagementTier        string `json:"access_management_tier"`
	ProductEdition              string `json:"product_edition"`
	IdentityMode                string `json:"identity_mode"`
	SharedAgentSpecificityOwner string `json:"shared_agent_specificity_owner"`
}

type deploymentContractFile struct {
	AccessManagementTier        any `json:"access_management_tier"`
	ProductEdition              any `json:"product_edition"`
	IdentityMode                any `json:"identity_mode"`
	SharedAgentSpecificityOwner any `json:"shared_agent_specificity_owner"`
}

func defaultDeploymentContract() DeploymentContract {
	return DeploymentContract{
		AccessManagementTier:        "release",
		ProductEdition:              "self_hosted_release",
		IdentityMode:                "local_only",
		SharedAgentSpecificityOwner: "root_admin",
	}
}

func NormalizeAccessManagementTier(v any) string {
	switch strings.TrimSpace(strings.ToLower(strings.TrimSpace(toString(v)))) {
	case "enterprise":
		return "enterprise"
	default:
		return "release"
	}
}

func NormalizeProductEdition(v any) string {
	switch strings.TrimSpace(strings.ToLower(strings.TrimSpace(toString(v)))) {
	case "enterprise", "self_hosted_enterprise":
		return "self_hosted_enterprise"
	case "hosted", "hosted_control_plane":
		return "hosted_control_plane"
	default:
		return "self_hosted_release"
	}
}

func NormalizeIdentityMode(v any) string {
	switch strings.TrimSpace(strings.ToLower(strings.TrimSpace(toString(v)))) {
	case "federated":
		return "federated"
	case "hybrid":
		return "hybrid"
	default:
		return "local_only"
	}
}

func NormalizeSharedAgentSpecificityOwner(v any) string {
	switch strings.TrimSpace(strings.ToLower(strings.TrimSpace(toString(v)))) {
	case "delegated_owner":
		return "delegated_owner"
	default:
		return "root_admin"
	}
}

func ResolveDeploymentContract() DeploymentContract {
	return resolveDeploymentContract(os.Getenv, os.ReadFile)
}

func resolveDeploymentContract(getenv func(string) string, readFile func(string) ([]byte, error)) DeploymentContract {
	contract := defaultDeploymentContract()

	if path := strings.TrimSpace(getenv(deploymentContractPathEnv)); path != "" {
		if data, err := readFile(path); err == nil {
			var raw deploymentContractFile
			if json.Unmarshal(data, &raw) == nil {
				contract.AccessManagementTier = NormalizeAccessManagementTier(raw.AccessManagementTier)
				contract.ProductEdition = NormalizeProductEdition(raw.ProductEdition)
				contract.IdentityMode = NormalizeIdentityMode(raw.IdentityMode)
				contract.SharedAgentSpecificityOwner = NormalizeSharedAgentSpecificityOwner(raw.SharedAgentSpecificityOwner)
			}
		}
	}

	if value := strings.TrimSpace(getenv(accessManagementTierEnv)); value != "" {
		contract.AccessManagementTier = NormalizeAccessManagementTier(value)
	}
	if value := strings.TrimSpace(getenv(productEditionEnv)); value != "" {
		contract.ProductEdition = NormalizeProductEdition(value)
	}
	if value := strings.TrimSpace(getenv(identityModeEnv)); value != "" {
		contract.IdentityMode = NormalizeIdentityMode(value)
	}
	if value := strings.TrimSpace(getenv(sharedAgentSpecificityOwnerEnv)); value != "" {
		contract.SharedAgentSpecificityOwner = NormalizeSharedAgentSpecificityOwner(value)
	}

	return contract
}

func (c DeploymentContract) ApplyUserSettings(settings map[string]any) map[string]any {
	settings["access_management_tier"] = c.AccessManagementTier
	settings["product_edition"] = c.ProductEdition
	settings["identity_mode"] = c.IdentityMode
	settings["shared_agent_specificity_owner"] = c.SharedAgentSpecificityOwner
	return settings
}

func (c DeploymentContract) RequiresBreakGlassRecovery() bool {
	if c.IdentityMode != "local_only" {
		return true
	}
	if c.AccessManagementTier == "enterprise" {
		return true
	}
	switch c.ProductEdition {
	case "self_hosted_enterprise", "hosted_control_plane":
		return true
	default:
		return false
	}
}

func toString(v any) string {
	switch typed := v.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(v)
	}
}
