### Reddit Community Scraper with Sentiment Analysis

#### Project Overview
This project, developed by Burcovschi David, is a Reddit community scraper built with Go using Chromedp for web scraping and OpenAI's GPT-4 for sentiment analysis. It navigates through the best communities on Reddit, extracts posts, analyzes sentiments of post contents and comments, and saves the structured data to a JSON file.

#### Features
- Automatically scrolls to the end of posts and community lists to capture all available content.
- Performs sentiment analysis on the posts and comments using OpenAI's GPT-4 model.
- Efficient error handling to ensure robust execution.
- Uses environment variables to securely manage the OpenAI API key.

#### Getting Started

##### Prerequisites
- Go (version 1.15 or newer recommended)
- Chromedp for running headless browser tasks
- An OpenAI API key

##### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/burcovschidavid/sentiment-analisys-golang-reddit
   cd sentiment-analisys-golang-reddit
   ```

2. **Install Dependencies**

   Install the required Go packages including `chromedp` for web scraping and `godotenv` for managing environment variables:

   ```bash
   go get github.com/chromedp/chromedp
   go get github.com/joho/godotenv
   ```

3. **Set Up Environment Variables**

   Create a `.env` file in the root of your project directory and add your OpenAI API key:

   ```plaintext
   OPENAI_API_KEY=your_openai_api_key_here
   ```

   Replace `your_openai_api_key_here` with your actual OpenAI API key.

4. **Build the Project**

   ```bash
   go build -o redditScraper
   ```

5. **Run the Project**

   ```bash
   ./redditScraper
   ```

##### Configuration
Ensure your `.env` file is correctly set up as described in the installation steps. Adjust the Chromedp settings in the script if necessary to match your specific requirements or debugging preferences.

#### Usage
Run the built executable to start scraping Reddit communities. The script will log into Reddit, navigate through community pages, and perform sentiment analysis on each post and its comments. Results will be saved in a file named `post.json` in the project directory.

#### Contributing
Contributions to this project are welcome. Please ensure to update tests as appropriate and adhere to the existing coding style.

#### License
This project is licensed under the MIT License - see the LICENSE file for details.

#### Support
For support, email [burcovschidavid@gmail.com](mailto:burcovschidavid@gmail.com) or raise an issue in the project's GitHub repository at [https://github.com/burcovschidavid/sentiment-analisys-golang-reddit](https://github.com/burcovschidavid/sentiment-analisys-golang-reddit).

 