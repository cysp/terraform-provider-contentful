package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func EditorInterfaceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"editor_layout": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceEditorLayoutItemValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceEditorLayoutItemValue{}.CustomType(ctx),
				},
				CustomType: TypedList[EditorInterfaceEditorLayoutItemValue]{}.CustomType(ctx),
				Optional:   true,
			},
			"controls": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceControlValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceControlValue{}.CustomType(ctx),
				},
				CustomType: TypedList[EditorInterfaceControlValue]{}.CustomType(ctx),
				Optional:   true,
			},
			"group_controls": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceGroupControlValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceGroupControlValue{}.CustomType(ctx),
				},
				CustomType: TypedList[EditorInterfaceGroupControlValue]{}.CustomType(ctx),
				Optional:   true,
			},
			"sidebar": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceSidebarValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceSidebarValue{}.CustomType(ctx),
				},
				CustomType: TypedList[EditorInterfaceSidebarValue]{}.CustomType(ctx),
				Optional:   true,
			},
		},
	}
}

func (v EditorInterfaceControlValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Required: true,
		},
		"widget_namespace": schema.StringAttribute{
			Optional: true,
		},
		"widget_id": schema.StringAttribute{
			Optional: true,
		},
		"settings": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Required: true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Required: true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field": schema.SingleNestedAttribute{
			Attributes: EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{}.SchemaAttributes(ctx),
			CustomType: EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{}.CustomType(ctx),
			Required:   true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"items": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{}.SchemaAttributes(ctx),
				CustomType: EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{}.CustomType(ctx),
			},
			CustomType: NewTypedListNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]().CustomType(ctx),
			Required:   true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field": schema.SingleNestedAttribute{
			Attributes: EditorInterfaceEditorLayoutItemGroupItemFieldValue{}.SchemaAttributes(ctx),
			CustomType: EditorInterfaceEditorLayoutItemGroupItemFieldValue{}.CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("field"),
					path.MatchRelative().AtParent().AtName("group"),
				),
			},
		},
		"group": schema.SingleNestedAttribute{
			Attributes: EditorInterfaceEditorLayoutItemGroupItemGroupValue{}.SchemaAttributes(ctx),
			CustomType: EditorInterfaceEditorLayoutItemGroupItemGroupValue{}.CustomType(ctx),
			Optional:   true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("field"),
					path.MatchRelative().AtParent().AtName("group"),
				),
			},
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"items": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: EditorInterfaceEditorLayoutItemGroupItemValue{}.SchemaAttributes(ctx),
				CustomType: EditorInterfaceEditorLayoutItemGroupItemValue{}.CustomType(ctx),
			},
			CustomType: NewTypedListNull[EditorInterfaceEditorLayoutItemGroupItemValue]().CustomType(ctx),
			Required:   true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group": schema.SingleNestedAttribute{
			Attributes: EditorInterfaceEditorLayoutItemGroupValue{}.SchemaAttributes(ctx),
			CustomType: EditorInterfaceEditorLayoutItemGroupValue{}.CustomType(ctx),
			Required:   true,
		},
	}
}

func (v EditorInterfaceGroupControlValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Required: true,
		},
		"widget_namespace": schema.StringAttribute{
			Optional: true,
		},
		"widget_id": schema.StringAttribute{
			Optional: true,
		},
		"settings": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
	}
}

func (v EditorInterfaceSidebarValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"widget_namespace": schema.StringAttribute{
			Required: true,
		},
		"widget_id": schema.StringAttribute{
			Required: true,
		},
		"settings": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
		"disabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
	}
}
