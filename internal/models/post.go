package models

import (
	"kimistore/internal/utils"
	"kimistore/pkg/http/paging"
	"time"
)

type Post struct {
	BaseModel
	Title       string `json:"title" gorm:"column:title;type:varchar(255)"`
	Publish     string `json:"publish" gorm:"type:text;column:publish;default:draft"`
	Content     string `json:"content" gorm:"type:text;column:content"`
	CoverURL    string `json:"cover_url" gorm:"type:text;column:cover_url"`
	Description string `json:"description" gorm:"type:text;column:description"`
}

func (Post) TableName() string {
	return "posts"
}

type OriginalPost struct {
	ID       string `json:"id"`
	Publish  string `json:"publish"`
	Comments []struct {
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
	} `json:"comments"`
	MetaKeywords    []string  `json:"metaKeywords"`
	Content         string    `json:"content"`
	Tags            []string  `json:"tags"`
	MetaTitle       string    `json:"metaTitle"`
	CreatedAt       time.Time `json:"createdAt"`
	Title           string    `json:"title"`
	CoverURL        string    `json:"coverUrl"`
	TotalViews      int       `json:"totalViews"`
	TotalShares     int       `json:"totalShares"`
	TotalComments   int       `json:"totalComments"`
	TotalFavorites  int       `json:"totalFavorites"`
	MetaDescription string    `json:"metaDescription"`
	Description     string    `json:"description"`
	Author          struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
	} `json:"author"`
	FavoritePerson []struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
	} `json:"favoritePerson"`
}

type CreatePostRequest struct {
	Title       *string `json:"title" binding:"required"`
	Publish     *string `json:"publish" binding:"required"`
	Content     *string `json:"content" binding:"required"`
	CoverURL    *string `json:"cover_url" binding:"required"`
	Description *string `json:"description" binding:"required"`
}

type GetPostResponseData struct {
	Post OriginalPost `json:"post"`
}

type GetPostResponse struct {
	Meta *utils.MetaData      `json:"meta"`
	Data *GetPostResponseData `json:"data"`
}

type PostFilterRequest struct {
	Publish *string `json:"publish" form:"publish"`
}

type ListPostFilter struct {
	PostFilterRequest
	Pager *paging.Pager
}

type ListPostResult struct {
	Filter  *ListPostFilter
	Records []Post `json:"posts"`
}

type ListPostResponse struct {
	Filter  *ListPostFilter
	Records []OriginalPost `json:"posts"`
}

type UpdatePostRequest struct {
	Title       *string `json:"title"`
	Publish     *string `json:"publish"`
	Content     *string `json:"content"`
	CoverURL    *string `json:"cover_url"`
	Description *string `json:"description"`
}

type DeletePostResponse struct {
	Meta    *utils.MetaData `json:"meta"`
	Message string          `json:"message"`
}
