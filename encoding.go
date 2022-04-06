package krakend

import (
	rss "github.com/devopsfaith/krakend-rss/v2"
	xml "github.com/devopsfaith/krakend-xml/v2"
	ginxml "github.com/devopsfaith/krakend-xml/v2/gin"
	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/proxy"
	routergin "github.com/luraproject/lura/v2/router/gin"
)

// RegisterEncoders registers all the available encoders
func RegisterEncoders() {
	xml.Register()
	rss.Register()

	routergin.RegisterRender(xml.Name, ginxml.Render)
	routergin.RegisterRender("custom-json", customJsonRender)
}

func customJsonRender(c *gin.Context, response *proxy.Response) {
	status := response.Metadata.StatusCode
	if response == nil {
		c.JSON(status, emptyResponse)
		return
	}
	c.JSON(status, response.Data)
}

var emptyResponse = gin.H{}
