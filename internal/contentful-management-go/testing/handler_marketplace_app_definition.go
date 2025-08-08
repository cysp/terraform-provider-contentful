package testing

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetMarketplaceAppDefinitions(_ context.Context, params cm.GetMarketplaceAppDefinitionsParams) (cm.GetMarketplaceAppDefinitionsRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	response := cm.GetMarketplaceAppDefinitionsOK{
		Sys: cm.GetMarketplaceAppDefinitionsOKSys{
			Type: cm.GetMarketplaceAppDefinitionsOKSysTypeArray,
		},
		Total: 0,
		Items: []cm.AppDefinition{},
	}

	for _, appDefinitionID := range params.SysIDIn {
		appDefinition, appDefinitionFound := ts.marketplaceAppDefinitions[appDefinitionID]
		if !appDefinitionFound {
			continue
		}

		response.Items = append(response.Items, *appDefinition)
		response.Total++
	}

	return &response, nil
}
