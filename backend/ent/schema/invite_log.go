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

// InviteLog holds the schema definition for the InviteLog entity.
type InviteLog struct {
	ent.Schema
}

func (InviteLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "invite_logs"},
	}
}

func (InviteLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("invite_id"),
		field.String("action").
			MaxLen(20),
		field.Int64("inviter_id"),
		field.Int64("invitee_id"),
		field.Int64("admin_id").
			Optional().
			Nillable(),
		field.Float("reward_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (InviteLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("invite", Invitation.Type).
			Ref("logs").
			Field("invite_id").
			Required().
			Unique(),
		edge.From("inviter", User.Type).
			Ref("invite_logs_as_inviter").
			Field("inviter_id").
			Required().
			Unique(),
		edge.From("invitee", User.Type).
			Ref("invite_logs_as_invitee").
			Field("invitee_id").
			Required().
			Unique(),
		edge.From("admin", User.Type).
			Ref("invite_logs_as_admin").
			Field("admin_id").
			Unique(),
	}
}

func (InviteLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action"),
		index.Fields("created_at"),
		index.Fields("inviter_id"),
		index.Fields("invitee_id"),
		index.Fields("invite_id"),
	}
}
