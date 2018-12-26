package logic

import (
	"context"

	"github.com/Juniper/contrail/pkg/errutil"
	"github.com/Juniper/contrail/pkg/models"
	"github.com/Juniper/contrail/pkg/models/basemodels"
	"github.com/twinj/uuid"

	"github.com/Juniper/contrail/pkg/services"
	"github.com/Juniper/contrail/pkg/services/baseservices"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

// Create logic
func (port *Port) Create(ctx context.Context, rp RequestParameters) (Response, error) {
	if len(port.ID) == 0 {
		port.ID = uuid.NewV4().String()
	}

	// if mac-address is specified, check against the exisitng ports
	// to see if there exists a port with the same mac-address

	if err := port.checkMacAddress(ctx, rp); err != nil {
		return nil, err
	}

	vn, err := port.getVirtualNetwork(ctx, rp)
	if err != nil {
		return nil, err
	}

	vmi, err := port.createVirtualMachineInterface(ctx, rp, vn)
	if err != nil {
		return nil, err
	}

	iip, err := port.allocateIPAddress(ctx, rp, vn, vmi)

	// TODO:
	// create interface route table for the port if
	// subnet has a host route for this port ip.

	return makePortResponse(vn, vmi, []*models.InstanceIP{iip})
}

// Update handles port update request
func (port *Port) Update(ctx context.Context, rp RequestParameters) (Response, error) {
	vmi, vn, err := port.readVNCPort(ctx, rp, port.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read vnc resources %s", port.ID)
	}

	fm := types.FieldMask{}

	if port.Name != "" {
		vmi.DisplayName = port.Name
		basemodels.FieldMaskAppend(&fm, models.VirtualMachineInterfaceFieldDisplayName)
	}

	if port.DeviceOwner != "network:router_interface" &&
		port.DeviceOwner != "network:router_gateway" && len(port.DeviceID) != 0 {
		if err := port.setVMInstance(ctx, rp, vmi); err != nil {
			return nil, err
		}
		basemodels.FieldMaskAppend(&fm, models.VirtualMachineInterfaceFieldVirtualMachineRefs)
	}
	//TODO handle key value pairs
	//TODO handle mac address change
	//TODO port security enabled update
	//TODO id perms update
	//TODO allowed_address_pairs update
	//TODO fixed_ips update

	if _, err = rp.WriteService.UpdateVirtualMachineInterface(ctx, &services.UpdateVirtualMachineInterfaceRequest{
		VirtualMachineInterface: vmi,
		FieldMask:               fm,
	}); err != nil {
		return nil, errors.Wrapf(err, "couldn't update vmi %s", vmi.GetUUID())
	}

	return makePortResponse(vn, vmi, vmi.GetInstanceIPBackRefs())
}

// Delete handles port delete requests
func (port *Port) Delete(ctx context.Context, rp RequestParameters, id string) (Response, error) {
	vmi, _, err := port.readVNCPort(ctx, rp, id)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read vnc resources %s", id)
	}

	if len(vmi.GetLogicalRouterBackRefs()) > 0 {
		return nil, errors.Errorf("L3PortInUse device_owner = network:router_interface port_id = %s", id)
	}

	// release instance IP address
	for _, iip := range vmi.GetInstanceIPBackRefs() {
		// TODO handle shared ip case
		if _, err := rp.WriteService.DeleteInstanceIP(ctx, &services.DeleteInstanceIPRequest{
			ID: iip.GetUUID(),
		}); err != nil {
			// instance ip could be deleted by svc monitor if it is
			// a shared ip. Ignore this error
		}
	}

	//TODO disassociate any floating IP used by instance

	if _, err := rp.WriteService.DeleteVirtualMachineInterface(ctx, &services.DeleteVirtualMachineInterfaceRequest{
		ID: vmi.GetUUID(),
	}); err != nil {
		return nil, errors.Wrapf(err, "couldn't delete virtual machine interface %s", vmi.GetUUID())
	}

	if vmID := port.getAsssociatedVirtualMachineID(vmi); vmID != "" {
		_, err = rp.WriteService.DeleteVirtualMachine(ctx, &services.DeleteVirtualMachineRequest{
			ID: vmID,
		})
		// delete instance if this was the last port
		if err != nil && !errutil.IsNotFound(err) && !errutil.IsConflict(err) {
			return nil, errors.Wrapf(err, "failed to delete virtual machine %s", vmID)
		}
	}

	//TODO delete any interface route table associated with the port to handle
	// subnet host route Neutron extension, un-reference others

	return &PortResponse{}, nil
}

// Read default implementation
func (port *Port) Read(ctx context.Context, rp RequestParameters, id string) (Response, error) {
	vmi, vn, err := port.readVNCPort(ctx, rp, id)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't read vnc resources %s", id)
	}

	return makePortResponse(vn, vmi, vmi.GetInstanceIPBackRefs())
}

// ReadAll logic
func (port *Port) ReadAll(ctx context.Context, rp RequestParameters, filters Filters, fields Fields) (Response, error) {
	// TODO implement ReadAll logic
	return []PortResponse{}, nil
}

