package create

import (
	"fmt"
	"net/http"

	"github.com/cli/cli/api"
)

// repoCreateInput represents input parameters for repoCreate
type repoCreateInput struct {
	Name        string `json:"name"`
	Visibility  string `json:"visibility"`
	HomepageURL string `json:"homepageUrl,omitempty"`
	Description string `json:"description,omitempty"`

	OwnerID string `json:"ownerId,omitempty"`
	TeamID  string `json:"teamId,omitempty"`

	HasIssuesEnabled bool `json:"hasIssuesEnabled"`
	HasWikiEnabled   bool `json:"hasWikiEnabled"`
}

// repoCreate creates a new GitHub repository
func repoCreate(client *http.Client, hostname string, input repoCreateInput) (*api.Repository, error) {
	apiClient := api.NewClientFromHTTP(client)

	var response struct {
		CreateRepository struct {
			Repository api.Repository
		}
	}

	if input.TeamID != "" {
		orgID, teamID, err := resolveOrganizationTeam(apiClient, hostname, input.OwnerID, input.TeamID)
		if err != nil {
			return nil, err
		}
		input.TeamID = teamID
		input.OwnerID = orgID
	} else if input.OwnerID != "" {
		orgID, err := resolveOrganization(apiClient, hostname, input.OwnerID)
		if err != nil {
			return nil, err
		}
		input.OwnerID = orgID
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := apiClient.GraphQL(hostname, `
	mutation RepositoryCreate($input: CreateRepositoryInput!) {
		createRepository(input: $input) {
			repository {
				id
				name
				owner { login }
				url
			}
		}
	}
	`, variables, &response)
	if err != nil {
		return nil, err
	}

	return api.InitRepoHostname(&response.CreateRepository.Repository, hostname), nil
}

// using API v3 here because the equivalent in GraphQL needs `read:org` scope
func resolveOrganization(client *api.Client, hostname, orgName string) (string, error) {
	var response struct {
		NodeID string `json:"node_id"`
	}
	err := client.REST(hostname, "GET", fmt.Sprintf("users/%s", orgName), nil, &response)
	return response.NodeID, err
}

// using API v3 here because the equivalent in GraphQL needs `read:org` scope
func resolveOrganizationTeam(client *api.Client, hostname, orgName, teamSlug string) (string, string, error) {
	var response struct {
		NodeID       string `json:"node_id"`
		Organization struct {
			NodeID string `json:"node_id"`
		}
	}
	err := client.REST(hostname, "GET", fmt.Sprintf("orgs/%s/teams/%s", orgName, teamSlug), nil, &response)
	return response.Organization.NodeID, response.NodeID, err
}
