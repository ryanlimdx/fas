// Contains the structure of the entities involved.
package models

type Applicant struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	EmploymentStatus string      `json:"employment_status"`
	MaritalStatus    string      `json:"marital_status"`
	Sex			     string      `json:"sex"`
	DateOfBirth      string      `json:"date_of_birth"`
	Household       []Household  `json:"household"`
}

type Household struct {
	ID               string      `json:"id"`
	ApplicantID      string      `json:"applicant_id"`
	Name             string      `json:"name"`
	Relationship     string      `json:"relationship"`
	Sex			     string      `json:"sex"`
	SchoolLevel      string      `json:"school_level"`
	EmploymentStatus string      `json:"employment_status"`
	DateOfBirth      string      `json:"date_of_birth"`
}