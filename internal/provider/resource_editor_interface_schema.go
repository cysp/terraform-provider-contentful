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
		Version:     1,
		Description: "Manages a Contentful Editor Interface configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "The ID of the space this editor interface belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment this editor interface belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type_id": schema.StringAttribute{
				Description: "The ID of the content type this editor interface configures.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"editor_layout": schema.ListNestedAttribute{
				Description: "Layout configuration for the editor interface, defining how fields and groups are organized.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceEditorLayoutItemValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectNull[EditorInterfaceEditorLayoutItemValue]().CustomType(ctx),
				},
				CustomType: TypedList[TypedObject[EditorInterfaceEditorLayoutItemValue]]{}.CustomType(ctx),
				Optional:   true,
			},
			"controls": schema.ListNestedAttribute{
				Description: "Field-level controls that specify which widget to use for editing each field.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceControlValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectNull[EditorInterfaceControlValue]().CustomType(ctx),
				},
				CustomType: TypedList[TypedObject[EditorInterfaceControlValue]]{}.CustomType(ctx),
				Optional:   true,
			},
			"group_controls": schema.ListNestedAttribute{
				Description: "Group-level controls that specify widgets for field groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceGroupControlValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectNull[EditorInterfaceGroupControlValue]().CustomType(ctx),
				},
				CustomType: TypedList[TypedObject[EditorInterfaceGroupControlValue]]{}.CustomType(ctx),
				Optional:   true,
			},
			"sidebar": schema.ListNestedAttribute{
				Description: "Configuration for sidebar widgets in the editor.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceSidebarValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectNull[EditorInterfaceSidebarValue]().CustomType(ctx),
				},
				CustomType: TypedList[TypedObject[EditorInterfaceSidebarValue]]{}.CustomType(ctx),
				Optional:   true,
			},
		},
	}
}

func (v EditorInterfaceControlValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Description: "ID of the field this control configures.",
			Required:    true,
		},
		"widget_namespace": schema.StringAttribute{
			Description: "Namespace of the widget (e.g., 'builtin', 'extension', 'app').",
			Optional:    true,
		},
		"widget_id": schema.StringAttribute{
			Description: "ID of the widget to use for this field.",
			Optional:    true,
		},
		"settings": schema.StringAttribute{
			Description: "Widget-specific settings in JSON format.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemFieldValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Description: "ID of the field to include in this group item.",
			Required:    true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field_id": schema.StringAttribute{
			Description: "ID of the field to include in this nested group item.",
			Required:    true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field": schema.SingleNestedAttribute{
			Description: "Field configuration for this nested group item.",
			Attributes:  EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue]().CustomType(ctx),
			Required:    true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemGroupItemGroupValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Description: "ID of the nested group.",
			Required:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name of the nested group.",
			Required:    true,
		},
		"items": schema.ListNestedAttribute{
			Description: "Items within this nested group.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: EditorInterfaceEditorLayoutItemGroupItemGroupItemValue{}.SchemaAttributes(ctx),
				CustomType: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]().CustomType(ctx),
			},
			CustomType: NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]]().CustomType(ctx),
			Required:   true,
		},
	}
}

//nolint:dupl
func (v EditorInterfaceEditorLayoutItemGroupItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"field": schema.SingleNestedAttribute{
			Description: "Field to include in this group item.",
			Attributes:  EditorInterfaceEditorLayoutItemGroupItemFieldValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue]().CustomType(ctx),
			Optional:    true,
			Validators: []validator.Object{
				objectvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("field"),
					path.MatchRelative().AtParent().AtName("group"),
				),
			},
		},
		"group": schema.SingleNestedAttribute{
			Description: "Nested group within this group item.",
			Attributes:  EditorInterfaceEditorLayoutItemGroupItemGroupValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue]().CustomType(ctx),
			Optional:    true,
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
			Description: "ID of the layout group.",
			Required:    true,
		},
		"name": schema.StringAttribute{
			Description: "Name of the layout group.",
			Required:    true,
		},
		"items": schema.ListNestedAttribute{
			Description: "Items within this layout group.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: EditorInterfaceEditorLayoutItemGroupItemValue{}.SchemaAttributes(ctx),
				CustomType: NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue]().CustomType(ctx),
			},
			CustomType: NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]]().CustomType(ctx),
			Required:   true,
		},
	}
}

func (v EditorInterfaceEditorLayoutItemValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group": schema.SingleNestedAttribute{
			Description: "Group definition for this editor layout item.",
			Attributes:  EditorInterfaceEditorLayoutItemGroupValue{}.SchemaAttributes(ctx),
			CustomType:  NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue]().CustomType(ctx),
			Required:    true,
		},
	}
}

func (v EditorInterfaceGroupControlValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"group_id": schema.StringAttribute{
			Description: "ID of the group this control configures.",
			Required:    true,
		},
		"widget_namespace": schema.StringAttribute{
			Description: "Namespace of the widget.",
			Optional:    true,
		},
		"widget_id": schema.StringAttribute{
			Description: "ID of the widget to use for this group.",
			Optional:    true,
		},
		"settings": schema.StringAttribute{
			Description: "Widget-specific settings in JSON format.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
	}
}

func (v EditorInterfaceSidebarValue) SchemaAttributes(_ context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"widget_namespace": schema.StringAttribute{
			Description: "Namespace of the sidebar widget.",
			Required:    true,
		},
		"widget_id": schema.StringAttribute{
			Description: "ID of the sidebar widget.",
			Required:    true,
		},
		"settings": schema.StringAttribute{
			Description: "Widget-specific settings in JSON format.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
		"disabled": schema.BoolAttribute{
			Description: "Whether this sidebar widget is disabled.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
	}
}
