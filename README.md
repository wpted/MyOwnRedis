# Write Your Own Redis Server

This is the solution for [Redis Challenge](https://codingchallenges.fyi/challenges/challenge-redis)
implemented using Go.

## About

Redis stands for Remote Dictionary Server. Similar to byte arrays, Redis strings store sequences of bytes, including
text, serialized objects, counter values and binary arrays.

## Idea

1. Be a copy cat. Translate the first version of Redis ( written in TCL ) to Go. Understand how redis is designed.
2. Initiate a tcp server that receives data frames from any client.
3. Decode the client payload using RESP, return error message if payload is not a valid RESP.
4. Depending on the command, check which corresponding action to perform.
5. Send data back to the client.

## Implementation Steps

1. Build functionality to serialise and de-serialise messages. The Redis server should follow the **Redis Serialisation
   Protocol (RESP)**.<br><br>
2. Create a **Light Memory-Mapped Database(LMDB)** server that listens on port <ins>6379</ins>, which is usually
   implemented as an embedded transactional key-value database. The connection uses TCP.<br><br>
3. Implement the core functions of Redis.
4. Test with the official Redis Client

### TODO

- [x] Concurrent CRUD.
- [x] Check server status ( **PING** )
- [x] Store and retrieve data ( **SET** and **GET** )
- [x] Altering and deleting data ( **SET** and **DEL** )
- [x] Incrementing and decrementing stored number ( **INCR** amd **DECR** )
- [x] Insert all the values and the head ( **LPUSH** ) or tail(**RPUSH**) of a list.
- [x] Show stored values in a list ( **LRANGE** )
- [x] Check whether a data exists ( **EXISTS** )
- [x] Set key expiration ( **EX**, **PX**, **EXAT** and **PXAT**)
- [x] Scan **keyspace** to get a list of keys ( **SCAN** )
- [x] Save the database state to disk. ( **SAVE** )
  <br><br>

## Program

Git pull the repo and run the program

```bash
    go run cmd/main.go
```

or compile it first then run the corresponding executable on the different platforms.

```bash
  # On Linux or MacOS
  go build -o rredis -v ./cmd/main.go
  # On Darwin MacOS
  env GOOS=darwin GOARCH=amd64 go build -o rredis -v ./cmd/main.go
  ./rredis
  
  # On Windows
  env GOOS=windows GOARCH=amd64 go build -o rredis.exe -v ./cmd/main.go
  rredis.exe
```

You can also run the existing binaries pulled directly from the repo.

```bash
# On Linux

./rredis

# On Windows
rredis.exe
```

After starting the program you'll see:
```text
RRedis listening on port 6379...
```

After starting the rredis server, connect it with a client that send valid RESP request.
From the Redis client, request **ping** should respond with a **PONG**.
```redis
127.0.0.1:6379 > ping
PONG
```

Try echo a cliché 'hello world'
```redis
127.0.0.1:6379 > echo hello world
hello world
```

## Commands
This section is highly inspired by [The Redis Command Page](https://redis.io/commands/).

- **PING**
    - Returns PONG. 
    This command is useful for testing whether a connection is healthy.
```redis
    127.0.0.1:6379 > PING
    PONG
```
- **ECHO**
    - Returns message.
    
```redis
    // Example
    127.0.0.1:6379 > echo hello world
    hello world
```
- **SET**
  - Set key to hold the string value. If key already holds a value, it is overwritten, regardless of its type. Any previous time to live associated with the key is discarded on successful SET operation.
```text
    // Syntax
  
    SET key value [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds]
```

```redis
    127.0.0.1:6379 > SET x 1
    OK
    127.0.0.1:6379 > GET x
    1
  
    // SET x with TTL
    127.0.0.1:6379 > SET x 1 EX 60
    OK
```
- **GET** 
  - Get the value of key. If the key does not exist the special value nil is returned. An error is returned if the value stored at key is not a string, because GET only handles string values.
```text
    // Syntax
    GET key
```

```redis
    127.0.0.1:6379 > SET x 1
    OK
    127.0.0.1:6379 > GET x
    1
```
- **INCR**
  - Increments the number stored at key by one. If the key does not exist, it is set to 0 before performing the operation. An error is returned if the key contains a value of the wrong type or contains a string that can not be represented as integer. This command is a string operation.
```text
    // Syntax
    INCR key
```

```redis
    127.0.0.1:6379 > SET x 1
    OK
    127.0.0.1:6379 > INCR x
    2
```
- **DECR**
    - Decrements the number stored at key by one. If the key does not exist, it is set to 0 before performing the operation. An error is returned if the key contains a value of the wrong type or contains a string that can not be represented as integer. This command is a string operation.
```text
    // Syntax
    INCR key
```

```redis
    127.0.0.1:6379 > SET x 1
    OK
    127.0.0.1:6379 > INCR x
    2
```
- **EXISTS**
  - Returns if key exists. We return the number of keys that exist from those specified as arguments. 
```text
    // Syntax
    INCR key
```

```redis
    127.0.0.1:6379 > SET x 1
    OK
    127.0.0.1:6379 > EXISTS x
    (integer) 1
    127.0.0.1:6379 > EXISTS keyThatDoesntExist
    (integer) 0
```
- **DEL**
  - Removes the specified keys and return the number of keys that were removed. A key is ignored if it does not exist.
```text
    // Syntax
    DEL key [key ...]
```

```redis
    127.0.0.1:6379 > SET key1 "Hello"
    OK
    127.0.0.1:6379 > SET key2 "World"
    OK
    127.0.0.1:6379 > del key1 key2
    (integer) 2
```
- **LPUSH**
  - Insert all the specified values at the head of the list stored at key and return the length of the list after the push operation. If key does not exist, it is created as empty list before performing the push operations. When key holds a value that is not a list, an error is returned.
```text
    // Syntax
    LPUSH key element [element ...]
```

```redis
    127.0.0.1:6379 > LPUSH key1 "Hello"
    (integer) 1
    127.0.0.1:6379 > LPUSH key1 "World"
    (integer) 2
    127.0.0.1:6379 > LRANGE key1 0 1
    (1) "Hello"
    (2) "World"
```
- **RPUSH**
  - Insert all the specified values at the tail of the list stored at key and return the length of the list after the push operation. If key does not exist, it is created as empty list before performing the push operations. When key holds a value that is not a list, an error is returned.
    
```text
    // Syntax
    RPUSH key element [element ...]
```

```redis
    127.0.0.1:6379 > RPUSH key1 "Hello"
    (integer) 1
    127.0.0.1:6379 > RPUSH key1 "World"
    (integer) 2
    127.0.0.1:6379 > LRANGE key1 0 1
    (1) "World"
    (2) "Hello"
```

- **LRANGE** 
  - Returns the specified elements of the list stored at key. The offsets start and stop are zero-based indexes, with 0 being the first element of the list (the head of the list), 1 being the next element and so on. These offsets can also be negative numbers indicating offsets starting at the end of the list. For example, -1 is the last element of the list, -2 the penultimate, and so on. Out of range indexes will not produce an error. If start is larger than the end of the list, an empty list is returned. If stop is larger than the actual end of the list, Redis will treat it like the last element of the list.

```text
    // Syntax
    LRANGE key start stop
```

```redis
    127.0.0.1:6379 > RPUSH key1 "Hello"
    (integer) 1
    127.0.0.1:6379 > RPUSH key1 "World"
    (integer) 2
    127.0.0.1:6379 > LRANGE key1 0 1
    (1) "World"
    (2) "Hello"
```
- **SCAN**
  - Save the DB for all existing keys. This command works different from the original Redis. 

```text
    // Syntax
    SCAN
```

```redis
    127.0.0.1:6379 > RPUSH key1 "Hello"
    (integer) 1
    127.0.0.1:6379 > SET x 1
    OK
    127.0.0.1:6379 > SCAN
    (1) "key1"
    (2) "x"
```

- **SAVE**
  - Save the DB in background and OK code is immediately returned. Save performs a snapshot and store it inside a csv file within the same folder of the program. The cycleTime and keysChanged should be positive numbers.

```text
    // Syntax
    SAVE [cycleTime(seconds) keysChanged]
```

```redis
    127.0.0.1:6379 > RPUSH key1 "Hello"
    (integer) 1
    127.0.0.1:6379 > RPUSH key1 "World"
    (integer) 2
    127.0.0.1:6379 > LRANGE key1 0 1
    (1) "World"
    (2) "Hello"
```


### Data Persistence
Unlike **Redis** persist data with AOF and RDB files, the current version of my Redis
## Supported data types

|             | First Byte | 
|:------------|:----------:|
| Strings     |     +      |
| Errors      |     -      |
| Integers    |     :      |
| Arrays      |     *      |
| BulkStrings |     $      |

## References
### Reads:

- [LMDB -- First version of Redis written in Tcl](https://gist.github.com/antirez/6ca04dd191bdb82aad9fb241013e88a8)
- [Reference counting: Harder than it sound](https://www.playingwithpointers.com/blog/refcounting-harder-than-it-sounds.html)
- [Introduction to Redis](https://redis.io/docs/about/)
- [RESP protocol spec](https://redis.io/docs/reference/protocol-spec/)
- [理解 Redis 的 RESP 協議](https://moelove.info/2017/03/05/理解-Redis-的-RESP-协议/)
- [How to implement instant messaging with WebSockets in Go](https://yalantis.com/blog/how-to-build-websockets-in-go/)
- [Redis file persistence](https://redis.io/docs/management/persistence/)
- [How to test TCP/UDP connection in Go](https://dev.to/williamhgough/how-to-test-tcpudp-connections-in-go---part-1-3bga)
- [Understanding the Redis protocol](https://subscription.packtpub.com/book/data/9781783988167/1/ch01lvl1sec17/understanding-the-redis-protocol)
- [Introduction to RESP](https://medium.com/@dassomnath/introduction-to-resp-redis-serialization-protocol-f3d0b8bd9cdc)
- [Data Fetching Patterns](https://nextjs.org/docs/app/building-your-application/data-fetching/patterns)
- [The Complete Guide to TCP/IP Connections in Golang](https://okanexe.medium.com/the-complete-guide-to-tcp-ip-connections-in-golang-1216dae27b5a)
- [Redis Client Handling](https://redis.io/docs/reference/clients/)
- [How to test TCP/UDP connections in Go](https://dev.to/williamhgough/how-to-test-tcpudp-connections-in-go---part-1-3bga)