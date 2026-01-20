package schema

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Invitation holds the schema definition for the Invitation entity.
type Invitation struct {
	ent.Schema
}

func (Invitation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_invites"},
	}
}

func (Invitation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("inviter_id"),
		field.Int64("invitee_id"),
		field.String("invite_code").
			MaxLen(6),
		field.Float("reward_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.String("status").
			MaxLen(20).
			Default(service.InviteStatusPending),
		field.Int64("confirmed_by").
			Optional().
			Nillable(),
		field.Time("confirmed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("created_at").
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (Invitation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("inviter", User.Type).
			Ref("sent_invitations").
			Field("inviter_id").
			Required().
			Unique(),
		edge.From("invitee", User.Type).
			Ref("received_invitation").
			Field("invitee_id").
			Required().
			Unique(),
		edge.From("confirmed_by_user", User.Type).
			Ref("confirmed_invites").
			Field("confirmed_by").
			Unique(),
		edge.To("logs", InviteLog.Type),
	}
}

func (Invitation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("invitee_id").
			Unique(),
		index.Fields("inviter_id"),
		index.Fields("status"),
		index.Fields("created_at"),
	}
}
