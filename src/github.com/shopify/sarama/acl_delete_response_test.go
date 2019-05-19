package sarama

import (
	"testing"
	"time"
	"vendor"
)

var (
	deleteAclsResponse = []byte{
		0, 0, 0, 100,
		0, 0, 0, 1,
		0, 0, // no error
		255, 255, // no error message
		0, 0, 0, 1, // 1 matching acl
		0, 0, // no error
		255, 255, // no error message
		2, // resource type
		0, 5, 't', 'o', 'p', 'i', 'c',
		0, 9, 'p', 'r', 'i', 'n', 'c', 'i', 'p', 'a', 'l',
		0, 4, 'h', 'o', 's', 't',
		4,
		3,
	}
)

func TestDeleteAclsResponse(t *testing.T) {
	resp := &vendor.DeleteAclsResponse{
		ThrottleTime: 100 * time.Millisecond,
		FilterResponses: []*vendor.FilterResponse{{
			MatchingAcls: []*vendor.MatchingAcl{{
				Resource: vendor.Resource{ResourceType: vendor.AclResourceTopic, ResourceName: "topic"},
				Acl:      vendor.Acl{Principal: "principal", Host: "host", Operation: vendor.AclOperationWrite, PermissionType: vendor.AclPermissionAllow},
			}},
		}},
	}

	vendor.testResponse(t, "", resp, deleteAclsResponse)
}
