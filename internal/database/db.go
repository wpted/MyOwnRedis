package database

type MemDb interface {
    Set(key string, value string)
    Get(key string) (string, error)
    Exists(key string) bool
    Delete(keys ...string) int
    Increment(key string) (int, error)
    Decrement(key string) (int, error)
    LeftPush(key string, values ...string) (int, error)
    RightPush(key string, values ...string) (int, error)
    LRange(key string, start, stop int) ([]string, error)
    SaveDatabase() error
}
