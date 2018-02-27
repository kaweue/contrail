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

//RESTCreatePolicyManagement handle a Create REST service.
func (service *ContrailService) RESTCreatePolicyManagement(c echo.Context) error {
	requestData := &models.CreatePolicyManagementRequest{}
	if err := c.Bind(requestData); err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"resource": "policy_management",
		}).Debug("bind failed on create")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}
	ctx := c.Request().Context()
	response, err := service.CreatePolicyManagement(ctx, requestData)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusCreated, response)
}

//CreatePolicyManagement handle a Create API
func (service *ContrailService) CreatePolicyManagement(
	ctx context.Context,
	request *models.CreatePolicyManagementRequest) (*models.CreatePolicyManagementResponse, error) {
	model := request.PolicyManagement
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

	return service.Next().CreatePolicyManagement(ctx, request)
}

//RESTUpdatePolicyManagement handles a REST Update request.
func (service *ContrailService) RESTUpdatePolicyManagement(c echo.Context) error {
	//id := c.Param("id")
	request := &models.UpdatePolicyManagementRequest{}
	if err := c.Bind(request); err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"resource": "policy_management",
		}).Debug("bind failed on update")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON format")
	}
	ctx := c.Request().Context()
	response, err := service.UpdatePolicyManagement(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}

//UpdatePolicyManagement handles a Update request.
func (service *ContrailService) UpdatePolicyManagement(
	ctx context.Context,
	request *models.UpdatePolicyManagementRequest) (*models.UpdatePolicyManagementResponse, error) {
	model := request.PolicyManagement
	if model == nil {
		return nil, common.ErrorBadRequest("Update body is empty")
	}
	return service.Next().UpdatePolicyManagement(ctx, request)
}

//RESTDeletePolicyManagement delete a resource using REST service.
func (service *ContrailService) RESTDeletePolicyManagement(c echo.Context) error {
	id := c.Param("id")
	request := &models.DeletePolicyManagementRequest{
		ID: id,
	}
	ctx := c.Request().Context()
	_, err := service.DeletePolicyManagement(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusNoContent, nil)
}

//RESTGetPolicyManagement a REST Get request.
func (service *ContrailService) RESTGetPolicyManagement(c echo.Context) error {
	id := c.Param("id")
	request := &models.GetPolicyManagementRequest{
		ID: id,
	}
	ctx := c.Request().Context()
	response, err := service.GetPolicyManagement(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}

//RESTListPolicyManagement handles a List REST service Request.
func (service *ContrailService) RESTListPolicyManagement(c echo.Context) error {
	var err error
	spec := common.GetListSpec(c)
	request := &models.ListPolicyManagementRequest{
		Spec: spec,
	}
	ctx := c.Request().Context()
	response, err := service.ListPolicyManagement(ctx, request)
	if err != nil {
		return common.ToHTTPError(err)
	}
	return c.JSON(http.StatusOK, response)
}
