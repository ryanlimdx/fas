// Contains the structure of the entities involved.
package models

type Scheme struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Criteria []Criteria `json:"criteria,omitempty"`
	Benefits []Benefit `json:"benefits,omitempty"`
}

type Criteria struct {
	ID	string `json:"id"`
	CriteriaLevel string `json:"criteria_level"`
	CriteriaType string `json:"criteria_type"`
	Status string `json:"status"`
}

type Benefit struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Amount float64 `json:"amount"`
}