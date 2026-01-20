package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Plan holds the schema definition for the Plan entity.
// 套餐配置（分组、价格、额度、购买二维码等）
type Plan struct {
	ent.Schema
}

func (Plan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "plans"},
	}
}

func (Plan) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			MaxLen(120).
			NotEmpty().
			Comment("套餐标题"),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("套餐描述"),
		field.Float("price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("套餐价格（CNY）"),
		field.String("group_name").
			MaxLen(80).
			Default("default").
			Comment("分组名称"),
		field.Int("group_sort").
			Default(0).
			Comment("分组排序"),
		field.Float("daily_quota").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("每日额度"),
		field.Float("total_quota").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("总额度"),
		field.String("purchase_qr_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("购买二维码URL"),
		field.Bool("enabled").
			Default(true).
			Comment("是否启用"),
		field.Int("sort_order").
			Default(0).
			Comment("排序"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (Plan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("group_name"),
		index.Fields("group_sort"),
		index.Fields("enabled"),
		index.Fields("sort_order"),
	}
}
