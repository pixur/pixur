package schema

import (
	wpb "github.com/golang/protobuf/ptypes/wrappers"
)

var DefaultConfiguration = Configuration{
	MinCommentLength: &wpb.Int64Value{
		Value: 1,
	},
	MaxCommentLength: &wpb.Int64Value{
		Value: 16384,
	},
	MinIdentLength: &wpb.Int64Value{
		Value: 1,
	},
	MaxIdentLength: &wpb.Int64Value{
		Value: 128,
	},
	MinFileNameLength: &wpb.Int64Value{
		Value: int64(len("a.a")),
	},
	MaxFileNameLength: &wpb.Int64Value{
		Value: 255,
	},
	MinUrlLength: &wpb.Int64Value{
		Value: int64(len("http://")),
	},
	MaxUrlLength: &wpb.Int64Value{
		Value: 2000,
	},
	MinTagLength: &wpb.Int64Value{
		Value: 1,
	},
	MaxTagLength: &wpb.Int64Value{
		Value: 64,
	},
	AnonymousCapability: &Configuration_CapabilitySet{
		Capability: []User_Capability{
			User_USER_CREATE,
		},
	},
	NewUserCapability: &Configuration_CapabilitySet{
		Capability: []User_Capability{
			User_PIC_READ,
			User_PIC_INDEX,
			User_PIC_UPDATE_VIEW_COUNTER,
			User_PIC_TAG_CREATE,
			User_PIC_COMMENT_CREATE,
			User_PIC_VOTE_CREATE,
			User_USER_READ_SELF,
		},
	},
}
