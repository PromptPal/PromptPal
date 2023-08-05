package schema

import "github.com/PromptPal/PromptPal/ent"

type userResponse struct {
	u *ent.User
}

func (u userResponse) ID() int32 {
	return int32(u.u.ID)
}

func (u userResponse) Name() string {
	return u.u.Name
}

func (u userResponse) Addr() string {
	return u.u.Addr
}
