# Blogbackend

Project Structure:
- Home Page (index.html): Displays a list of published articles.

- Article Page (article.html): Displays full content of an article with its publication date.

Admin section
- Dashboard (admin.html): Lists all articles with options to edit, delete, or add a new one.

- Add Article Page (add.html): A form with fields: title, content, date of publication, and image.

- Edit Article Page (edit.html): Same form as the Add Article Page but pre-filled with existing article data.

# Backend (Go)
All articles and users are stored in a data.json file. Basic Authentication checks the JSON file for user credentials.

# Endpoints
Public API
- GET /articles → Retrieves all articles.
- GET /article/{id} → Retrieves a specific article.

Admin API
- POST /admin/add → Adds a new article.
- PUT /admin/edit/{id} → Edits an existing article.
- DELETE /admin/delete/{id} → Removes an article.

Authentication
- Basic Auth: Username & password stored in users.json.
- Middleware checks credentials before allowing access to admin endpoints.

# Frontend
It has three sections:
- Home Page: Fetch articles from /articles and display them.
- Article Page: Fetch article data from /article/{id}.
- Admin Pages: Use forms to post data to the backend.


