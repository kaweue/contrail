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

//RESTCreateAddressGroup handle a Create REST service.
func (service *ContrailService) RESTCreateAddressGroup(c echo.Context) error {
	requestData := &models.CreateAddressGroupRequest{}
	if err := c.Bind(requestData); err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"resource": "address_group",
		}).Debug("bind failed on create")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}
	ctx := c.Request().Context()
	response, err := service.CreateAddressGroup(ctx, requestData)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusCreated, response)
}

//CreateAddressGroup handle a Create API
func (service *ContrailService) CreateAddressGroup(
	ctx context.Context,
	request *models.CreateAddressGroupRequest) (*models.CreateAddressGroupResponse, error) {
	model := request.AddressGroup
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

	return service.Next().CreateAddressGroup(ctx, request)
}

//RESTUpdateAddressGroup handles a REST Update request.
func (service *ContrailService) RESTUpdateAddressGroup(c echo.Context) error {
	//id := c.Param("id")
	request := &models.UpdateAddressGroupRequest{}
	if err := c.Bind(request); err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"resource": "address_group",
		}).Debug("bind failed on update")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}
	ctx := c.Request().Context()
	response, err := service.UpdateAddressGroup(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}

//UpdateAddressGroup handles a Update request.
func (service *ContrailService) UpdateAddressGroup(
	ctx context.Context,
	request *models.UpdateAddressGroupRequest) (*models.UpdateAddressGroupResponse, error) {
	model := request.AddressGroup
	if model == nil {
		return nil, common.ErrorBadRequest("Update body is empty")
	}
	return service.Next().UpdateAddressGroup(ctx, request)
}

//RESTDeleteAddressGroup delete a resource using REST service.
func (service *ContrailService) RESTDeleteAddressGroup(c echo.Context) error {
	id := c.Param("id")
	request := &models.DeleteAddressGroupRequest{
		ID: id,
	}
	ctx := c.Request().Context()
	_, err := service.DeleteAddressGroup(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

//RESTGetAddressGroup a REST Get request.
func (service *ContrailService) RESTGetAddressGroup(c echo.Context) error {
	id := c.Param("id")
	request := &models.GetAddressGroupRequest{
		ID: id,
	}
	ctx := c.Request().Context()
	response, err := service.GetAddressGroup(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}

//RESTListAddressGroup handles a List REST service Request.
func (service *ContrailService) RESTListAddressGroup(c echo.Context) error {
	var err error
	spec := common.GetListSpec(c)
	request := &models.ListAddressGroupRequest{
		Spec: spec,
	}
	ctx := c.Request().Context()
	response, err := service.ListAddressGroup(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}
