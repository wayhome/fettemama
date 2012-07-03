package shared

type BlogDB interface {
	Connect()
	Disconnect()

	StorePost(post *BlogPost) (int64, error)
	GetPost(post_id int64) (BlogPost, error)
	GetPostsForTimespan(start_timestamp, end_timestamp, order int64) (posts []BlogPost, err error)
	GetLastNPosts(num_to_get int32) (posts []BlogPost, err error)

	StoreComment(comment *PostComment) (int64, error)
	GetComments(post_id int64) (comments []PostComment, err error)
}
