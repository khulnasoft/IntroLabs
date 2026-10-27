# Background Concepts

## Overview

This section provides the background knowledge needed to understand and complete this lab on XSS and IDOR vulnerabilities.

## Key Concepts

### Cross-Site Scripting (XSS)

**Definition:** XSS is a security vulnerability that allows attackers to inject malicious scripts into web pages viewed by other users.

**How it works:**
1. An application accepts user input without proper validation
2. The input is reflected back to the user or stored in the application
3. When the page is rendered, the malicious script executes in the victim's browser
4. The script can steal cookies, session tokens, or perform actions on behalf of the user

**Types of XSS:**

1. **Reflected XSS**: The malicious script is reflected off a web application, typically in error messages or search results. The payload is delivered via a crafted URL.

2. **Stored XSS**: The malicious script is permanently stored on the target server (e.g., in a database). It affects all users who view the contaminated page.

3. **DOM-based XSS**: The vulnerability exists in client-side code rather than server-side code. The attack payload is executed as a result of modifying the DOM environment.

**Example:**
```html
<!-- Vulnerable code -->
<input type="text" name="search" value="{{ user_input }}">
```

If `user_input` is `"><script>alert('XSS')</script>`, the resulting HTML is:
```html
<input type="text" name="search" value=""><script>alert('XSS')</script>">
```

### Insecure Direct Object Reference (IDOR)

**Definition:** IDOR is a security vulnerability where an application provides direct access to objects based on user-supplied input without proper authorization checks.

**How it works:**
1. An application uses sequential or predictable identifiers (e.g., `/profile?id=1`, `/profile?id=2`)
2. The application checks if the user is authenticated but not if they're authorized to access the specific object
3. An attacker can manipulate the identifier to access other users' data
4. This leads to unauthorized data access or modification

**Example:**
```python
# Vulnerable code
@app.route('/profile')
def profile():
    user_id = request.args.get('id')
    user = get_user(user_id)  # No authorization check!
    return render_template('profile.html', user=user)
```

A user who should only access their own profile can access `/profile?id=2` to view another user's data.

**Common scenarios:**
- Viewing other users' profiles
- Accessing other users' orders
- Downloading other users' files
- Modifying other users' settings

## Technology Stack

This lab uses the following technologies:

- **Python 3.11**: Programming language for the web application
- **Flask**: Lightweight web framework for Python
- **Docker**: Containerization platform for running the lab
- **Docker Compose**: Tool for defining and running multi-container applications

## Vulnerability Classes

### XSS Vulnerability Class

**What makes it vulnerable:**
- Lack of output encoding
- Trusting user input without validation
- Using user input directly in HTML context

**Attack scenarios:**
- Session hijacking through cookie theft
- Credential harvesting
- Defacement of websites
- Phishing attacks
- Malware distribution

**Impact:**
- Unauthorized access to user accounts
- Data theft
- Privacy violations
- Reputation damage

### IDOR Vulnerability Class

**What makes it vulnerable:**
- Sequential or predictable identifiers
- Missing authorization checks
- Direct object access without validation
- Over-reliance on security through obscurity

**Attack scenarios:**
- Accessing other users' personal information
- Viewing other users' financial data
- Modifying other users' orders
- Deleting other users' data

**Impact:**
- Privacy violations
- Data breaches
- Financial loss
- Legal compliance violations

## Security Best Practices

### For XSS Prevention

1. **Output Encoding**: Always encode user-supplied data before rendering in HTML
   - Use context-specific encoding (HTML, JavaScript, URL, CSS)
   - Libraries: OWASP Java Encoder, DOMPurify, etc.

2. **Input Validation**: Validate and sanitize all user input
   - Whitelist allowed characters
   - Reject obviously malicious input
   - Validate on both client and server side

3. **Content Security Policy (CSP)**: Implement CSP headers
   - Restrict script sources
   - Prevent inline scripts
   - Use nonces or hashes for scripts

4. **HTTP Headers**: Set security headers
   - `X-XSS-Protection: 1; mode=block`
   - `X-Content-Type-Options: nosniff`

### For IDOR Prevention

1. **Authorization Checks**: Always verify user permissions
   - Check if user has access to the specific object
   - Implement proper access control lists
   - Use role-based access control (RBAC)

2. **Indirect Object References**: Use indirect references
   - Map sequential IDs to random tokens
   - Use UUIDs instead of sequential IDs
   - Validate the mapping server-side

3. **Session-Based Authorization**: Bind objects to sessions
   - Verify object ownership before access
   - Use session tokens for validation
   - Implement proper session management

4. **Logging and Monitoring**: Log access attempts
   - Monitor for suspicious access patterns
   - Alert on unauthorized access attempts
   - Audit object access regularly

## Further Reading

- [OWASP XSS Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [OWASP IDOR Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html)
- [OWASP Top 10 Web Application Security Risks](https://owasp.org/www-project-top-ten/)
- [CSP Quick Reference](https://content-security-policy.com/)
