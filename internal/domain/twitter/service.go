package twitter

type TwitterService interface {
	NewTweet(tweetData Tweet) error
	GetTweet(id int64) (Tweet, error)
	GetUsersTweets(userId int64) ([]Tweet, error)
}
