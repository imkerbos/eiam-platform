package models

// OrganizationType organization type enum
type OrganizationType int

const (
	OrgTypeHeadquarters OrganizationType = 1 // headquarters
	OrgTypeBranch       OrganizationType = 2 // branch
	OrgTypeDepartment   OrganizationType = 3 // department
	OrgTypeTeam         OrganizationType = 4 // team
)

func (o OrganizationType) String() string {
	switch o {
	case OrgTypeHeadquarters:
		return "headquarters"
	case OrgTypeBranch:
		return "branch"
	case OrgTypeDepartment:
		return "department"
	case OrgTypeTeam:
		return "team"
	default:
		return "unknown"
	}
}

// Organization organization structure
type Organization struct {
	BaseModel
	Name        string           `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Code        string           `json:"code" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,max=50"`
	Type        OrganizationType `json:"type" gorm:"type:tinyint;not null;index"` // organization type
	ParentID    *string          `json:"parent_id" gorm:"type:varchar(36);index"`
	Level       int              `json:"level" gorm:"default:1"`
	Path        string           `json:"path" gorm:"type:varchar(500);index"` // organization path, e.g.: /1/2/3
	Sort        int              `json:"sort" gorm:"default:0"`
	Description string           `json:"description" gorm:"type:varchar(500)"`
	Manager     string           `json:"manager" gorm:"type:varchar(36);index"` // manager user ID
	Location    string           `json:"location" gorm:"type:varchar(200)"`     // location
	Phone       string           `json:"phone" gorm:"type:varchar(50)"`         // contact phone
	Email       string           `json:"email" gorm:"type:varchar(100)"`        // contact email
	Status      Status           `json:"status" gorm:"type:tinyint;default:1;index"`

	// Relationships
	Parent      *Organization  `json:"parent" gorm:"foreignKey:ParentID"`
	Children    []Organization `json:"children" gorm:"foreignKey:ParentID"`
	ManagerUser *User          `json:"manager_user" gorm:"foreignKey:Manager;references:ID"`
	Groups      []Group        `json:"groups" gorm:"foreignKey:OrganizationID"`
	Users       []User         `json:"users" gorm:"many2many:organization_users;"`
}

// TableName specify table name
func (Organization) TableName() string {
	return "organizations"
}

// Group user group
type Group struct {
	BaseModel
	Name           string  `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Code           string  `json:"code" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,max=50"`
	Description    string  `json:"description" gorm:"type:varchar(500)"`
	OrganizationID *string `json:"organization_id" gorm:"type:varchar(36);index"`
	Status         Status  `json:"status" gorm:"type:tinyint;default:1;index"`

	// Relationships
	Organization *Organization `json:"organization" gorm:"foreignKey:OrganizationID"`
	Users        []User        `json:"users" gorm:"many2many:user_groups;"`
	Roles        []Role        `json:"roles" gorm:"many2many:group_roles;"`
}

// TableName specify table name
func (Group) TableName() string {
	return "groups"
}

// Role role
type Role struct {
	BaseModel
	Name        string  `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Code        string  `json:"code" gorm:"type:varchar(50);uniqueIndex;not null" validate:"required,max=50"`
	Description string  `json:"description" gorm:"type:varchar(500)"`
	Type        string  `json:"type" gorm:"type:varchar(50);default:'custom'"`  // system, custom, application
	IsSystem    bool    `json:"is_system" gorm:"default:false"`                 // is system role
	Scope       string  `json:"scope" gorm:"type:varchar(50);default:'global'"` // global, organization, application
	ScopeID     *string `json:"scope_id" gorm:"type:varchar(36);index"`         // scope ID (organization ID or application ID)
	Status      Status  `json:"status" gorm:"type:tinyint;default:1;index"`

	// Relationships
	Users             []User             `json:"users" gorm:"many2many:user_roles;"`
	Groups            []Group            `json:"groups" gorm:"many2many:group_roles;"`
	Permissions       []Permission       `json:"permissions" gorm:"many2many:role_permissions;"`
	Applications      []Application      `json:"applications" gorm:"many2many:role_applications;"`
	ApplicationGroups []ApplicationGroup `json:"application_groups" gorm:"many2many:role_application_groups;"`
}

// TableName specify table name
func (Role) TableName() string {
	return "roles"
}

// Permission permission
type Permission struct {
	BaseModel
	Name          string  `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Code          string  `json:"code" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required,max=100"`
	Resource      string  `json:"resource" gorm:"type:varchar(100);not null"` // resource
	Action        string  `json:"action" gorm:"type:varchar(50);not null"`    // action: create, read, update, delete, execute
	Description   string  `json:"description" gorm:"type:varchar(500)"`
	Category      string  `json:"category" gorm:"type:varchar(50);not null"`    // permission category: system, application, data
	ApplicationID *string `json:"application_id" gorm:"type:varchar(36);index"` // related application ID
	IsSystem      bool    `json:"is_system" gorm:"default:false"`               // is system permission
	Status        Status  `json:"status" gorm:"type:tinyint;default:1;index"`

	// Relationships
	Roles       []Role       `json:"roles" gorm:"many2many:role_permissions;"`
	Application *Application `json:"application" gorm:"foreignKey:ApplicationID"`
}

// TableName specify table name
func (Permission) TableName() string {
	return "permissions"
}
