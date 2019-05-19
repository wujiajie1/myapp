package sarama

import "vendor"

type AclFilter struct {
	Version                   int
	ResourceType              vendor.AclResourceType
	ResourceName              *string
	ResourcePatternTypeFilter vendor.AclResourcePatternType
	Principal                 *string
	Host                      *string
	Operation                 vendor.AclOperation
	PermissionType            vendor.AclPermissionType
}

func (a *AclFilter) encode(pe vendor.packetEncoder) error {
	pe.putInt8(int8(a.ResourceType))
	if err := pe.putNullableString(a.ResourceName); err != nil {
		return err
	}

	if a.Version == 1 {
		pe.putInt8(int8(a.ResourcePatternTypeFilter))
	}

	if err := pe.putNullableString(a.Principal); err != nil {
		return err
	}
	if err := pe.putNullableString(a.Host); err != nil {
		return err
	}
	pe.putInt8(int8(a.Operation))
	pe.putInt8(int8(a.PermissionType))

	return nil
}

func (a *AclFilter) decode(pd vendor.packetDecoder, version int16) (err error) {
	resourceType, err := pd.getInt8()
	if err != nil {
		return err
	}
	a.ResourceType = vendor.AclResourceType(resourceType)

	if a.ResourceName, err = pd.getNullableString(); err != nil {
		return err
	}

	if a.Version == 1 {
		pattern, err := pd.getInt8()

		if err != nil {
			return err
		}

		a.ResourcePatternTypeFilter = vendor.AclResourcePatternType(pattern)
	}

	if a.Principal, err = pd.getNullableString(); err != nil {
		return err
	}

	if a.Host, err = pd.getNullableString(); err != nil {
		return err
	}

	operation, err := pd.getInt8()
	if err != nil {
		return err
	}
	a.Operation = vendor.AclOperation(operation)

	permissionType, err := pd.getInt8()
	if err != nil {
		return err
	}
	a.PermissionType = vendor.AclPermissionType(permissionType)

	return nil
}
