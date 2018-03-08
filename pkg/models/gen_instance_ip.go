package models

import (
	"github.com/Juniper/contrail/pkg/common"
)

//To skip import error.
var _ = common.OPERATION

// MakeInstanceIP makes InstanceIP
// nolint
func MakeInstanceIP() *InstanceIP {
	return &InstanceIP{
		//TODO(nati): Apply default
		UUID:                  "",
		ParentUUID:            "",
		ParentType:            "",
		FQName:                []string{},
		IDPerms:               MakeIdPermsType(),
		DisplayName:           "",
		Annotations:           MakeKeyValuePairs(),
		Perms2:                MakePermType2(),
		ConfigurationVersion:  0,
		ServiceHealthCheckIP:  false,
		SecondaryIPTrackingIP: MakeSubnetType(),
		InstanceIPAddress:     "",
		InstanceIPMode:        "",
		SubnetUUID:            "",
		InstanceIPFamily:      "",
		ServiceInstanceIP:     false,
		InstanceIPLocalIP:     false,
		InstanceIPSecondary:   false,
	}
}

// MakeInstanceIP makes InstanceIP
// nolint
func InterfaceToInstanceIP(i interface{}) *InstanceIP {
	m, ok := i.(map[string]interface{})
	_ = m
	if !ok {
		return nil
	}
	return &InstanceIP{
		//TODO(nati): Apply default
		UUID:                  common.InterfaceToString(m["uuid"]),
		ParentUUID:            common.InterfaceToString(m["parent_uuid"]),
		ParentType:            common.InterfaceToString(m["parent_type"]),
		FQName:                common.InterfaceToStringList(m["fq_name"]),
		IDPerms:               InterfaceToIdPermsType(m["id_perms"]),
		DisplayName:           common.InterfaceToString(m["display_name"]),
		Annotations:           InterfaceToKeyValuePairs(m["annotations"]),
		Perms2:                InterfaceToPermType2(m["perms2"]),
		ConfigurationVersion:  common.InterfaceToInt64(m["configuration_version"]),
		ServiceHealthCheckIP:  common.InterfaceToBool(m["service_health_check_ip"]),
		SecondaryIPTrackingIP: InterfaceToSubnetType(m["secondary_ip_tracking_ip"]),
		InstanceIPAddress:     common.InterfaceToString(m["instance_ip_address"]),
		InstanceIPMode:        common.InterfaceToString(m["instance_ip_mode"]),
		SubnetUUID:            common.InterfaceToString(m["subnet_uuid"]),
		InstanceIPFamily:      common.InterfaceToString(m["instance_ip_family"]),
		ServiceInstanceIP:     common.InterfaceToBool(m["service_instance_ip"]),
		InstanceIPLocalIP:     common.InterfaceToBool(m["instance_ip_local_ip"]),
		InstanceIPSecondary:   common.InterfaceToBool(m["instance_ip_secondary"]),

		VirtualMachineInterfaceRefs: InterfaceToInstanceIPVirtualMachineInterfaceRefs(m["virtual_machine_interface_refs"]),

		PhysicalRouterRefs: InterfaceToInstanceIPPhysicalRouterRefs(m["physical_router_refs"]),

		VirtualRouterRefs: InterfaceToInstanceIPVirtualRouterRefs(m["virtual_router_refs"]),

		NetworkIpamRefs: InterfaceToInstanceIPNetworkIpamRefs(m["network_ipam_refs"]),

		VirtualNetworkRefs: InterfaceToInstanceIPVirtualNetworkRefs(m["virtual_network_refs"]),
	}
}

func InterfaceToInstanceIPNetworkIpamRefs(i interface{}) []*InstanceIPNetworkIpamRef {
	list, ok := i.([]interface{})
	if !ok {
		return nil
	}
	result := []*InstanceIPNetworkIpamRef{}
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		_ = m
		if !ok {
			return nil
		}
		result = append(result, &InstanceIPNetworkIpamRef{
			UUID: common.InterfaceToString(m["uuid"]),
			To:   common.InterfaceToStringList(m["to"]),
		})
	}

	return result
}

func InterfaceToInstanceIPVirtualNetworkRefs(i interface{}) []*InstanceIPVirtualNetworkRef {
	list, ok := i.([]interface{})
	if !ok {
		return nil
	}
	result := []*InstanceIPVirtualNetworkRef{}
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		_ = m
		if !ok {
			return nil
		}
		result = append(result, &InstanceIPVirtualNetworkRef{
			UUID: common.InterfaceToString(m["uuid"]),
			To:   common.InterfaceToStringList(m["to"]),
		})
	}

	return result
}

func InterfaceToInstanceIPVirtualMachineInterfaceRefs(i interface{}) []*InstanceIPVirtualMachineInterfaceRef {
	list, ok := i.([]interface{})
	if !ok {
		return nil
	}
	result := []*InstanceIPVirtualMachineInterfaceRef{}
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		_ = m
		if !ok {
			return nil
		}
		result = append(result, &InstanceIPVirtualMachineInterfaceRef{
			UUID: common.InterfaceToString(m["uuid"]),
			To:   common.InterfaceToStringList(m["to"]),
		})
	}

	return result
}

func InterfaceToInstanceIPPhysicalRouterRefs(i interface{}) []*InstanceIPPhysicalRouterRef {
	list, ok := i.([]interface{})
	if !ok {
		return nil
	}
	result := []*InstanceIPPhysicalRouterRef{}
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		_ = m
		if !ok {
			return nil
		}
		result = append(result, &InstanceIPPhysicalRouterRef{
			UUID: common.InterfaceToString(m["uuid"]),
			To:   common.InterfaceToStringList(m["to"]),
		})
	}

	return result
}

func InterfaceToInstanceIPVirtualRouterRefs(i interface{}) []*InstanceIPVirtualRouterRef {
	list, ok := i.([]interface{})
	if !ok {
		return nil
	}
	result := []*InstanceIPVirtualRouterRef{}
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		_ = m
		if !ok {
			return nil
		}
		result = append(result, &InstanceIPVirtualRouterRef{
			UUID: common.InterfaceToString(m["uuid"]),
			To:   common.InterfaceToStringList(m["to"]),
		})
	}

	return result
}

// MakeInstanceIPSlice() makes a slice of InstanceIP
// nolint
func MakeInstanceIPSlice() []*InstanceIP {
	return []*InstanceIP{}
}

// InterfaceToInstanceIPSlice() makes a slice of InstanceIP
// nolint
func InterfaceToInstanceIPSlice(i interface{}) []*InstanceIP {
	list := common.InterfaceToInterfaceList(i)
	if list == nil {
		return nil
	}
	result := []*InstanceIP{}
	for _, item := range list {
		result = append(result, InterfaceToInstanceIP(item))
	}
	return result
}
