# Menshen

Ephemeral git hosting service. Repositories are public, anonymous, and deleted after 48 hours of inactivity.

## Demo

https://terminus.re/

## Usage

### 1. Create a Repository
```bash
curl <your-service-url>/new
```

### 2. Use with Git
```bash
git clone http://<your-service-url>/<hash> repo
cd repo
# ... make changes ...
git push origin master
```

### 3. ToDo

- Add configuration options
    - Repo TTL
    - Port
- Add metrics
- Add proper logging
- Add route /ttl to get the TTL of a repository
- Add route /ttl/<hash> to get the TTL of a repository
- Add rate limiting
- Add route /delete/<hash> to delete a repository