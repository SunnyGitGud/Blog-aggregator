## 🛠 Project: CLI RSS Aggregator

**Features:**

- User authentication (`login`, `register`, `whoami`)  
- Feed management (`addfeed`, `feeds`, `follow`, `following`)  
- Automatic RSS fetching with a long-running loop (`agg` command)  
- Post storage with metadata (`posts` table)  
- Browse posts per user with limits (`browse` command)  

**Technologies Used:**

- Go (CLI & backend logic)  
- PostgreSQL (database & SQL queries with `sqlc`)  
- UUIDs, timestamps, and SQL nullable types  
- RSS parsing (using XML structs in Go)  
- Go standard library: `time`, `context`, `log`, `database/sql`  
- Migrations using `goose`

## ⚡ Key Challenges & Solutions

| Problem                                                      | Solution                                                     |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| Handling SQL timestamps (`created_at`, `updated_at`)         | Learned to use `time.Now()` and `sql.NullTime` for nullable fields |
| Complex SQL queries with joins (`feed_follow`, `posts`, `users`) | Used `JOIN`, `INNER JOIN`, and `LEFT JOIN` correctly; `sqlc` generated Go structs helped a lot |
| Avoiding duplicate records (feeds & posts)                   | Added unique constraints and handled duplicate key errors gracefully |
| Continuous RSS fetching without DOS-ing servers              | Implemented `time.Ticker` and `scrapeFeeds` loop with delays |
| Handling various RSS feed date formats                       | Used `time.Parse` with multiple layouts, stored as `sql.NullTime` |
| Printing readable feed descriptions                          | Initially printed raw HTML, later learned about HTML stripping for clean console output |
| Logged-in middleware for DRY code                            | Created a higher-order function to automatically inject the current user into handlers |
| Testing CLI commands locally                                 | Used real RSS feeds (TechCrunch, Hacker News, Boot.dev) to simulate aggregation |

## 📂 Project Structure

```tree
blog/
├─ internal/
│  ├─ database/          # generated structs (sqlc)
│  ├─ config/            # Config parsing
├─ sql/
│  ├─ queries/           # SQL queries (sqlc)
│  ├─ schema/            # goose migrations for tables: users, feeds, posts, feed_follows
├─ main.go               # CLI entry point
├─ handlers.go           # handler functions
├─ middleware.go         # dry functions
├─ scrapeFeeds.go        # feed scraping and parsing


```
## 🚀 How it Works

1. **Add a feed**:  
   ```bash
   gator addfeed "Hacker News" "https://news.ycombinator.com/rss"
   ```

2. **Follow a feed**:  
   ```bash
   gator follow "https://news.ycombinator.com/rss"
   ```

3. **Run aggregator loop**:  
   ```bash
   gator agg 1m
   ```

- Collects feeds every 1 minute  
- Saves posts to the database  
- Prints feed post titles  

4. **Browse posts**:  
   ```bash
   gator browse 5
   ```

- Shows the 5 most recent posts for the logged-in user  

## 📖 Learning Highlights

- Deep dive into **Go + PostgreSQL integration**  
- Learned **higher-order functions** for middleware  
- Handled **real-world RSS feed quirks** like missing fields and inconsistent timestamps  
- Learned **SQL best practices** for joins, constraints, and ordering  
- Built a **long-running CLI service** with proper error handling  

## 🔗 Example RSS Feeds I Tested

- [techcrunch](https://techcrunch.com/feed/)  
- [Hacker News](https://news.ycombinator.com/rss)  
- [Boot.dev Blog](https://blog.boot.dev/index.xml)  

## 📌 Future Enhancements

- HTML parsing & sanitization for post descriptions  
- CLI search/filter commands for posts  
- Notifications for new posts  
- Web interface for viewing feeds  

## 📝 Summary

This project was more than just a Go CLI app — it was a **hands-on journey in backend development, SQL, and data processing**. Every bug and error was a learning opportunity, from handling nullable timestamps to continuous feed aggregation.  

I now have a strong foundation for building **robust, database-backed CLI applications** and scraping pipelines in Go.
