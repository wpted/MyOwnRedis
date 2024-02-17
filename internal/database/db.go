package database

type MemDb interface {
    Set(key string, value string)
    Get(key string) (string, error)
    Exists(key string) bool
    Delete(keys ...string) int
    Increment(key string) (int, error)
    Decrement(key string) error
    LeftPush(key string, values ...string) (int, error)
    RightPush(key string, values ...string) (int, error)
    SaveDatabase() error
    LoadDatabase() error
}
