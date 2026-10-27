package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"latihan-fiber2/app/model"
)

var (
	ErrInvalidID    = errors.New("format id tidak valid")
	ErrNotOwner     = errors.New("bukan pemilik data")
	ErrOwnReview    = errors.New("tidak boleh menilai ulasan sendiri")
	ErrAlreadyVoted = errors.New("sudah pernah menilai ulasan ini")
	ErrReplyLimit   = errors.New("jumlah balasan sudah mencapai batas")
)

const MaxRepliesPerReview = 50

type ReviewRepository interface {
	EnsureIndexes(ctx context.Context) error
	Create(ctx context.Context, r model.Review) (model.Review, error)
	FindByID(ctx context.Context, id string) (model.Review, error)
	FindByCourse(ctx context.Context, q model.ReviewQuery) ([]model.Review, error)
	UpdateOwn(ctx context.Context, id string, userID int, req model.UpdateReviewRequest) (model.Review, error)
	Delete(ctx context.Context, id string, userID int, anyOwner bool) error
	AddReply(ctx context.Context, id string, reply model.Reply) (model.Review, error)
	MarkHelpful(ctx context.Context, id string, userID int) (model.Review, error)
	Summary(ctx context.Context, courseID int) (model.ReviewSummary, error)
}

// reviewDocument adalah bentuk ulasan SEBAGAIMANA TERSIMPAN di MongoDB
type reviewDocument struct {
	ID           bson.ObjectID   `bson:"_id,omitempty"`
	CourseID     int             `bson:"course_id"`
	UserID       int             `bson:"user_id"`
	Username     string          `bson:"username"`
	Rating       int             `bson:"rating"`
	Comment      string          `bson:"comment"`
	Tags         []string        `bson:"tags"`
	Replies      []replyDocument `bson:"replies"`
	HelpfulCount int             `bson:"helpful_count"`
	HelpfulBy    []int           `bson:"helpful_by"`
	CreatedAt    time.Time       `bson:"created_at"`
	UpdatedAt    time.Time       `bson:"updated_at"`
}

type replyDocument struct {
	ID        bson.ObjectID `bson:"_id"`
	UserID    int           `bson:"user_id"`
	Username  string        `bson:"username"`
	Body      string        `bson:"body"`
	CreatedAt time.Time     `bson:"created_at"`
}

func (d reviewDocument) toModel() model.Review {
	replies := make([]model.Reply, 0, len(d.Replies))
	for _, r := range d.Replies {
		replies = append(replies, model.Reply{
			ID:        r.ID.Hex(),
			UserID:    r.UserID,
			Username:  r.Username,
			Body:      r.Body,
			CreatedAt: r.CreatedAt,
		})
	}

	tags := d.Tags
	if tags == nil {
		tags = []string{}
	}

	return model.Review{
		ID:           d.ID.Hex(),
		CourseID:     d.CourseID,
		UserID:       d.UserID,
		Username:     d.Username,
		Rating:       d.Rating,
		Comment:      d.Comment,
		Tags:         tags,
		Replies:      replies,
		HelpfulCount: d.HelpfulCount,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

type reviewMongoRepository struct {
	coll *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewMongoRepository{coll: db.Collection("course_reviews")}
}

func now() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func parseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, ErrInvalidID
	}
	return oid, nil
}

// EnsureIndexes dijalankan sekali saat aplikasi menyala. Membuat index yang
// sudah ada tidak melakukan apa pun, sehingga aman dipanggil berulang kali.
func (r *reviewMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("course_user_unique"),
		},
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "_id", Value: -1}},
			Options: options.Index().SetName("course_newest"),
		},
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "tags", Value: 1}},
			Options: options.Index().SetName("course_tags"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index course_reviews: %w", err)
	}
	return nil
}

