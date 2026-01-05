package Domain

type Currency string

const (
	CurrencyINR Currency = "INR"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

func IsValidCurrency(c string) bool {
	switch Currency(c) {
	case CurrencyINR, CurrencyUSD, CurrencyEUR:
		return true
	}
	return false
}

type SplitType string

const (
	SplitTypeGroup  SplitType = "GROUP"
	SplitTypeDirect SplitType = "DIRECT"
)

func IsValidSplitType(t string) bool {
	switch SplitType(t) {
	case SplitTypeGroup, SplitTypeDirect:
		return true
	}
	return false
}

type SplitDivisionType string

const (
	SplitDivisionEqual  SplitDivisionType = "EQUAL"
	SplitDivisionCustom SplitDivisionType = "CUSTOM"
)

func IsValidSplitDivisionType(d string) bool {
	switch SplitDivisionType(d) {
	case SplitDivisionEqual, SplitDivisionCustom:
		return true
	}
	return false
}

type GroupRole string

const (
	GroupRoleOwner  GroupRole = "OWNER"
	GroupRoleAdmin  GroupRole = "ADMIN"
	GroupRoleMember GroupRole = "MEMBER"
)

func IsValidGroupRole(r string) bool {
	switch GroupRole(r) {
	case GroupRoleOwner, GroupRoleAdmin, GroupRoleMember:
		return true
	}
	return false
}

func IsValidAssignableRole(r string) bool {
	switch GroupRole(r) {
	case GroupRoleAdmin, GroupRoleMember:
		return true
	}
	return false
}
