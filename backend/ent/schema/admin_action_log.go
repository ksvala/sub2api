package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AdminActionLog holds the schema definition for the AdminActionLog entity.
type AdminActionLog struct {
	ent.Schema
}

func (AdminActionLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "admin_action_logs"},
	}
}

func (AdminActionLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("admin_id").
			Optional().
			Nillable(),
		field.String("action").
			MaxLen(64),
		field.String("resource_type").
			MaxLen(64),
		field.Int64("resource_id").
			Optional().
			Nillable(),
		field.String("payload").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("ip_address").
			Optional().
			Nillable().
			MaxLen(64),
		field.String("user_agent").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (AdminActionLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("admin", User.Type).
			Ref("admin_action_logs").
			Field("admin_id").
			Unique(),
	}
}

func (AdminActionLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("admin_id"),
		index.Fields("resource_type"),
		index.Fields("created_at"),
	}
}
