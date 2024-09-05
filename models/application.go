// Contains the structure of the entities involved.
package models

type Application struct {
	ID          string `json:"id"`
	ApplicantID string `json:"applicant_id"`
	SchemeID    string `json:"scheme_id"`
	Status      string `json:"status"`
	AppliedDate string `json:"applied_date"`
}
