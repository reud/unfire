# unfire

ツイートして一定時間経ったツイートを削除するアプリケーション(オプションでいいねも削除出来るよ！)

![logo](https://user-images.githubusercontent.com/31179220/96333627-089c2b00-10a6-11eb-9e57-43aa0e9c14ee.jpg)

## usage

### portal.reud.net

https://portal.reud.net/unfire

### direct

ここにアクセス

https://unfire.reud.app/api/v1/auth/login

### redis

docker run --name redis -d -p 6379:6379 redis redis-server --appendonly yes

## Options

https://unfire.reud.app/api/v1/auth/login

URLパラメータを付加してオプションの設定が可能になります。

| パラメータ                              | タイプ     | デフォルト | 内容                                                                                                   | 補足                                                               |
| ---------------------------------- | ------- | ----- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| delete\_like                       | booleam | FALSE | いいねを削除するかどうか                                                                                         |                                                                  |
| delete\_like\_count                | number  | 30    | (delete\_likeがtrueの時使用)<br>何件以上になったらいいねを削除するか( 1以上1000以下で指定)                                         |                                                                  |
| keep\_legendary\_tweet\_v1\_enable | boolean | FALSE | バズったツイートを削除するかどうか                                                                                    |                                                                  |
| keep\_legendary\_tweet\_v1\_border | number  | 20000 | (keey\_legendary\_tweet\_v1\_countがtrueの時使用)<br>ここに指定された数以上のいいねがついたツイートは削除しない<br>(15以上10000000未満で指定) | 取ってきたツイートに対して、<br>filterしているだけなので 150件以上これに入ると<br>ツイートが削除されなくなる。 |
| callback\_url                      | string  | nil   | 実行完了時の遷移先
# TODO(やる順)
- 昨日(安定を取って二日前？)ツイートしたものだけ削除
  - これが出来たらdeployして稼働させる
- リファクタ
- いいねの削除
- 毎回認証させる？(そもそもセッション(Cookie)にアクセストークンおくのおkなん？__)

- running as direct

`<project-root>/scripts/manager.sh run direct`

- running with docker

`<project-root>/scripts/manager.sh run docker`

- health

http://unfire.reud.app/health

# 参考



- [Goで書いたサーバーをHerokuにDocker Deployする - Qiita](https://qiita.com/croquette0212/items/2b85aa2c6b2933244f07)
- [Heroku Dockerの使い所](https://www.slideshare.net/kon_yu/heroku-docker)
