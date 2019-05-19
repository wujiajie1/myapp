package sarama

import (
	"testing"
	"time"
	"vendor"
)

var aclDescribeResponseError = []byte{
	0, 0, 0, 100,
	0, 8, // error
	0, 5, 'e', 'r', 'r', 'o', 'r',
	0, 0, 0, 1, // 1 resource
	2, // cluster type
	0, 5, 't', 'o', 'p', 'i', 'c',
	0, 0, 0, 1, // 1 acl
	0, 9, 'p', 'r', 'i', 'n', 'c', 'i', 'p', 'a', 'l',
	0, 4, 'h', 'o', 's', 't',
	4, // write
	3, // allow
}

func TestAclDescribeResponse(t *testing.T) {
	errmsg := "error"
	resp := &vendor.DescribeAclsResponse{
		ThrottleTime: 100 * time.Millisecond,
		Err:          vendor.ErrBrokerNotAvailable,
		ErrMsg:       &errmsg,
		ResourceAcls: []*vendor.ResourceAcls{{
			Resource: vendor.Resource{
				ResourceName: "topic",
				ResourceType: vendor.AclResourceTopic,
			},
			Acls: []*vendor.Acl{
				{
					Principal:      "principal",
					Host:           "host",
					Operation:      vendor.AclOperationWrite,
					PermissionType: vendor.AclPermissionAllow,
				},
			},
		}},
	}

	vendor.testResponse(t, "describe", resp, aclDescribeResponseError)
}
