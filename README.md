# unfire

# TODO(やる順)
- 昨日(安定を取って二日前？)ツイートしたものだけ削除
  - これが出来たらdeployして稼働させる
- リファクタ
- いいねの削除
- 毎回認証させる？(そもそもセッション(Cookie)にアクセストークンおくのおkなん？__)

- running as direct

`root/scripts/manager.sh run direct`

- running with docker

`root/scripts/manager.sh run docker`

- health

https://unfire.herokuapp.com/api/v1/health

# 参考



- [Goで書いたサーバーをHerokuにDocker Deployする - Qiita](https://qiita.com/croquette0212/items/2b85aa2c6b2933244f07)
- [Heroku Dockerの使い所](https://www.slideshare.net/kon_yu/heroku-docker)