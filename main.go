package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

/*
IMPROVEMENTS TO DO:
a. to make scroll automatically to end of community to gather all posts
*/
import (
	"strings"
)

type Community struct {
	URL   string
	Posts []Post
}

// Post defines the structure for a blog post with nested comments.
type Post struct {
	URL            string // URL of the post
	Name           string // Title of the post
	AuthorUsername string // Username of the post's author
	Sentiment      string
	Content        string
	Comments       []Comment // Slice of comments
}

// Comment defines the structure for a comment with potential sub-comments.
type Comment struct {
	AuthorUsername string // Username of the comment's author
	Content        string // Text content of the comment
	ThingId        string
	ParentId       string
	Sentiment      string
	Depth          int
	SubComments    []Comment
}

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Function to analyze sentiment using OpenAI's GPT-4.0 API
func analyzeSentiment(text string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY") // Using the key directly for clarity here
	if apiKey == "" {
		log.Println("API key not found in environment variables")
		return "", fmt.Errorf("API key is missing")
	}

	url := "https://api.openai.com/v1/chat/completions" // Updated endpoint for chat-based responses
	promptText := fmt.Sprintf("Analyze the sentiment of the following text and respond with one word: %s. Possible sentiments: sad, happy, angry, excited, supportive, discouraging, approving, negation.", text)
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-4", // Ensure you're specifying the correct model for chat
		"messages": []map[string]string{
			{"role": "system", "content": "You are a sentiment analysis assistant."},
			{"role": "user", "content": promptText},
		},
	})
	if err != nil {
		log.Printf("Error marshaling request: %v\n", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error during API call: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	var response struct { // Redefine based on the specific structure of the chat API response
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Error decoding response: %v\n", err)
		return "", err
	}

	if len(response.Choices) > 0 && len(response.Choices[0].Message.Content) > 0 {
		return response.Choices[0].Message.Content, nil
	}
	log.Println("No response from API")
	return "", fmt.Errorf("no response from API")
}

func main() {
	// Disable headless mode to see the browser UI
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.WindowSize(6920, 4080),
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 5000*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.reddit.com/login/"),
		chromedp.Sleep(3*time.Second),
		chromedp.WaitReady(`#login-username`, chromedp.ByID),
		chromedp.SendKeys(`#login-username`, "ReadingDangerous6771", chromedp.ByID),
		chromedp.WaitReady(`#login-password`, chromedp.ByID),
		chromedp.SendKeys(`#login-password`, "Dsoft12345", chromedp.ByID),
		chromedp.SendKeys(`#login-password`, "\n", chromedp.ByID),
	)

	if err != nil {
		log.Fatalf("Failed to run chromedp: %v", err)
	}
	time.Sleep(5 * time.Second)
	log.Println("Login attempt completed")
	navigateCommunities(ctx) // Assuming you want to call this after login
}

func navigateCommunities(ctx context.Context) {
	var numberOfPagesOfCommunitiesString string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.reddit.com/best/communities/1"),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate("document.querySelector('shreddit-ud-pagination').shadowRoot.querySelector('icon-caret-down').click()", nil),
		chromedp.WaitVisible("#directory-pagination .page-number:last-of-type", chromedp.ByQuery),
		chromedp.Text("#directory-pagination .page-number:last-of-type", &numberOfPagesOfCommunitiesString, chromedp.ByQuery),
	)
	println("NUMBER OF PAGES: " + numberOfPagesOfCommunitiesString)
	if err != nil {
		log.Fatalf("Failed to navigate and extract text: %v", err)
	}

	numPages, err := strconv.Atoi(strings.ReplaceAll(numberOfPagesOfCommunitiesString, "...", ""))
	if err != nil {
		log.Fatalf("Failed to convert string to integer: %v", err)
	}

	var communities []Community
	for i := 1; i <= numPages; i++ {
		communities_array, err := retrieveURLs(ctx, i)
		if err != nil {
			log.Printf("Failed to retrieve URLs for page %d: %v", i, err)
		}
		communities = append(communities, communities_array...)
	}

	writePostsToFile(communities)

}

