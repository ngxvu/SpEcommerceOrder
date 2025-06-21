package repo

import (
	model "basesource/internal/models"
	pgGorm "basesource/internal/repo/pg-gorm"
	"basesource/internal/utils"
	"context"
	"gorm.io/gorm"
	"time"
)

type PostRepository struct {
	db pgGorm.PGInterface
}

func NewPostRepository(newPgRepo pgGorm.PGInterface) *PostRepository {
	return &PostRepository{
		db: newPgRepo,
	}
}

type PostRepositoryInterface interface {
	CreatePost(ctx context.Context, tx *gorm.DB, post model.Post) (*model.GetPostResponse, error)
	CheckPostExistById(ctx context.Context, tx *gorm.DB, id string) (bool, error)
	PostExistsByName(ctx context.Context, tx *gorm.DB, title string) (bool, error)
	GetPostV1(ctx context.Context, tx *gorm.DB, id string) (*model.Post, error)
	GetDetailPost(ctx context.Context, tx *gorm.DB, id string) (*model.GetPostResponse, error)
	GetListPost(ctx context.Context, filter *model.ListPostFilter, pgRepo *gorm.DB) (*model.ListPostResponse, error)
	UpdatePost(ctx context.Context, tx *gorm.DB, post model.Post) (*model.GetPostResponse, error)
	DeletePost(ctx context.Context, tx *gorm.DB, post model.Post) (*model.DeletePostResponse, error)
}

func (p *PostRepository) CheckPostExistById(ctx context.Context, tx *gorm.DB, id string) (bool, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	var count int64
	if err := tx.Model(&model.Post{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *PostRepository) PostExistsByName(ctx context.Context, tx *gorm.DB, title string) (bool, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	var count int64
	if err := tx.Model(&model.Post{}).Where("title = ?", title).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *PostRepository) GetPostV1(ctx context.Context, tx *gorm.DB, id string) (*model.Post, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	var post model.Post
	if err := tx.Where("id = ?", id).First(&post).Error; err != nil {
		return nil, err
	}

	return &post, nil
}

func (p *PostRepository) CreatePost(ctx context.Context, tx *gorm.DB, post model.Post) (*model.GetPostResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	// Save the post to the database
	if err := tx.WithContext(ctx).Create(&post).Error; err != nil {
		return nil, err
	}

	// Map the post to the response format
	responseData := p.mapperPostsToResponse(post)

	response := &model.GetPostResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &responseData,
	}

	return response, nil

}

func (p *PostRepository) mapperPostsToResponse(post model.Post) model.GetPostResponseData {
	postID := utils.UUIDtoString(post.ID)

	responseData := &model.GetPostResponseData{}
	responseData.Post.ID = postID
	responseData.Post.Title = post.Title
	responseData.Post.Publish = post.Publish
	responseData.Post.Content = post.Content
	responseData.Post.CoverURL = post.CoverURL
	responseData.Post.Description = post.Description
	responseData.Post.CreatedAt = post.CreatedAt

	// Set default values for fields not in Post model
	responseData.Post.Comments = []struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		AvatarURL string    `json:"avatarUrl"`
		Message   string    `json:"message"`
		PostedAt  time.Time `json:"postedAt"`
		Users     []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatarUrl"`
		} `json:"users"`
		ReplyComment []struct {
			ID       string    `json:"id"`
			UserID   string    `json:"userId"`
			Message  string    `json:"message"`
			PostedAt time.Time `json:"postedAt"`
			TagUser  string    `json:"tagUser,omitempty"`
		} `json:"replyComment"`
	}{}
	responseData.Post.MetaKeywords = []string{}
	responseData.Post.Tags = []string{}
	responseData.Post.MetaTitle = ""
	responseData.Post.TotalViews = 0
	responseData.Post.TotalShares = 0
	responseData.Post.TotalComments = 0
	responseData.Post.TotalFavorites = 0
	responseData.Post.MetaDescription = ""
	responseData.Post.Author = struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
	}{}
	responseData.Post.FavoritePerson = []struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
	}{}

	return *responseData
}

func (p *PostRepository) GetDetailPost(ctx context.Context, tx *gorm.DB, id string) (*model.GetPostResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	var post model.Post
	if err := tx.Where("id = ?", id).First(&post).Error; err != nil {
		return nil, err
	}

	responseData := p.mapperPostsToResponse(post)

	response := &model.GetPostResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &responseData,
	}

	return response, nil
}

func (p *PostRepository) GetListPost(ctx context.Context, filter *model.ListPostFilter, tx *gorm.DB) (*model.ListPostResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	query := tx.Model(&model.Post{})

	result := &model.ListPostResult{
		Filter:  filter,
		Records: []model.Post{},
	}

	if filter.PostFilterRequest.Publish != nil {
		query = query.Where("publish = ?", *filter.PostFilterRequest.Publish)
	}

	filter.Pager.SortableFields = []string{"created_at"}

	pager := filter.Pager

	err := pager.DoQuery(&result.Records, query).Error
	if err != nil {
		return nil, err

	}

	mapper := model.GetPostResponseData{}

	var mapperList []model.OriginalPost

	for i := 0; i < len(result.Records); i++ {
		mapper = p.mapperPostsToResponse(result.Records[i])
		mapperList = append(mapperList, mapper.Post)
	}

	response := &model.ListPostResponse{
		Filter:  filter,
		Records: mapperList,
	}

	return response, nil
}

func (p *PostRepository) UpdatePost(ctx context.Context, tx *gorm.DB, post model.Post) (*model.GetPostResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	// Save the updated post to the database
	if err := tx.WithContext(ctx).Save(&post).Error; err != nil {
		return nil, err
	}

	responseData := p.mapperPostsToResponse(post)

	response := &model.GetPostResponse{
		Meta: utils.NewMetaData(ctx),
		Data: &responseData,
	}

	return response, nil
}

func (p *PostRepository) DeletePost(ctx context.Context, tx *gorm.DB, post model.Post) (*model.DeletePostResponse, error) {

	var cancel context.CancelFunc
	if tx == nil {
		tx, cancel = p.db.DBWithTimeout(ctx)
		defer cancel()
	}

	// Delete the post from the database
	if err := tx.WithContext(ctx).Delete(&post).Error; err != nil {
		return nil, err
	}

	return &model.DeletePostResponse{
		Meta:    utils.NewMetaData(ctx),
		Message: "Post deleted successfully",
	}, nil
}
