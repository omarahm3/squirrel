package common

import (
	"html/template"

	"github.com/gin-gonic/gin"
)

func LoadHtmlTemplates(server *gin.Engine, templates map[string]string) error {
	for _, strTemplate := range templates {
		tpl, err := template.New("index.html").Parse(strTemplate)

		if err != nil {
			return err
		}

		server.SetHTMLTemplate(tpl)
	}

	return nil
}
