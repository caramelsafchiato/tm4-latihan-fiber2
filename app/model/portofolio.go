package model

type Portfolio struct {
	ID           string        `json:"id" bson:"_id,omitempty"`
	StudentID    int           `json:"student_id" bson:"student_id"`
	Headline     string        `json:"headline" bson:"headline"`
	Bio          string        `json:"bio" bson:"bio"`
	Projects     []Project     `json:"projects" bson:"projects"`
	Certificates []Certificate `json:"certificates" bson:"certificates"`
}

type Project struct {
	ID           string   `json:"id" bson:"_id,omitempty"`
	Title        string   `json:"title" bson:"title"`
	Technologies []string `json:"technologies" bson:"technologies"`
	Link         string   `json:"link" bson:"link"`
	Year         int      `json:"year" bson:"year"`
}

type Certificate struct {
	Name   string `json:"name" bson:"name"`
	Issuer string `json:"issuer" bson:"issuer"`
	Date   string `json:"date" bson:"date"`
}

type CreateUpdatePortfolioRequest struct {
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
}

type AddProjectRequest struct {
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	Link         string   `json:"link"`
	Year         int      `json:"year"`
}