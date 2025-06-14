package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EditorInterfaceModel struct {
	ID            types.String                                    `tfsdk:"id"`
	SpaceID       types.String                                    `tfsdk:"space_id"`
	EnvironmentID types.String                                    `tfsdk:"environment_id"`
	ContentTypeID types.String                                    `tfsdk:"content_type_id"`
	EditorLayout  TypedList[EditorInterfaceEditorLayoutItemValue] `tfsdk:"editor_layout"`
	Controls      TypedList[EditorInterfaceControlValue]          `tfsdk:"controls"`
	GroupControls TypedList[EditorInterfaceGroupControlValue]     `tfsdk:"group_controls"`
	Sidebar       TypedList[EditorInterfaceSidebarValue]          `tfsdk:"sidebar"`
}

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
