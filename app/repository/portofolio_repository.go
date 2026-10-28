package repository

import (
	"context"
	"errors"
	"latihan-fiber2/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPortfolioNotFound    = errors.New("portofolio tidak ditemukan")
	ErrProjectLimitExceeded = errors.New("jumlah proyek sudah mencapai batas maksimal 20")
)

type PortfolioRepository interface {
	GetByStudentID(ctx context.Context, studentID int) (*model.Portfolio, error)
	UpsertPortfolio(ctx context.Context, studentID int, p model.Portfolio) (*model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, proj model.Project) (*model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	GetTopTechnologies(ctx context.Context) ([]bson.M, error)
}

type portfolioRepo struct {
	db *mongo.Database
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioRepo{db: db}
}

func (r *portfolioRepo) collection() *mongo.Collection {
	return r.db.Collection("student_portfolios")
}

// Mapping internal MongoDB BSON <-> Go Model
type portfolioDoc struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty"`
	StudentID    int                 `bson:"student_id"`
	Headline     string              `bson:"headline"`
	Bio          string              `bson:"bio"`
	Projects     []projectDoc        `bson:"projects"`
	Certificates []model.Certificate `bson:"certificates"`
}

type projectDoc struct {
	ID           string   `bson:"_id,omitempty"`
	Title        string   `bson:"title"`
	Technologies []string `bson:"technologies"`
	Link         string   `bson:"link"`
	Year         int      `bson:"year"`
}

func (doc portfolioDoc) toModel() model.Portfolio {
	projects := make([]model.Project, len(doc.Projects))
	for i, p := range doc.Projects {
		projects[i] = model.Project{
			ID:           p.ID,
			Title:        p.Title,
			Technologies: p.Technologies,
			Link:         p.Link,
			Year:         p.Year,
		}
	}
	return model.Portfolio{
		ID:           doc.ID.Hex(),
		StudentID:    doc.StudentID,
		Headline:     doc.Headline,
		Bio:          doc.Bio,
		Projects:     projects,
		Certificates: doc.Certificates,
	}
}

func (r *portfolioRepo) GetByStudentID(ctx context.Context, studentID int) (*model.Portfolio, error) {
	var doc portfolioDoc
	err := r.collection().FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		return nil, ErrPortfolioNotFound
	}
	if err != nil {
		return nil, err
	}
	p := doc.toModel()
	return &p, nil
}

func (r *portfolioRepo) UpsertPortfolio(ctx context.Context, studentID int, p model.Portfolio) (*model.Portfolio, error) {
	filter := bson.M{"student_id": studentID}
	update := bson.M{
		"$set": bson.M{
			"headline":     p.Headline,
			"bio":          p.Bio,
			"projects":     p.Projects,
			"certificates": p.Certificates,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var doc portfolioDoc
	err := r.collection().FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err != nil {
		return nil, err
	}
	result := doc.toModel()
	return &result, nil
}

func (r *portfolioRepo) AddProject(ctx context.Context, studentID int, proj model.Project) (*model.Portfolio, error) {
	if proj.ID == "" {
		proj.ID = primitive.NewObjectID().Hex()
	}

	filter := bson.M{
		"student_id": studentID,
		"projects.19": bson.M{"$exists": false},
	}

	update := bson.M{
		"$push": bson.M{
			"projects": proj,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var doc portfolioDoc
	err := r.collection().FindOneAndUpdate(ctx, filter, update, opts).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		var existing portfolioDoc
		if errCheck := r.collection().FindOne(ctx, bson.M{"student_id": studentID}).Decode(&existing); errCheck == nil {
			if len(existing.Projects) >= 20 {
				return nil, ErrProjectLimitExceeded
			}
		}
		return nil, ErrPortfolioNotFound
	}
	if err != nil {
		return nil, err
	}
	result := doc.toModel()
	return &result, nil
}

func (r *portfolioRepo) DeleteProject(ctx context.Context, studentID int, projectID string) error {
	filter := bson.M{"student_id": studentID}
	update := bson.M{
		"$pull": bson.M{
			"projects": bson.M{"_id": projectID},
		},
	}
	res, err := r.collection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrPortfolioNotFound
	}
	return nil
}

func (r *portfolioRepo) GetTopTechnologies(ctx context.Context) ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$projects.technologies"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.collection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}