package persistence

import "context"

// KeyValueストア
type Datastore interface {
	SetString(ctx context.Context, key string, value string) error
	GetString(ctx context.Context, key string) (string, error)
	// List出ない場合はエラー
	AppendString(ctx context.Context, key string, value string) error
	// 末尾の要素を削除して削除した要素を返す(Listでないならエラー)
	LastPop(ctx context.Context, key string) (string, error)
	// KeyのListの長さを返す(Listでないならエラー)
	ListLen(ctx context.Context, key string) (int64, error)
	// GetStringByIndex Stringをindex指定で返す。
	GetStringByIndex(ctx context.Context, key string, index int64) (string, error)
	// SortedSetにint64を追加する (score == value)
	InsertInt64(ctx context.Context, key string, value int64) error
	// SortedSetの最小値を取得する。
	GetMinElement(ctx context.Context, key string) (string, error)
	// PopMin: SortedSetの最小値を削除する。
	PopMin(ctx context.Context, key string) error
	// SortedSetに任意の値を追加する。
	Insert(ctx context.Context, key string, score float64, member interface{}) error
}
