package sarama

import (
	"testing"
	"vendor"
)

var (
	aclDeleteRequestNullsv1 = []byte{
		0, 0, 0, 1,
		1,
		255, 255,
		1, // Any
		255, 255,
		255, 255,
		11,
		3,
	}

	aclDeleteRequestv1 = []byte{
		0, 0, 0, 1,
		1, // any
		0, 6, 'f', 'i', 'l', 't', 'e', 'r',
		1, // Any Filter
		0, 9, 'p', 'r', 'i', 'n', 'c', 'i', 'p', 'a', 'l',
		0, 4, 'h', 'o', 's', 't',
		4, // write
		3, // allow
	}

	aclDeleteRequestNulls = []byte{
		0, 0, 0, 1,
		1,
		255, 255,
		255, 255,
		255, 255,
		11,
		3,
	}

	aclDeleteRequest = []byte{
		0, 0, 0, 1,
		1, // any
		0, 6, 'f', 'i', 'l', 't', 'e', 'r',
		0, 9, 'p', 'r', 'i', 'n', 'c', 'i', 'p', 'a', 'l',
		0, 4, 'h', 'o', 's', 't',
		4, // write
		3, // allow
	}

	aclDeleteRequestArray = []byte{
		0, 0, 0, 2,
		1,
		0, 6, 'f', 'i', 'l', 't', 'e', 'r',
		0, 9, 'p', 'r', 'i', 'n', 'c', 'i', 'p', 'a', 'l',
		0, 4, 'h', 'o', 's', 't',
		4, // write
		3, // allow
		2,
		0, 5, 't', 'o', 'p', 'i', 'c',
		255, 255,
		255, 255,
		6,
		2,
	}
)

func TestDeleteAclsRequest(t *testing.T) {
	req := &vendor.DeleteAclsRequest{
		Filters: []*vendor.AclFilter{{
			ResourceType:   vendor.AclResourceAny,
			Operation:      vendor.AclOperationAlterConfigs,
			PermissionType: vendor.AclPermissionAllow,
		}},
	}

	vendor.testRequest(t, "delete request nulls", req, aclDeleteRequestNulls)

	req.Filters[0].ResourceName = vendor.nullString("filter")
	req.Filters[0].Principal = vendor.nullString("principal")
	req.Filters[0].Host = vendor.nullString("host")
	req.Filters[0].Operation = vendor.AclOperationWrite

	vendor.testRequest(t, "delete request", req, aclDeleteRequest)

	req.Filters = append(req.Filters, &vendor.AclFilter{
		ResourceType:   vendor.AclResourceTopic,
		ResourceName:   vendor.nullString("topic"),
		Operation:      vendor.AclOperationDelete,
		PermissionType: vendor.AclPermissionDeny,
	})

	vendor.testRequest(t, "delete request array", req, aclDeleteRequestArray)
}

func TestDeleteAclsRequestV1(t *testing.T) {
	req := &vendor.DeleteAclsRequest{
		Version: 1,
		Filters: []*vendor.AclFilter{{
			ResourceType:              vendor.AclResourceAny,
			Operation:                 vendor.AclOperationAlterConfigs,
			PermissionType:            vendor.AclPermissionAllow,
			ResourcePatternTypeFilter: vendor.AclPatternAny,
		}},
	}

	vendor.testRequest(t, "delete request nulls", req, aclDeleteRequestNullsv1)

	req.Filters[0].ResourceName = vendor.nullString("filter")
	req.Filters[0].Principal = vendor.nullString("principal")
	req.Filters[0].Host = vendor.nullString("host")
	req.Filters[0].Operation = vendor.AclOperationWrite

	vendor.testRequest(t, "delete request", req, aclDeleteRequestv1)
}
