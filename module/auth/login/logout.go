package login

import (
	"github.com/0987363/aproxy/lib/rfweb"
)

type LogoutResource struct {
	rfweb.BaseResource
}

func (self *LogoutResource) Get(ctx *rfweb.Context) {
	session := ctx.Session()
	session.Clear(ctx.W)
	redirectToLogin(ctx.W, ctx.R, false)
}
