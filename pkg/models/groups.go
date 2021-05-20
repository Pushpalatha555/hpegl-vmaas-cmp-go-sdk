// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package models

// InstanceCounts
type InstanceCounts struct {
	All int `json:"all"`
}

// ServerCounts
type ServerCounts struct {
	All           int `json:"all"`
	Host          int `json:"host"`
	Hypervisor    int `json:"hypervisor"`
	ContainerHost int `json:"containerHost"`
	Vm            int `json:"vm"`
	Baremetal     int `json:"baremetal"`
	Unmanaged     int `json:"unmanaged"`
}

// GroupsZones
type GroupsZones struct {
}

// GroupsStats
type GroupsStats struct {
	InstanceCounts GroupsStatsInstanceCounts `json:"instanceCounts"`
	ServerCounts   GroupsStatsServerCounts   `json:"serverCounts"`
}

// GroupsStatsInstanceCounts
type GroupsStatsInstanceCounts struct {
	All int `json:"all"`
}

// GroupsStatsServerCounts
type GroupsStatsServerCounts struct {
	All           int `json:"all"`
	Host          int `json:"host"`
	Hypervisor    int `json:"hypervisor"`
	ContainerHost int `json:"containerHost"`
	Vm            int `json:"vm"`
	Baremetal     int `json:"baremetal"`
	Unmanaged     int `json:"unmanaged"`
}

// Groups
type Groups struct {
	Groups *[]Group `json:"groups"`
}

// GroupResp
type GroupResp struct {
	Group *Group `json:"group"`
}

// Group
type Group struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Location    string  `json:"location"`
	AccountId   int     `json:"accountId"`
	Zones       []Zones `json:"zones"`
	ServerCount int     `json:"serverCount"`
	Id          int     `json:"id"`
	Active      bool    `json:"active"`
	DateCreated string  `json:"dateCreated"`
	LastUpdated string  `json:"lastUpdated"`
	Stats       Stats   `json:"stats"`
}

// Zones
type Zones struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Stats
type Stats struct {
	InstanceCounts InstanceCounts `json:"instanceCounts"`
	ServerCounts   ServerCounts   `json:"serverCounts"`
}