// retrieveURLs retrieves href attributes from all anchor tags within elements with the class .items-start.
func retrieveURLs(ctx context.Context, pageNumber int) ([]Community, error) {
	println(fmt.Sprintf("Trying to retrieve url for page %d", pageNumber))
	var urls []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf("https://www.reddit.com/best/communities/%d", pageNumber)),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('.items-start a')).map(a => a.href)`, &urls),
	)
	if err != nil {
		return nil, err
	}
	var communities []Community

	for _, url := range urls {
		community_obj, _ := navigateCommunityPosts(ctx, url)
		communities = append(communities, community_obj)
	}

	return communities, nil
}

func navigateCommunityPosts(ctx context.Context, communityURL string) (Community, error) {
	var posts_object []string
	var posts []Post
	var community Community
	err := chromedp.Run(ctx,
		chromedp.Navigate(communityURL),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate("Array.from(document.querySelectorAll('a[slot=\"full-post-link\"]')).map(a => a.href)", &posts_object),
	)

	if err != nil {
		println("ERROR NAVIGATE COMMUNITY POSTS:")
		fmt.Println(err)
	}

	community.URL = communityURL
	for _, post_url := range posts_object {
		post, _ := navigateCommunityPost(ctx, post_url)
		posts = append(posts, post)
	}
	community.Posts = posts

	return community, nil
}

func navigateCommunityPost(ctx context.Context, postURL string) (Post, error) {
	var post Post
	var comments_authors []string
	var comments_text []string
	var comments_thing_id []string
	var comments_parent_id []string
	var comments_depth []string
	post.URL = postURL

	// Scroll until no more new content is loaded
	var lastHeight, currentHeight int64 = 0, 0
	for {
		err := chromedp.Run(ctx,
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight); document.body.scrollHeight;`, &currentHeight),
		)
		if err != nil {
			return post, fmt.Errorf("failed during scroll evaluation: %v", err)
		}

		fmt.Printf("Current Height: %d, Last Height: %d\n", currentHeight, lastHeight)

		if lastHeight == currentHeight {
			break // Break if the scroll did not increase the page height
		}
		lastHeight = currentHeight

		time.Sleep(2 * time.Second) // Sleep to allow for lazy-loaded content to load
	}

	// Navigate to the post URL
	if err := chromedp.Run(ctx, chromedp.Navigate(postURL)); err != nil {
		log.Printf("Error navigating to post: %v", err)
	}

	// Execute a series of tasks, continuing even if some fail
	tasks := chromedp.Tasks{
		chromedp.Text("h1[slot=title]", &post.Name, chromedp.NodeVisible, chromedp.AtLeast(0)),
		chromedp.Text("div[data-post-click-location=\"text-body\"]", &post.Content, chromedp.NodeVisible, chromedp.AtLeast(0)),
		chromedp.Text(".author-name:first-of-type", &post.AuthorUsername, chromedp.NodeVisible, chromedp.AtLeast(0)),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		log.Printf("Error fetching post details: %v", err)
	}

	// Extract comments
	commentTasks := chromedp.Tasks{
		chromedp.Evaluate(`Array.from(document.querySelectorAll('shreddit-comment')).map(a => a.getAttribute('thingid'))`, &comments_thing_id),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('shreddit-comment')).map(a => a.getAttribute('parentid'))`, &comments_parent_id),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('shreddit-comment')).map(a => a.getAttribute('depth'))`, &comments_depth),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('faceplate-tracker[noun=comment_author]')).map(a => a.innerText)`, &comments_authors),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('div[slot=comment]')).map(a => a.innerText)`, &comments_text),
	}

	if err := chromedp.Run(ctx, commentTasks); err != nil {
		log.Printf("Error fetching comments: %v", err)
	}

	// Build comments hierarchy and analyze sentiment if there is content
	post.Comments = buildCommentsHierarchy(comments_authors, comments_text, comments_thing_id, comments_parent_id, comments_depth)
	if post.Content != "" || post.Name != "" {
		if sentiment, err := analyzeSentiment("Content:" + post.Content + " Title:" + post.Name); err == nil {
			post.Sentiment = sentiment
		} else {
			log.Printf("Error analyzing sentiment: %v", err)
		}
	}

	return post, nil
}

func buildCommentsHierarchy(authors, texts, thingIDs, parentIDs, depths []string) []Comment {
	commentsMap := make(map[string]*Comment)
	var sentiment string
	for i := range authors {
		depth, _ := strconv.Atoi(depths[i])

		comment := &Comment{
			AuthorUsername: authors[i],
			Content:        texts[i],
			ThingId:        thingIDs[i],
			ParentId:       parentIDs[i],
			Depth:          depth,
			SubComments:    []Comment{},
		}
		sentiment, _ = analyzeSentiment(comment.Content)
		comment.Sentiment = sentiment
		commentsMap[thingIDs[i]] = comment
	}

	var rootComments []Comment

	for _, comment := range commentsMap {
		if comment.ParentId == "" || comment.Depth == 0 {
			rootComments = append(rootComments, *comment)
		} else {
			parent, exists := commentsMap[comment.ParentId]
			if exists {
				parent.SubComments = append(parent.SubComments, *comment)
			}
		}
	}

	return rootComments
}

func writePostsToFile(communities []Community) error {
	file, err := os.Create("post.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(communities)
}
