package sarama

import "vendor"

//Resource holds information about acl resource type
type Resource struct {
	ResourceType       vendor.AclResourceType
	ResourceName       string
	ResoucePatternType vendor.AclResourcePatternType
}

func (r *Resource) encode(pe vendor.packetEncoder, version int16) error {
	pe.putInt8(int8(r.ResourceType))

	if err := pe.putString(r.ResourceName); err != nil {
		return err
	}

	if version == 1 {
		if r.ResoucePatternType == vendor.AclPatternUnknown {
			vendor.Logger.Print("Cannot encode an unknown resource pattern type, using Literal instead")
			r.ResoucePatternType = vendor.AclPatternLiteral
		}
		pe.putInt8(int8(r.ResoucePatternType))
	}

	return nil
}

func (r *Resource) decode(pd vendor.packetDecoder, version int16) (err error) {
	resourceType, err := pd.getInt8()
	if err != nil {
		return err
	}
	r.ResourceType = vendor.AclResourceType(resourceType)

	if r.ResourceName, err = pd.getString(); err != nil {
		return err
	}
	if version == 1 {
		pattern, err := pd.getInt8()
		if err != nil {
			return err
		}
		r.ResoucePatternType = vendor.AclResourcePatternType(pattern)
	}

	return nil
}

//Acl holds information about acl type
type Acl struct {
	Principal      string
	Host           string
	Operation      vendor.AclOperation
	PermissionType vendor.AclPermissionType
}

func (a *Acl) encode(pe vendor.packetEncoder) error {
	if err := pe.putString(a.Principal); err != nil {
		return err
	}

	if err := pe.putString(a.Host); err != nil {
		return err
	}

	pe.putInt8(int8(a.Operation))
	pe.putInt8(int8(a.PermissionType))

	return nil
}

func (a *Acl) decode(pd vendor.packetDecoder, version int16) (err error) {
	if a.Principal, err = pd.getString(); err != nil {
		return err
	}

	if a.Host, err = pd.getString(); err != nil {
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

//ResourceAcls is an acl resource type
type ResourceAcls struct {
	Resource
	Acls []*Acl
}

func (r *ResourceAcls) encode(pe vendor.packetEncoder, version int16) error {
	if err := r.Resource.encode(pe, version); err != nil {
		return err
	}

	if err := pe.putArrayLength(len(r.Acls)); err != nil {
		return err
	}
	for _, acl := range r.Acls {
		if err := acl.encode(pe); err != nil {
			return err
		}
	}

	return nil
}

func (r *ResourceAcls) decode(pd vendor.packetDecoder, version int16) error {
	if err := r.Resource.decode(pd, version); err != nil {
		return err
	}

	n, err := pd.getArrayLength()
	if err != nil {
		return err
	}

	r.Acls = make([]*Acl, n)
	for i := 0; i < n; i++ {
		r.Acls[i] = new(Acl)
		if err := r.Acls[i].decode(pd, version); err != nil {
			return err
		}
	}

	return nil
}
