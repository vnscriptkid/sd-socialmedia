# Doubts
- How to scale feed table?
    - {user_id: [postToday, postYesterday, postWeekAgo, ...]}
    - As number of posts grow, will it be possible to store all posts in a list?
    - Maximum size of record in dynamodb?

## Redis list 
```sh
# Add to list
lpush feed:1 post1
lpush feed:1 post2

# Get list
lrange feed:1 0 -1
```