func (port *Port) getAsssociatedVirtualMachineID(vmi *models.VirtualMachineInterface) string {
	if vmi.GetParentType() == models.KindVirtualMachine {
		return vmi.GetParentUUID()
	}

	if len(vmi.GetVirtualMachineRefs()) > 0 {
		return vmi.VirtualMachineRefs[0].GetUUID()
	}

	return ""
}

func (port *Port) getAssociatedVirtualNetwork(ctx context.Context, rp RequestParameters,
	vmi *models.VirtualMachineInterface,
) (*models.VirtualNetwork, error) {
	vnRefs := vmi.GetVirtualNetworkRefs()
	if len(vnRefs) == 0 {
		return nil, nil
	}
	vnRes, err := rp.ReadService.GetVirtualNetwork(ctx, &services.GetVirtualNetworkRequest{
		ID: vnRefs[0].GetUUID(),
	})

	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get virtual network %s", vnRefs[0].GetUUID())
	}
	return vnRes.GetVirtualNetwork(), nil
}

func (port *Port) readVNCPort(
	ctx context.Context, rp RequestParameters, id string,
) (*models.VirtualMachineInterface, *models.VirtualNetwork, error) {
	vmiRes, err := rp.ReadService.GetVirtualMachineInterface(ctx, &services.GetVirtualMachineInterfaceRequest{
		ID: id,
	})

	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't get virtual machine interface %s", id)
	}

	vmi := vmiRes.GetVirtualMachineInterface()
	vn, err := port.getAssociatedVirtualNetwork(ctx, rp, vmi)
	if err != nil {
		return nil, nil, err
	}
	return vmi, vn, nil
}

func (port *Port) allocateIPAddress(
	ctx context.Context, rp RequestParameters, vn *models.VirtualNetwork, vmi *models.VirtualMachineInterface,
) (*models.InstanceIP, error) {

	//TODO handle fixed_ips
	if len(vn.NetworkIpamRefs) == 0 {
		return nil, errors.Errorf("virtual network %v has no network ipam refs", vn.GetUUID())
	}

	return port.createInstanceIP(ctx, rp, vn, vmi, "", "")
}

func (port *Port) createInstanceIP(
	ctx context.Context, rp RequestParameters, vn *models.VirtualNetwork, vmi *models.VirtualMachineInterface,
	subnetUUID string, ipAddress string,
) (*models.InstanceIP, error) {

	ipUUID := "e3aaed67-be5b-4515-a624-b4a14e96aa08"
	iip := &models.InstanceIP{
		Name:              ipUUID,
		UUID:              ipUUID,
		SubnetUUID:        subnetUUID,
		InstanceIPAddress: ipAddress,
		Perms2: &models.PermType2{
			Owner: port.TenantID,
		},
	}

	iip.AddVirtualMachineInterfaceRef(&models.InstanceIPVirtualMachineInterfaceRef{
		UUID: vmi.UUID,
	})

	iip.AddVirtualNetworkRef(&models.InstanceIPVirtualNetworkRef{
		UUID: vn.UUID,
	})

	iipRes, err := rp.WriteService.CreateInstanceIP(ctx, &services.CreateInstanceIPRequest{
		InstanceIP: iip,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "couldn't create instance ip for port %v", port.Name)
	}

	return iipRes.InstanceIP, nil
}

func (port *Port) getVirtualNetwork(
	ctx context.Context, rp RequestParameters,
) (*models.VirtualNetwork, error) {
	res, err := rp.ReadService.GetVirtualNetwork(ctx, &services.GetVirtualNetworkRequest{
		ID: port.NetworkID,
	})
	return res.GetVirtualNetwork(), err
}

func (port *Port) getProjectID() string {
	uuid, err := uuid.Parse(port.TenantID)
	if err != nil {
		return ""
	}
	return uuid.String()
}

func (port *Port) ensureInstanceExists(
	ctx context.Context, rp RequestParameters,
) (*models.VirtualMachine, error) {
	vm := &models.VirtualMachine{
		Name: port.DeviceID,
		Perms2: &models.PermType2{
			Owner: port.getProjectID(),
		},
	}

	uuid, err := uuid.Parse(port.DeviceID)
	// if instance_id is not a valid uuid, let
	// virtual_machine_create generate uuid for the vm
	if err == nil {
		vm.UUID = uuid.String()
	}

	//TODO: Handle baremetal
	vm.ServerType = "virtual-server"

	createRes, err := rp.WriteService.CreateVirtualMachine(ctx, &services.CreateVirtualMachineRequest{
		VirtualMachine: vm,
	})

	if errutil.IsConflict(err) {
		// VM already exists try to read id
		readRes, err := rp.ReadService.GetVirtualMachine(ctx, &services.GetVirtualMachineRequest{
			ID: vm.GetUUID(),
		})

		if err != nil {
			return nil, errors.Wrapf(err, "couldn't get virtual machine uuid %s", vm.GetUUID())
		}
		//TODO: Handle baremetal
		vm = readRes.GetVirtualMachine()
	} else if err != nil {
		return nil, errors.Wrapf(err, "couldn't ensure vm instance (%s) existence", vm.GetUUID())
	} else {
		vm = createRes.GetVirtualMachine()
	}

	return vm, nil
}

