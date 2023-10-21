package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Comment struct {
	CommentID int
	UserID    int
	Username  string
	PostID    int
	Text      string
	CreatedAt time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type CommentModel struct {
	DB *sql.DB
}

func (m *CommentModel) Insert(user_id, post_id int, comment string) (int, error) {
	stmt := `INSERT INTO comments (user_id, post_id, comment, created_at)
	VALUES (?, ?, ?, datetime('now','localtime'))`
	result, err := m.DB.Exec(stmt, user_id, post_id, comment)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	// Use the LastInsertId() method on the result to get the ID of our
	// newly inserted record in the snippets table.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// The ID returned has the type int64, so we convert it to an int type
	// before returning.
	return int(id), nil
}

func (m *CommentModel) GetComments(post_id int) ([]*Comment, error) {
	stmt := `SELECT c.comment_id, c.user_id, u.username, c.comment, c.created_at
             FROM comments c
             JOIN users u ON c.user_id = u.user_id 
             WHERE c.post_id = $1`
	rows, err := m.DB.Query(stmt, post_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*Comment{}

	for rows.Next() {
		c := &Comment{}
		err = rows.Scan(&c.CommentID, &c.UserID, &c.Username, &c.Text, &c.CreatedAt)
		if err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
