package services

import (
	"context"
	"github.com/Juniper/contrail/pkg/common"
	"github.com/Juniper/contrail/pkg/generated/models"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"net/http"

	log "github.com/sirupsen/logrus"
)

//RESTCreateServiceInstance handle a Create REST service.
func (service *ContrailService) RESTCreateServiceInstance(c echo.Context) error {
	requestData := &models.CreateServiceInstanceRequest{}
	if err := c.Bind(requestData); err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"resource": "service_instance",
		}).Debug("bind failed on create")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}
	ctx := c.Request().Context()
	response, err := service.CreateServiceInstance(ctx, requestData)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusCreated, response)
}

//CreateServiceInstance handle a Create API
func (service *ContrailService) CreateServiceInstance(
	ctx context.Context,
	request *models.CreateServiceInstanceRequest) (*models.CreateServiceInstanceResponse, error) {
	model := request.ServiceInstance
	if model.UUID == "" {
		model.UUID = uuid.NewV4().String()
	}
	auth := common.GetAuthCTX(ctx)
	if auth == nil {
		return nil, common.ErrorUnauthenticated
	}

	if model.FQName == nil {
		if model.DisplayName != "" {
			model.FQName = []string{auth.DomainID(), auth.ProjectID(), model.DisplayName}
		} else {
			model.FQName = []string{auth.DomainID(), auth.ProjectID(), model.UUID}
		}
	}
	model.Perms2 = &models.PermType2{}
	model.Perms2.Owner = auth.ProjectID()

	return service.Next().CreateServiceInstance(ctx, request)
}

//RESTUpdateServiceInstance handles a REST Update request.
func (service *ContrailService) RESTUpdateServiceInstance(c echo.Context) error {
	//id := c.Param("id")
	request := &models.UpdateServiceInstanceRequest{}
	if err := c.Bind(request); err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"resource": "service_instance",
		}).Debug("bind failed on update")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}
	ctx := c.Request().Context()
	response, err := service.UpdateServiceInstance(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}

//UpdateServiceInstance handles a Update request.
func (service *ContrailService) UpdateServiceInstance(
	ctx context.Context,
	request *models.UpdateServiceInstanceRequest) (*models.UpdateServiceInstanceResponse, error) {
	model := request.ServiceInstance
	if model == nil {
		return nil, common.ErrorBadRequest("Update body is empty")
	}
	return service.Next().UpdateServiceInstance(ctx, request)
}

//RESTDeleteServiceInstance delete a resource using REST service.
func (service *ContrailService) RESTDeleteServiceInstance(c echo.Context) error {
	id := c.Param("id")
	request := &models.DeleteServiceInstanceRequest{
		ID: id,
	}
	ctx := c.Request().Context()
	_, err := service.DeleteServiceInstance(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

//RESTGetServiceInstance a REST Get request.
func (service *ContrailService) RESTGetServiceInstance(c echo.Context) error {
	id := c.Param("id")
	request := &models.GetServiceInstanceRequest{
		ID: id,
	}
	ctx := c.Request().Context()
	response, err := service.GetServiceInstance(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}

//RESTListServiceInstance handles a List REST service Request.
func (service *ContrailService) RESTListServiceInstance(c echo.Context) error {
	var err error
	spec := common.GetListSpec(c)
	request := &models.ListServiceInstanceRequest{
		Spec: spec,
	}
	ctx := c.Request().Context()
	response, err := service.ListServiceInstance(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}
