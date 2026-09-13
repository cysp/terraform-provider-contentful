package provider

import (
	"context"
	"errors"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = (*appEventSubscriptionResource)(nil)
	_ resource.ResourceWithConfigure   = (*appEventSubscriptionResource)(nil)
	_ resource.ResourceWithIdentity    = (*appEventSubscriptionResource)(nil)
	_ resource.ResourceWithImportState = (*appEventSubscriptionResource)(nil)
)

//nolint:ireturn
func NewAppEventSubscriptionResource() resource.Resource { return &appEventSubscriptionResource{} }

type appEventSubscriptionResource struct{ providerData ContentfulProviderData }

func appEventSubscriptionIdentityAttributeNames() []string {
	return []string{"organization_id", "app_definition_id"}
}

func (r *appEventSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_event_subscription"
}
func (r *appEventSubscriptionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppEventSubscriptionResourceSchema(ctx)
}
func (r *appEventSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}
func (r *appEventSubscriptionResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(appEventSubscriptionIdentityAttributeNames())
}
func (r *appEventSubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, appEventSubscriptionIdentityAttributeNames(), req, resp)
}

func (r *appEventSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppEventSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	data, diags, consistencyDiags := r.put(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appEventSubscriptionIdentityAttributeNames(), &data)...)
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *appEventSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppEventSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	organization, appDefinition, diags := appEventSubscriptionIDs(state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.GetAppEventSubscription(ctx, cm.GetAppEventSubscriptionParams{OrganizationID: organization, AppDefinitionID: appDefinition})
	if err == nil && contentfulResponseIsNotFound(response) {
		resp.State.RemoveResource(ctx)

		return
	}

	result, ok := response.(*cm.AppEventSubscription)
	if err != nil || !ok || result == nil {
		resp.Diagnostics.AddError("Failed to read app event subscription", appEventSubscriptionErrorDetail(response, err))

		return
	}

	data, projectionDiags, identityDiags := appEventSubscriptionResponse(ctx, *result, state)
	resp.Diagnostics.Append(projectionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appEventSubscriptionIdentityAttributeNames(), &data)...)
	resp.Diagnostics.Append(identityDiags...)
}

func (r *appEventSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AppEventSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Topics.Equal(state.Topics) && plan.TargetURL.Equal(state.TargetURL) && plan.FilterFunctionID.Equal(state.FilterFunctionID) && plan.TransformationFunctionID.Equal(state.TransformationFunctionID) && plan.HandlerFunctionID.Equal(state.HandlerFunctionID) && plan.OrganizationID.Equal(state.OrganizationID) && plan.AppDefinitionID.Equal(state.AppDefinitionID) {
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appEventSubscriptionIdentityAttributeNames(), &state)...)

		return
	}

	ctx, cancel, diags := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	data, diags, consistencyDiags := r.put(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appEventSubscriptionIdentityAttributeNames(), &data)...)
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *appEventSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppEventSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	organization, appDefinition, diags := appEventSubscriptionIDs(state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.DeleteAppEventSubscription(ctx, cm.DeleteAppEventSubscriptionParams{OrganizationID: organization, AppDefinitionID: appDefinition})
	if err == nil {
		if _, ok := response.(*cm.NoContent); ok {
			return
		}

		if contentfulResponseIsNotFound(response) {
			return
		}
	}

	resp.Diagnostics.AddError("Failed to delete app event subscription", appEventSubscriptionErrorDetail(response, err))
}

// put is the shared complete-document upsert boundary. It does not replay or
// infer success after ambiguous writes; the provider transport's ordinary
// mutation retry policy applies. There is no version precondition on this API.
func (r *appEventSubscriptionResource) put(ctx context.Context, plan AppEventSubscriptionModel) (AppEventSubscriptionModel, diag.Diagnostics, diag.Diagnostics) {
	organization, appDefinition, diags := appEventSubscriptionIDs(plan)
	request, requestDiags := plan.ToAppEventSubscriptionData()
	diags.Append(requestDiags...)

	if diags.HasError() {
		return AppEventSubscriptionModel{}, diags, nil
	}

	response, err := r.providerData.client.PutAppEventSubscription(ctx, &request, cm.PutAppEventSubscriptionParams{OrganizationID: organization, AppDefinitionID: appDefinition})
	if err == nil {
		switch result := response.(type) {
		case *cm.PutAppEventSubscriptionOK:
			return reconcileAppEventSubscriptionResponse(ctx, cm.AppEventSubscription(*result), plan)
		case *cm.PutAppEventSubscriptionCreated:
			return reconcileAppEventSubscriptionResponse(ctx, cm.AppEventSubscription(*result), plan)
		}
	}

	diags.AddError("Failed to upsert app event subscription", appEventSubscriptionErrorDetail(response, err)+" A failed response does not establish whether the write was stored. Refresh an existing resource, or inspect and import this singleton before retrying an ambiguous create.")

	return AppEventSubscriptionModel{}, diags, nil
}

// Remote errors may echo sensitive URLs even in sys.id. Only numeric status
// and allowlisted classifications are exposed, never raw response/error text.
func appEventSubscriptionErrorDetail(response any, err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "The Contentful request exceeded its deadline."
	case errors.Is(err, context.Canceled):
		return "The Contentful request was canceled."
	}

	detail := "Contentful did not return a usable app event subscription response."
	if status, ok := response.(interface{ GetStatusCode() int }); ok {
		detail = fmt.Sprintf("Contentful returned HTTP %d.", status.GetStatusCode())
	}

	if errorResponse, ok := response.(cm.ErrorResponse); ok {
		if remote, hasError := errorResponse.GetError(); hasError {
			switch remote.Sys.ID {
			case "NotFound", "AccessDenied", "ValidationFailed", "UnprocessableEntity", "RateLimitExceeded", "VersionMismatch":
				detail += " " + remote.Sys.ID + "."
			}
		}
	}

	return detail
}
