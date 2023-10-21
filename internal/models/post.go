package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Post struct {
	ID       int
	Title    string
	Content  string
	Created  time.Time
	UserId   int
	Author   string
	Likes    int
	Dislikes int
	Categories []string
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type PostModel struct {
	DB *sql.DB
}

// This will insert a new User into the database.
func (p *PostModel) Insert(title string, content string, userId int) (int, error) {
	var author string
	err := p.DB.QueryRow("SELECT username FROM users WHERE user_id = ?", userId).Scan(&author)
	if err != nil {
		return 0, err
	}

	stmt := `INSERT INTO posts (title, content, created_date, user_id, author)
	VALUES (?, ?, datetime('now','localtime'), ?, ?)`
	result, err := p.DB.Exec(stmt, title, content, userId, author)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (p *PostModel) Get(id int) (*Post, error) {
	stmt := `SELECT post_id, title, content, created_date, user_id, author FROM posts
	WHERE post_id = ?`

	row := p.DB.QueryRow(stmt, id)

	post := &Post{}

	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Created, &post.UserId, &post.Author)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return post, nil
}

func (p *PostModel) GetUserPosts(userID int) ([]*Post, error) {
	stmt := `SELECT post_id, title, content, created_date, user_id  FROM posts WHERE user_id = ?
	ORDER BY post_id DESC `
	rows, err := p.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	posts := []*Post{}
	for rows.Next() {
		post := &Post{}

		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.Created, &post.UserId)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (p *PostModel) Latest() ([]*Post, error) {
	stmt := `SELECT post_id, title, content, created_date, user_id  FROM posts
	ORDER BY post_id DESC LIMIT 20`
	rows, err := p.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	posts := []*Post{}
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.Created, &post.UserId)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// getTagIDByName принимает имя категории и возвращает соответствующий идентификатор категории (tag_id) из базы данных.
func (p *PostModel) LatestWithCategory(categories []string) ([]*Post, error) {
	// Create the query with dynamic placeholders
	placeholders := make([]string, len(categories))
	for i := range categories {
		placeholders[i] = "?"
	}
	categoryPlaceholders := strings.Join(placeholders, ",")

	stmt := `SELECT post_id, title, content, created_date, user_id FROM posts
             WHERE post_id IN (
                 SELECT post_id FROM posts_categories
                 WHERE category_id IN (
                     SELECT category_id FROM categories
                     WHERE category IN (` + categoryPlaceholders + `)
                 )
             )
             ORDER BY post_id DESC`

	// Convert the []string slice to []interface{} slice
	args := make([]interface{}, len(categories))
	for i, v := range categories {
		args[i] = v
	}

	rows, err := p.DB.Query(stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*Post{}
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.Created, &post.UserId)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