func (r *reviewMongoRepository) Create(ctx context.Context, rv model.Review) (model.Review, error) {
	t := now()
	doc := reviewDocument{
		ID:           bson.NewObjectID(),
		CourseID:     rv.CourseID,
		UserID:       rv.UserID,
		Username:     rv.Username,
		Rating:       rv.Rating,
		Comment:      rv.Comment,
		Tags:         rv.Tags,
		Replies:      []replyDocument{},
		HelpfulBy:    []int{},
		CreatedAt:    t,
		UpdatedAt:    t,
	}
	if doc.Tags == nil {
		doc.Tags = []string{}
	}

	if _, err := r.coll.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return model.Review{}, ErrDuplicate
		}
		return model.Review{}, fmt.Errorf("menyimpan ulasan: %w", err)
	}
	return doc.toModel(), nil
}

func (r *reviewMongoRepository) FindByID(ctx context.Context, id string) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}

	var doc reviewDocument
	if err := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Review{}, ErrNotFound
		}
		return model.Review{}, fmt.Errorf("mengambil ulasan: %w", err)
	}
	return doc.toModel(), nil
}

// FindByCourse mengambil ulasan satu kelas, terbaru lebih dulu.
func (r *reviewMongoRepository) FindByCourse(ctx context.Context, q model.ReviewQuery) ([]model.Review, error) {
	filter := bson.M{"course_id": q.CourseID}

	if q.Tag != "" {
		filter["tags"] = q.Tag
	}
	if q.Before != "" {
		before, err := parseID(q.Before)
		if err != nil {
			return nil, err
		}
		filter["_id"] = bson.M{"$lt": before}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(int64(q.Limit))
	opts.SetProjection(bson.M{"helpful_by": 0})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar ulasan: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []reviewDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("membaca daftar ulasan: %w", err)
	}

	result := make([]model.Review, 0, len(docs))
	for _, d := range docs {
		result = append(result, d.toModel())
	}
	return result, nil
}

func (r *reviewMongoRepository) UpdateOwn(ctx context.Context, id string, userID int, req model.UpdateReviewRequest) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}

	updateFields := bson.M{"updated_at": now()}
	if req.Rating != nil {
		updateFields["rating"] = *req.Rating
	}
	if req.Comment != nil {
		updateFields["comment"] = *req.Comment
	}
	if req.Tags != nil {
		updateFields["tags"] = *req.Tags
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDoc reviewDocument
	err = r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid, "user_id": userID},
		bson.M{"$set": updateFields},
		opts,
	).Decode(&updatedDoc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var existing reviewDocument
			if errCheck := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&existing); errCheck != nil {
				return model.Review{}, ErrNotFound
			}
			return model.Review{}, ErrNotOwner
		}
		return model.Review{}, fmt.Errorf("mengubah ulasan: %w", err)
	}

	return updatedDoc.toModel(), nil
}

func (r *reviewMongoRepository) Delete(ctx context.Context, id string, userID int, anyOwner bool) error {
	oid, err := parseID(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": oid}
	if !anyOwner {
		filter["user_id"] = userID
	}

	res, err := r.coll.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("menghapus ulasan: %w", err)
	}
	if res.DeletedCount == 0 {
		var existing reviewDocument
		if errCheck := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&existing); errCheck != nil {
			return ErrNotFound
		}
		return ErrNotOwner
	}
	return nil
}

func (r *reviewMongoRepository) AddReply(ctx context.Context, id string, reply model.Reply) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}

	replyDoc := replyDocument{
		ID:        bson.NewObjectID(),
		UserID:    reply.UserID,
		Username:  reply.Username,
		Body:      reply.Body,
		CreatedAt: now(),
	}

	filter := bson.M{
		"_id": oid,
		"$expr": bson.M{
			"$lt": bson.A{bson.M{"$size": "$replies"}, MaxRepliesPerReview},
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDoc reviewDocument
	err = r.coll.FindOneAndUpdate(
		ctx,
		filter,
		bson.M{
			"$push": bson.M{"replies": replyDoc},
			"$set":  bson.M{"updated_at": now()},
		},
		opts,
	).Decode(&updatedDoc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var existing reviewDocument
			if errCheck := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&existing); errCheck != nil {
				return model.Review{}, ErrNotFound
			}
			if len(existing.Replies) >= MaxRepliesPerReview {
				return model.Review{}, ErrReplyLimit
			}
		}
		return model.Review{}, fmt.Errorf("menambahkan balasan: %w", err)
	}

	return updatedDoc.toModel(), nil
}

