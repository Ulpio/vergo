package billing

type PlanLimits struct {
	MaxProjects int `json:"max_projects"`
	MaxMembers  int `json:"max_members"`
	MaxStorageMB int `json:"max_storage_mb"`
}

var Plans = map[string]PlanLimits{
	"free": {
		MaxProjects:  3,
		MaxMembers:   5,
		MaxStorageMB: 100,
	},
	"pro": {
		MaxProjects:  50,
		MaxMembers:   50,
		MaxStorageMB: 10240,
	},
	"enterprise": {
		MaxProjects:  -1, // unlimited
		MaxMembers:   -1,
		MaxStorageMB: -1,
	},
}

func GetLimits(plan string) PlanLimits {
	if l, ok := Plans[plan]; ok {
		return l
	}
	return Plans["free"]
}
