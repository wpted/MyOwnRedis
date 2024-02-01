package database

type MemDb interface {
    Set(key string, value string)
    Get(key string) (string, error)
    Exists(key string) bool
    Delete(key string)
    Increment(key string) error
    Decrement(key string) error
    LeftPush(key string, values ...string) (int, error)
    RightPush(key string, values ...string) (int, error)
    SaveDatabase() error
    LoadDatabase() error
}
