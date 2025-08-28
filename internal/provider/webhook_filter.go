package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookFilterValue struct {
	Not    TypedObject[WebhookFilterNotValue]    `tfsdk:"not"`
	Equals TypedObject[WebhookFilterEqualsValue] `tfsdk:"equals"`
	In     TypedObject[WebhookFilterInValue]     `tfsdk:"in"`
	Regexp TypedObject[WebhookFilterRegexpValue] `tfsdk:"regexp"`
}

type WebhookFilterNotValue struct {
	Equals TypedObject[WebhookFilterEqualsValue] `tfsdk:"equals"`
	In     TypedObject[WebhookFilterInValue]     `tfsdk:"in"`
	Regexp TypedObject[WebhookFilterRegexpValue] `tfsdk:"regexp"`
}

type WebhookFilterEqualsValue struct {
	Doc   types.String `tfsdk:"doc"`
	Value types.String `tfsdk:"value"`
}

type WebhookFilterInValue struct {
	Doc    types.String            `tfsdk:"doc"`
	Values TypedList[types.String] `tfsdk:"values"`
}

type WebhookFilterRegexpValue struct {
	Doc     types.String `tfsdk:"doc"`
	Pattern types.String `tfsdk:"pattern"`
}
