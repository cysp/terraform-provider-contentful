package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func EditorInterfaceResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
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
					Attributes: EditorInterfaceEditorLayoutValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceEditorLayoutValue{}.CustomType(ctx),
				},
				Optional: true,
			},
			"controls": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceControlValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceControlValue{}.CustomType(ctx),
				},
				Optional: true,
			},
			"group_controls": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceGroupControlValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceGroupControlValue{}.CustomType(ctx),
				},
				Optional: true,
			},
			"sidebar": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: EditorInterfaceSidebarValue{}.SchemaAttributes(ctx),
					CustomType: EditorInterfaceSidebarValue{}.CustomType(ctx),
				},
				Optional: true,
			},
		},
	}
}
