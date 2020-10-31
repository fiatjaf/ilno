package database

var (
	presetSQLITE3 map[string]string = map[string]string{
		"create": `
		CREATE TABLE IF NOT EXISTS comments (
			tid REFERENCES threads(id), 
			id INTEGER PRIMARY KEY, 
			parent INTEGER,
			created FLOAT NOT NULL,
			modified FLOAT,
			mode INTEGER,
			text VARCHAR,
            key VARCHAR,
			author VARCHAR,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			voters BLOB NOT NULL
		);
		CREATE TABLE IF NOT EXISTS preferences (
			key VARCHAR PRIMARY KEY, 
			value VARCHAR
		);
		CREATE TABLE IF NOT EXISTS threads (
			id INTEGER PRIMARY KEY,
			uri VARCHAR(256) UNIQUE,
			title VARCHAR(256)
		);
		CREATE TRIGGER IF NOT EXISTS remove_stale_threads
    	AFTER DELETE ON comments
    	BEGIN
    		DELETE FROM threads WHERE id NOT IN (SELECT tid FROM comments);
    	END;
		`,

		"preference_get": `SELECT value FROM preferences WHERE key=$1;`,
		"preference_set": `INSERT INTO preferences (key, value) VALUES ($1, $2);`,

		"thread_get_by_uri": `SELECT * FROM threads WHERE uri=$1;`,
		"thread_get_by_id":  `SELECT * FROM threads WHERE id=$1;`,
		"thread_new":        `INSERT INTO threads (uri, title) VALUES ($1, $2);`,

		"comment_new": `INSERT INTO comments (
        	tid, parent, created, modified, mode,
			text, key, author, voters
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`,
		"comment_get_by_id": `SELECT * FROM comments WHERE id=$1`,
		"comment_is_previously_approved_author": `SELECT CASE WHEN EXISTS(
			SELECT * FROM comments WHERE email=$1 AND mode=1 AND created > strftime("%s", DATETIME("now", "-6 month"))
		) THEN 1 ELSE 0 END;`,
		"comment_count_reply": `SELECT comments.parent,count(*)
			FROM comments INNER JOIN threads ON threads.uri=$1 AND comments.tid=threads.id AND ($2 | comments.mode = $2) GROUP BY comments.parent`,
		"comment_fetch_by_uri": `SELECT comments.* FROM comments INNER JOIN threads ON
			threads.uri=? AND comments.tid=threads.id AND (? | comments.mode) = ?`,
		"comment_count": `SELECT threads.uri, COUNT(comments.id) FROM comments LEFT OUTER JOIN 
		threads ON threads.id = tid AND comments.mode = 1 GROUP BY threads.uri`,
		"comment_activate":     `UPDATE comments SET mode=1 WHERE id=$1 AND mode=2;`,
		"comment_edit":         `UPDATE comments SET text=$1, author=$2, modified=$3 WHERE id=$4`,
		"comment_delete_check": `SELECT COUNT(*) FROM comments WHERE parent=?`,
		"comment_delete_hard":  `DELETE FROM comments WHERE id=?`,
		"comment_delete_soft":  `UPDATE comments SET mode=4, text='', author='', key='' WHERE id=?`,
		"comment_delete_stale": `DELETE FROM comments 
		WHERE mode=4 AND id NOT IN (SELECT parent FROM comments WHERE parent IS NOT NULL)`,
		"comment_vote_set": `UPDATE comments SET likes=?, dislikes=?, voters=? WHERE id=?`,
	}
)

var presetSQL map[string]map[string]string = map[string]map[string]string{
	"sqlite3": presetSQLITE3,
}
