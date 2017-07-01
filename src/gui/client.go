package gui

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	ErrTimeout = errors.New("timeout")
)

type (
	// ClientFunc is a client function.
	ClientFunc func(ctx context.Context, port int) ([]byte, error)

	// Values represents key/value pairs.
	Values map[string]*string
)

/*
	<<< FOR BOARDS, THREADS & POSTS >>>
*/

// GetBoards obtains boards in which the bbsnode is subscribed.
func GetBoards() ClientFunc {
	return gen("get_boards", nil)
}

// NewBoard creates a new board.
func NewBoard(boardName, boardDescription, boardSubmissionAddresses, seed *string) ClientFunc {
	return gen("new_board", Values{
		"name":                 boardName,
		"description":          boardDescription,
		"submission_addresses": boardSubmissionAddresses,
		"seed":                 seed,
	})
}

// RemoveBoard removes a board.
func RemoveBoard(board *string) ClientFunc {
	return gen("remove_board", Values{
		"board": board,
	})
}

// GetBoardPage obtains the board page of specified board of public key.
func GetBoardPage(board *string) ClientFunc {
	return gen("get_boardpage", Values{
		"board": board,
	})
}

// GetThreads obtains threads of a specified board of public key.
func GetThreads(board *string) ClientFunc {
	return gen("get_threads", Values{
		"board": board,
	})
}

// NewThread creates a new thread on specified board.
func NewThread(board, threadName, threadDescription *string) ClientFunc {
	return gen("new_thread", Values{
		"board":       board,
		"name":        threadName,
		"description": threadDescription,
	})
}

// RemoveThread removes a thread on specified board.
func RemoveThread(board, thread *string) ClientFunc {
	return gen("remove_thread", Values{
		"board":  board,
		"thread": thread,
	})
}

// GetThreadPage obtains a thread page of specified board and thread.
func GetThreadPage(board, thread *string) ClientFunc {
	return gen("get_threadpage", Values{
		"board":  board,
		"thread": thread,
	})
}

// GetPosts obtains the posts of a thread of specified board and thread.
func GetPosts(board, thread *string) ClientFunc {
	return gen("get_posts", Values{
		"board":  board,
		"thread": thread,
	})
}

// NewPost creates a new post on specified board and thread.
func NewPost(board, thread, postTitle, postBody *string) ClientFunc {
	return gen("new_post", Values{
		"board":  board,
		"thread": thread,
		"title":  postTitle,
		"body":   postBody,
	})
}

// RemovePost removes a post in specified board, thread and post reference.
func RemovePost(board, thread, post *string) ClientFunc {
	return gen("remove_post", Values{
		"board":  board,
		"thread": thread,
		"post":   post,
	})
}

// ImportThread imports a thread from a board to another.
func ImportThread(fromBoard, thread, toBoard *string) ClientFunc {
	return gen("import_thread", Values{
		"from_board": fromBoard,
		"thread":     thread,
		"to_board":   toBoard,
	})
}

/*
	<<< HELPER FUNCTIONS >>>
*/

// Asynchronously requests from api.
func request(port int, path string, data url.Values) (chan []byte, chan error) {
	bChan, eChan := make(chan []byte), make(chan error)
	go func() {
		resp, e := http.PostForm(
			fmt.Sprintf("http://127.0.0.1:%d/api/%s", port, path),
			data,
		)
		if e != nil {
			eChan <- e
			return
		}
		defer resp.Body.Close()
		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			eChan <- e
			return
		}
		bChan <- body
		return
	}()
	return bChan, eChan
}

// Generates a method of requesting data from api.
func gen(path string, values Values) ClientFunc {
	return func(ctx context.Context, port int) ([]byte, error) {
		// Get values.
		urlValues := url.Values{}
		for k, v := range values {
			urlValues[k] = []string{*v}
		}
		bChan, eChan := request(port, path, urlValues)
		select {
		case <-ctx.Done():
			return nil, ErrTimeout
		case e := <-eChan:
			return nil, e
		case body := <-bChan:
			return body, nil
		}
	}
}