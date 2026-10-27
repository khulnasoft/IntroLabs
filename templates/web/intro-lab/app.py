#!/usr/bin/env python3
"""
Introduction to Web Security Lab
Demonstrates XSS and IDOR vulnerabilities for educational purposes.
"""

from flask import Flask, request, render_template_string, jsonify
import os

app = Flask(__name__)

# In-memory data store (for IDOR demonstration)
users = {
    1: {"name": "Alice", "email": "alice@example.com", "role": "user"},
    2: {"name": "Bob", "email": "bob@example.com", "role": "admin"},
    3: {"name": "Charlie", "email": "charlie@example.com", "role": "user"},
}

comments = []

@app.route('/')
def index():
    """Main page with search functionality (vulnerable to XSS)"""
    query = request.args.get('q', '')
    return render_template_string('''
<!DOCTYPE html>
<html>
<head>
    <title>Web Security Lab</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 15px; margin-bottom: 20px; border-radius: 5px; }
        .search-box { margin: 20px 0; }
        input[type="text"] { padding: 10px; width: 300px; }
        button { padding: 10px 20px; background: #007bff; color: white; border: none; cursor: pointer; }
        button:hover { background: #0056b3; }
        .section { margin: 30px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        code { background: #f4f4f4; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="warning">
        <strong>⚠️ Educational Environment Only</strong><br>
        This application contains intentional security vulnerabilities for learning purposes.
        Never deploy this in production.
    </div>

    <h1>Web Security Lab</h1>
    <p>Welcome to the Introduction to Web Security lab. This application demonstrates common vulnerabilities.</p>

    <div class="section">
        <h2>Search (XSS Vulnerability)</h2>
        <p>Try searching for: <code>&lt;script&gt;alert('XSS')&lt;/script&gt;</code></p>
        <div class="search-box">
            <form method="GET">
                <input type="text" name="q" placeholder="Search..." value="{{ query }}">
                <button type="submit">Search</button>
            </form>
        </div>
        {% if query %}
        <p>You searched for: {{ query }}</p>
        {% endif %}
    </div>

    <div class="section">
        <h2>User Profiles (IDOR Vulnerability)</h2>
        <p>View your profile: <a href="/profile?id=1">/profile?id=1</a></p>
        <p>Try accessing other users' profiles by changing the ID!</p>
    </div>

    <div class="section">
        <h2>Comments</h2>
        <form method="POST" action="/comment">
            <input type="text" name="comment" placeholder="Leave a comment..." style="width: 400px;">
            <button type="submit">Post</button>
        </form>
        <ul>
            {% for comment in comments %}
            <li>{{ comment }}</li>
            {% endfor %}
        </ul>
    </div>

    <div class="section">
        <h2>Learning Objectives</h2>
        <ul>
            <li>Identify the reflected XSS vulnerability in the search feature</li>
            <li>Exploit the IDOR vulnerability to access other users' data</li>
            <li>Understand why these vulnerabilities are dangerous</li>
            <li>Learn how to prevent these issues</li>
        </ul>
    </div>
</body>
</html>
    ''', query=query)

@app.route('/profile')
def profile():
    """User profile page (vulnerable to IDOR)"""
    user_id = request.args.get('id', '1')
    try:
        user_id = int(user_id)
    except ValueError:
        return "Invalid user ID", 400

    user = users.get(user_id)
    if not user:
        return "User not found", 404

    return render_template_string('''
<!DOCTYPE html>
<html>
<head>
    <title>User Profile</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .profile { background: #f8f9fa; padding: 20px; border-radius: 5px; }
        .label { font-weight: bold; }
    </style>
</head>
<body>
    <h1>User Profile</h1>
    <div class="profile">
        <p><span class="label">ID:</span> {{ user_id }}</p>
        <p><span class="label">Name:</span> {{ user.name }}</p>
        <p><span class="label">Email:</span> {{ user.email }}</p>
        <p><span class="label">Role:</span> {{ user.role }}</p>
    </div>
    <p><a href="/">← Back to home</a></p>
</body>
</html>
    ''', user_id=user_id, user=user)

@app.route('/comment', methods=['POST'])
def comment():
    """Comment submission (also vulnerable to stored XSS)"""
    comment = request.form.get('comment', '')
    if comment:
        comments.append(comment)
    return 'Comment posted! <a href="/">← Back</a>'

@app.route('/health')
def health():
    """Health check endpoint"""
    return jsonify({"status": "healthy"}), 200

if __name__ == '__main__':
    # Bind to all interfaces (Docker will map to 127.0.0.1 on host)
    app.run(host='0.0.0.0', port=8080, debug=True)
