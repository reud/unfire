package hook

import "unfire/model"

// PreRunTaskHook ユーザのアクセストークンが取得される前に実行する。
func PreRunTaskHook(_ *model.User) error {
	return nil
}
