package handler

import (
	"log/slog"
	"net/http"

	"github.com/alanwade2001/go-sepa-engine-data/model"
	"github.com/alanwade2001/go-sepa-engine-settle/internal/service"
	"github.com/alanwade2001/go-sepa-infra/routing"

	"github.com/gin-gonic/gin"
)

type Execution struct {
	service *service.Execution
}

func NewExecution(service *service.Execution, r *routing.Router) *Execution {
	execution := &Execution{
		service: service,
	}

	r.Router.POST("/executions", execution.PostExecution)

	return execution
}

// postInitiation adds an initiations from JSON received in the request body.
func (d *Execution) PostExecution(c *gin.Context) {

	execution := &model.Execution{}
	if err := c.BindJSON(execution); err != nil {
		slog.Error("failed to bind body", "Error", err)
		c.IndentedJSON(http.StatusInternalServerError, err)
	} else {
		slog.Debug("executing settlement")
		if newExecution, err := d.service.Execute(execution); err != nil {
			slog.Error("failed to post execution", "Error", err)
			c.IndentedJSON(http.StatusInternalServerError, newExecution)
		} else {
			c.IndentedJSON(http.StatusCreated, newExecution)
		}
	}
}