func (r *reviewMongoRepository) MarkHelpful(ctx context.Context, id string, userID int) (model.Review, error) {
	oid, err := parseID(id)
	if err != nil {
		return model.Review{}, err
	}

	filter := bson.M{
		"_id":        oid,
		"user_id":    bson.M{"$ne": userID},
		"helpful_by": bson.M{"$ne": userID},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedDoc reviewDocument
	err = r.coll.FindOneAndUpdate(
		ctx,
		filter,
		bson.M{
			"$inc":  bson.M{"helpful_count": 1},
			"$push": bson.M{"helpful_by": userID},
		},
		opts,
	).Decode(&updatedDoc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			var existing reviewDocument
			if errCheck := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&existing); errCheck != nil {
				return model.Review{}, ErrNotFound
			}
			if existing.UserID == userID {
				return model.Review{}, ErrOwnReview
			}
			return model.Review{}, ErrAlreadyVoted
		}
		return model.Review{}, fmt.Errorf("menandai ulasan bermanfaat: %w", err)
	}

	return updatedDoc.toModel(), nil
}

func (r *reviewMongoRepository) Summary(ctx context.Context, courseID int) (model.ReviewSummary, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"course_id": courseID}}},
		{{Key: "$facet", Value: bson.M{
			"stats": bson.A{
				bson.M{"$group": bson.M{
					"_id":       nil,
					"count":     bson.M{"$sum": 1},
					"avgRating": bson.M{"$avg": "$rating"},
				}},
			},
			"distribution": bson.A{
				bson.M{"$group": bson.M{
					"_id":   "$rating",
					"count": bson.M{"$sum": 1},
				}},
			},
			"top_tags": bson.A{
				bson.M{"$unwind": "$tags"},
				bson.M{"$group": bson.M{
					"_id":   "$tags",
					"count": bson.M{"$sum": 1},
				}},
				bson.M{"$sort": bson.M{"count": -1, "_id": 1}},
				bson.M{"$limit": 5},
			},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return model.ReviewSummary{}, fmt.Errorf("menghitung ringkasan ulasan: %w", err)
	}
	defer cursor.Close(ctx)

	type summaryResult struct {
		Stats []struct {
			Count     int     `bson:"count"`
			AvgRating float64 `bson:"avgRating"`
		} `bson:"stats"`
		Distribution []struct {
			Rating int `bson:"_id"`
			Count  int `bson:"count"`
		} `bson:"distribution"`
		TopTags []struct {
			Tag   string `bson:"_id"`
			Count int    `bson:"count"`
		} `bson:"top_tags"`
	}

	var results []summaryResult
	if err := cursor.All(ctx, &results); err != nil {
		return model.ReviewSummary{}, fmt.Errorf("membaca ringkasan ulasan: %w", err)
	}

	distMap := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
	summary := model.ReviewSummary{
		CourseID:     courseID,
		Distribution: distMap,
		TopTags:      []model.TagCount{},
	}

	if len(results) == 0 || len(results[0].Stats) == 0 {
		return summary, nil
	}

	summary.Count = results[0].Stats[0].Count
	summary.AverageRating = results[0].Stats[0].AvgRating

	for _, d := range results[0].Distribution {
		distMap[fmt.Sprintf("%d", d.Rating)] = d.Count
	}
	summary.Distribution = distMap

	for _, t := range results[0].TopTags {
		summary.TopTags = append(summary.TopTags, model.TagCount{
			Tag:   t.Tag,
			Total: t.Count,
		})
	}

	return summary, nil
}