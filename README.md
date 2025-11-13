# newsletter-service

## Deployment

### Prerequisites
- Install Heroku CLI: https://devcenter.heroku.com/articles/heroku-cli
- Have a Heroku account

### Steps

1. Create a Heroku app:
heroku create newsletter-

2. Add Heroku Postgres (optional if used):
heroku addons:create heroku-postgresql:hobby-dev

3. Set environment variables (config vars):
heroku config:set UPSTASH_REDIS_URL=your_redis_url
heroku config:set SMTP_HOST=smtp.example.com
heroku config:set SMTP_USER=username
heroku config:set SMTP_PASS=yourpassword

4. Deploy the app:
git push heroku main

5. Scale worker dynos:
heroku ps:scale worker=1

6. View logs:
heroku logs --tail

This setup should be sufficient to deploy and run your backend and worker services with Redis and SMTP integrations on Heroku.
