# Introduction to Web Security

A simple web application demonstrating common security vulnerabilities including Cross-Site Scripting (XSS) and Insecure Direct Object References (IDOR).

## ⚠️ Security Warning

This environment contains intentionally vulnerable applications designed for:
- Security education and training
- Authorized security research
- Vulnerability remediation practice

**Never deploy this environment in production or expose it to untrusted networks.**

## Prerequisites

- Docker Engine 20.10+
- Docker Compose v2
- At least 256 MiB RAM available
- Basic understanding of HTML and HTTP
- Familiarity with web browsers and developer tools

## Quick Start

```bash
# From the project root
labctl run web/intro-lab

# Or manually with Docker Compose
cd templates/web/intro-lab
docker compose up -d
```

The application will be available at: http://127.0.0.1:8081

## Learning Objectives

By completing this lab, you will learn:

1. **Reflected XSS**: Identify and understand reflected Cross-Site Scripting vulnerabilities
2. **IDOR**: Learn about Insecure Direct Object References and how to exploit them
3. **Security Testing**: Practice identifying security issues in web applications
4. **Remediation**: Review input validation and output encoding techniques

## Vulnerabilities Demonstrated

### 1. Reflected XSS in Search

The search feature reflects user input without proper output encoding.

**Location**: Search box on the home page

**Test**: Search for `<script>alert('XSS')</script>`

**Impact**: Arbitrary JavaScript execution in the victim's browser

### 2. Insecure Direct Object Reference (IDOR)

The profile endpoint allows accessing any user's data by manipulating the ID parameter.

**Location**: `/profile?id=<user_id>`

**Test**: Access `/profile?id=2` to view Bob's profile (should be inaccessible)

**Impact**: Unauthorized access to other users' sensitive data

### 3. Stored XSS in Comments

Comments are stored and displayed without sanitization.

**Location**: Comment submission form

**Test**: Submit `<script>alert('XSS')</script>` as a comment

**Impact**: Persistent XSS affecting all users who view comments

## Lab Structure

```
templates/web/intro-lab/
├── lab.yaml              # Lab manifest
├── compose.yaml          # Docker Compose configuration
├── app.py                # Vulnerable Flask application
├── README.md            # This file
├── lessons/              # Learning materials
│   ├── objectives.md    # Detailed objectives
│   ├── concepts.md      # Background concepts
│   └── remediation.md   # How to fix the vulnerabilities
├── tests/                # Validation tests
│   ├── smoke.sh         # Smoke test script
│   └── expected-results.json
└── LICENSE
```

## Usage

### Starting the Lab

```bash
labctl run web/intro-lab
```

### Viewing Logs

```bash
labctl logs web/intro-lab
```

### Stopping the Lab

```bash
labctl stop web/intro-lab
```

### Resetting the Lab

```bash
labctl reset web/intro-lab
```

### Removing the Lab

```bash
labctl remove web/intro-lab
```

## Manual Docker Compose Usage

If you prefer to use Docker Compose directly:

```bash
# Start services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down

# Stop and remove volumes
docker compose down -v
```

## Verification

Run the smoke test to verify the lab is working correctly:

```bash
./tests/smoke.sh
```

## Exercise Steps

### Step 1: Explore the Application

1. Open http://127.0.0.1:8081 in your browser
2. Try the search feature with normal queries
3. View your profile at `/profile?id=1`
4. Post a comment

### Step 2: Identify XSS Vulnerability

1. In the search box, enter: `<script>alert('XSS')</script>`
2. Observe the alert box appearing
3. Check the page source to see how the input is reflected
4. Try different XSS payloads

### Step 3: Exploit IDOR Vulnerability

1. Start at your profile: `/profile?id=1`
2. Try accessing `/profile?id=2` (Bob's profile)
3. Try `/profile?id=3` (Charlie's profile)
4. Notice that you can view all users' data

### Step 4: Test Stored XSS

1. Submit a comment with: `<img src=x onerror=alert('XSS')>`
2. Refresh the page
3. Observe the alert appearing automatically

### Step 5: Review Remediation

See `lessons/remediation.md` for how to fix these vulnerabilities.

## Expected Results

After completing this lab, you should be able to:

- [ ] Identify the reflected XSS vulnerability in the search feature
- [ ] Successfully exploit the XSS vulnerability
- [ ] Identify the IDOR vulnerability in the profile endpoint
- [ ] Access other users' data via IDOR
- [ ] Understand the security implications of both vulnerabilities
- [ ] Implement proper input validation and output encoding
- [ ] Implement proper access control for IDOR prevention

## Cleanup

To completely remove the lab and all its data:

```bash
labctl remove web/intro-lab --volumes
```

Or with Docker Compose:

```bash
docker compose down -v
```

## Troubleshooting

### Port Already in Use

If port 8081 is already in use, modify the port mapping in `compose.yaml`:

```yaml
ports:
  - "127.0.0.1:8082:8080"
```

### Container Won't Start

Check the logs:

```bash
docker compose logs app
```

### Health Check Failing

Ensure the application is responding:

```bash
curl http://127.0.0.1:8081/health
```

## References

- [OWASP XSS Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP IDOR Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html)

## License

This lab template is licensed under Apache-2.0.

## Contributing

To improve this lab or report issues, please refer to the main project's CONTRIBUTING.md.
