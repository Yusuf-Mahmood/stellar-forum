package models

import "time"

type Data struct {
	UserProfile     []UserProfile
	Post            []Post
	MemesPosts      []MemesPosts
	GamingPosts     []GamingPosts
	EducationPosts  []EducationPosts
	TechnologyPosts []TechnologyPosts
	SciencePosts    []SciencePosts
	SportsPosts     []SportsPosts
}

// Post represents a post with user and content information
type Post struct {
	ID         int
	UserID     int
	Username   string
	Content    string
	CreatedAt  time.Time
	FormatDate string
	Media      []Media
	Likes      int
	Dislikes   int
	LikeIcon string
	DislikeIcon string
	ProfileColor string
	ComCount   int
	Comment    []Comment
}

type MemesPosts struct {
	CategoriesID int
	PostID       int
	UserID       int
	Username     string
	Content      string
	CreatedAt    time.Time
	FormatDate   string
	Media        []Media
	Likes        int
	Dislikes     int
	LikeIcon string
	DislikeIcon string
	ProfileColor string
	ComCount     int
	Comment      []Comment
}

type GamingPosts struct {
	CategoriesID int
	PostID       int
	UserID       int
	Username     string
	Content      string
	CreatedAt    time.Time
	FormatDate   string
	Media        []Media
	Likes        int
	Dislikes     int
	LikeIcon string
	DislikeIcon string
	ProfileColor string
	ComCount     int
	Comment      []Comment
}

type EducationPosts struct {
	CategoriesID int
	PostID       int
	UserID       int
	Username     string
	Content      string
	CreatedAt    time.Time
	FormatDate   string
	Media        []Media
	Likes        int
	Dislikes     int
	LikeIcon string
	DislikeIcon string
	ProfileColor string
	ComCount     int
	Comment      []Comment
}

type TechnologyPosts struct {
	CategoriesID int
	PostID       int
	UserID       int
	Username     string
	Content      string
	CreatedAt    time.Time
	FormatDate   string
	Media        []Media
	Likes        int
	Dislikes     int
	LikeIcon string
	DislikeIcon string
	ProfileColor string
	ComCount     int
	Comment      []Comment
}

type SciencePosts struct {
	CategoriesID int
	PostID       int
	UserID       int
	Username     string
	Content      string
	CreatedAt    time.Time
	FormatDate   string
	Media        []Media
	Likes        int
	Dislikes     int
	LikeIcon string
	DislikeIcon string
	ProfileColor string
	ComCount     int
	Comment      []Comment
}

type SportsPosts struct {
	CategoriesID int
	PostID       int
	UserID       int
	Username     string
	Content      string
	CreatedAt    time.Time
	FormatDate   string
	Media        []Media
	Likes        int
	Dislikes     int
	LikeIcon string
	DislikeIcon string
	ProfileColor string
	ComCount     int
	Comment      []Comment
}

type Comment struct {
	ComID         int
	PostID        int
	ComUsername   string
	ComContent    string
	ComCreatedAt  time.Time
	ComFormatDate string
	ComLikes      int
	ComDislikes   int
	ComLikeIcon 	  string
	ComDislikeIcon string
	ComProfile string
}

// UserProfile struct to hold user profile data, including posts liked, created, and disliked
type UserProfile struct {
	UserID        int
	Username      string
	ProfileColor  string
	LikedPosts    []Post
	CreatedPosts  []Post
	DislikedPosts []Post
}

// Media represents a media file linked to a post
type Media struct {
	FilePath string
	FileType string
}
