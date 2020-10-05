# unfire

## Options

| パラメータ                              | タイプ     | デフォルト | 内容                                                                                                   | 補足                                                               |
| ---------------------------------- | ------- | ----- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| delete\_like                       | booleam | FALSE | いいねを削除するかどうか                                                                                         |                                                                  |
| delete\_like\_count                | number  | 30    | (delete\_likeがtrueの時使用)<br>何件以上になったらいいねを削除するか( 1以上1000未満で指定)                                         |                                                                  |
| keep\_legendary\_tweet\_v1\_enable | boolean | FALSE | バズったツイートを削除するかどうか                                                                                    |                                                                  |
| keep\_legendary\_tweet\_v1\_border | number  | 20000 | (keey\_legendary\_tweet\_v1\_countがtrueの時使用)<br>ここに指定された数以上のいいねがついたツイートは削除しない<br>(15以上10000000未満で指定) | 取ってきたツイートに対して、<br>filterしているだけなので 150件以上これに入ると<br>ツイートが削除されなくなる。 |

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