func (port *Port) setVMInstance(ctx context.Context, rp RequestParameters,
	vmi *models.VirtualMachineInterface) error {
	//TODO: Delete old virtual machine object associated with the port

	if len(port.DeviceID) == 0 {
		vmi.VirtualMachineRefs = nil
		return nil
	}

	vm, err := port.ensureInstanceExists(ctx, rp)
	if err != nil {
		return err
	}

	vmi.AddVirtualMachineRef(&models.VirtualMachineInterfaceVirtualMachineRef{
		UUID: vm.GetUUID(),
		To:   vm.GetFQName(),
	})

	return nil
}

func (port *Port) setPortSecurity(
	ctx context.Context, rp RequestParameters, vmi *models.VirtualMachineInterface, vn *models.VirtualNetwork,
) error {
	vmi.PortSecurityEnabled = port.PortSecurityEnabled
	if !vmi.PortSecurityEnabled {
		vmi.PortSecurityEnabled = vn.PortSecurityEnabled
	}

	res, err := rp.ReadService.ListSecurityGroup(ctx, &services.ListSecurityGroupRequest{
		Spec: &baseservices.ListSpec{
			ObjectUUIDs: port.SecurityGroups,
			Fields:      []string{"uuid", "fqname"},
		},
	})

	if err != nil {
		return errors.Wrapf(err, "couldn't list security groups %v", port.SecurityGroups)
	}

	securityGroups := res.GetSecurityGroups()
	for _, sc := range securityGroups {
		vmi.AddSecurityGroupRef(&models.VirtualMachineInterfaceSecurityGroupRef{
			UUID: sc.GetUUID(),
		})
	}

	if len(vmi.SecurityGroupRefs) == 0 && vmi.PortSecurityEnabled {
		vmi.AddSecurityGroupRef(&models.VirtualMachineInterfaceSecurityGroupRef{
			To: []string{"default-domain", "default-project", "__no_rule__"},
		})
	}

	//TODO Handle default security group

	return nil
}

func (port *Port) createVirtualMachineInterface(
	ctx context.Context, rp RequestParameters, vn *models.VirtualNetwork,
) (*models.VirtualMachineInterface, error) {

	vmi := &models.VirtualMachineInterface{
		UUID:       port.ID,
		ParentType: models.KindProject,
		ParentUUID: port.getProjectID(),
		IDPerms: &models.IdPermsType{
			Enable: true,
		},
		Perms2: &models.PermType2{
			Owner: port.TenantID,
		},
	}

	if len(port.Name) == 0 {
		vmi.Name = port.ID
	} else {
		vmi.Name = port.Name
		vmi.DisplayName = port.Name
	}

	if len(port.MacAddress) != 0 {
		vmi.VirtualMachineInterfaceMacAddresses = &models.MacAddressesType{
			MacAddress: []string{port.MacAddress},
		}
	}

	vmi.AddVirtualNetworkRef(&models.VirtualMachineInterfaceVirtualNetworkRef{
		UUID: vn.GetUUID(),
	})

	if port.DeviceOwner != "network:router_interface" &&
		port.DeviceOwner != "network:router_gateway" && len(port.DeviceID) != 0 {
		if err := port.setVMInstance(ctx, rp, vmi); err != nil {
			return nil, err
		}
	}

	vmi.VirtualMachineInterfaceDeviceOwner = port.DeviceOwner
	if port.BindingVnicType != "" {
		kvps := &models.KeyValuePairs{}
		kvps.KeyValuePair = append(kvps.KeyValuePair, &models.KeyValuePair{
			Key:   "vnic_type",
			Value: port.BindingVnicType,
		})
		vmi.VirtualMachineInterfaceBindings = kvps
	}

	if err := port.setPortSecurity(ctx, rp, vmi, vn); err != nil {
		return nil, errors.Wrap(err, "couldn't setup port security")
	}

	//TODO Handle allowed address pair
	//TODO Handle fixed ips

	vmiRes, err := rp.WriteService.CreateVirtualMachineInterface(ctx, &services.CreateVirtualMachineInterfaceRequest{
		VirtualMachineInterface: vmi,
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create virtual-machine-interface")
	}

	return vmiRes.GetVirtualMachineInterface(), nil
}

func (port *Port) checkMacAddress(ctx context.Context, rp RequestParameters) error {
	if len(port.MacAddress) == 0 {
		return nil
	}

	res, err := rp.ReadService.ListVirtualMachineInterface(ctx, &services.ListVirtualMachineInterfaceRequest{
		Spec: &baseservices.ListSpec{
			Filters: []*baseservices.Filter{
				{
					Key:    "virtual_machine_interface_mac_addresses",
					Values: []string{port.MacAddress},
				},
			},
		},
	})

	if err != nil {
		return nil
	}

	if res.GetVirtualMachineInterfaceCount() != 0 {
		errors.Errorf("MacAddressInUse: mac_address = %s", port.MacAddress)
	}

	return nil
}
