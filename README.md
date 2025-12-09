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

### 3. Check TTL
```bash
# Get global TTL
curl <your-service-url>/ttl

# Get remaining TTL for a repository
curl <your-service-url>/ttl/<hash>
```

### 4. Delete a Repository
```bash
curl -X DELETE <your-service-url>/delete/<hash>
```

## Configuration

Create a `config.yaml` file:

```yaml
port: 8080
ttl: 48h
```

## ToDo

