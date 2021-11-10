// (C) Copyright 2021 Hewlett Packard Enterprise Development LP
package models

type CmpSetupCheck struct {
	Success      bool   `json:"success"`
	BuildVersion string `json:"buildVersion"`
	ApplianceUrl string `json:"applianceUrl"`
	SetupNeeded  bool   `json:"setupNeeded"`
}
