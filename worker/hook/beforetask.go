package hook

import "unfire/model"

// BeforeRunTaskHook ツイートが削除される前に実行される。
func BeforeRunTaskHook(u *model.User, tts []model.TweetSimple) ([]model.TweetSimple, error) {
	tweets := tts
	var err error

	for _, f := range getPreRunTaskHooks() {
		tweets, err = f(u, tweets)
		if err != nil {
			return []model.TweetSimple{}, nil
		}
	}

	return tweets, nil
}

// getPreRunTaskHooks Hook を取得する。ここにHookを追加していく。
func getPreRunTaskHooks() []func(u *model.User, tts []model.TweetSimple) ([]model.TweetSimple, error) {
	return []func(u *model.User, tts []model.TweetSimple) ([]model.TweetSimple, error){keepLegendaryTweetsV1}
}

func keepLegendaryTweetsV1(u *model.User, tts []model.TweetSimple) ([]model.TweetSimple, error) {
	if !u.Options.KeepLegendaryTweetV1Enable {
		return tts, nil
	}

	var ret []model.TweetSimple

	for _, v := range tts {
		if !v.Retweeted && v.FavoriteCount >= u.Options.KeepLegendaryTweetV1Border {
			continue
		}
		ret = append(ret, v)
	}
	return ret, nil
}
